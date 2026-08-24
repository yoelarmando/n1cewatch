package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/n1cewatch/n1cewatch/lib/consumer/ioc"
	"github.com/n1cewatch/n1cewatch/lib/consumer/sigma"
	"github.com/n1cewatch/n1cewatch/lib/distributor"
	"github.com/n1cewatch/n1cewatch/lib/emitter"
	"github.com/n1cewatch/n1cewatch/lib/event"
	"github.com/n1cewatch/n1cewatch/lib/provider/audit"
	"github.com/n1cewatch/n1cewatch/lib/provider/proc"
	"github.com/n1cewatch/n1cewatch/lib/store"
	"github.com/n1cewatch/n1cewatch/lib/watch"
)

var (
	jsonOut        = flag.Bool("json", false, "JSON output")
	storePath      = flag.String("store-path", "/var/log/n1cewatch.db", "SQLite path")
	storeQueryPort = flag.Int("store-query-port", 8081, "HTTP query API port")
	rulesPath      = flag.String("rules", "packs/anomaly", "Sigma rules dir")
	auditLog       = flag.String("audit-log", "/var/log/audit/audit.log", "auditd log")
	udpTarget      = flag.String("udp-target", "", "UDP SIEM target host:port")
	noEBPF         = flag.Bool("no-ebpf", false, "force auditd/proc fallback")
	throttleRate   = flag.Float64("throttle-rate", 1.0, "max sigma per rule per sec")
	throttleBurst  = flag.Int("throttle-burst", 5, "burst")
)

func main() {
	flag.Parse()
	fmt.Printf("N1ceWatch Blue v0.1.0 — Ubuntu all-versions (BTF probe: %v, noEBPF=%v)\n", probeBTF(), *noEBPF)

	// Select provider — all-Ubuntu orchestration
	var providerCh <-chan event.Event
	var providerName string
	stopFuncs := []func() error{}

	if !*noEBPF && probeBTF() {
		// Try eBPF — handle import on Windows dev gracefully
		if err := tryEBPF(&providerCh, &providerName, &stopFuncs); err != nil {
			log.Printf("[WARN] eBPF failed (%v), falling back to auditd", err)
			providerCh, providerName = tryAudit(&stopFuncs)
		}
	} else {
		providerCh, providerName = tryAudit(&stopFuncs)
	}
	if providerCh == nil {
		providerCh, providerName = tryProc(&stopFuncs)
	}
	log.Printf("[INFO] Provider: %s (rules=%s store=%s:%d)", providerName, *rulesPath, *storePath, *storeQueryPort)

	// Distributor
	dist := distributor.New(*throttleRate, *throttleBurst)
	sigmaCh := make(chan event.Event, 512)
	iocCh := make(chan event.Event, 128)
	fsCh := make(chan event.Event, 64)

	// fsnotify watcher (GhostCatcher pattern) — dedicated channel to avoid distributor deadlock
	fsNotifyCh := make(chan event.Event, 64)
	if w, err := watch.New(); err == nil {
		fsEvents := w.Start()
		go func() {
			for ev := range fsEvents {
				log.Printf("[DEBUG] fsnotify event: %s %s", ev.TargetFilename, ev.FileAction)
				select {
				case fsNotifyCh <- ev:
				default:
					log.Printf("[WARN] fsNotifyCh full, dropping %s", ev.TargetFilename)
				}
			}
		}()
		defer w.Close()
	} else {
		close(fsNotifyCh)
		log.Printf("[WARN] fsnotify init failed: %v", err)
	}

	dist.Start(providerCh, sigmaCh, iocCh, fsCh)

	// Consumers
	sigmaEngine, _ := sigma.New(*rulesPath, *throttleRate, *throttleBurst)
	iocEngine := ioc.New("resources/iocs/c2-iocs.txt", "resources/iocs/filename-iocs.txt")

	sigmaAlerts := sigmaEngine.Start(sigmaCh)
	iocAlerts := iocEngine.Start(iocCh)

	// Store + Emitter
	st, err := store.New(*storePath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()
	go st.ServeHTTP(*storeQueryPort)

	em, _ := emitter.New("/var/log/n1cewatch/events.jsonl", *udpTarget)
	defer em.Close()

	// Fan-in alerts — include fsNotifyCh
	alertCh := make(chan event.Alert, 512)
	go func() {
		defer close(alertCh)
		for {
			select {
			case a, ok := <-sigmaAlerts:
				if !ok {
					sigmaAlerts = nil
				} else {
					log.Printf("[DEBUG] sigma alert: %s %s", a.RuleName, a.Event.CommandLine)
					alertCh <- a
				}
			case a, ok := <-iocAlerts:
				if !ok {
					iocAlerts = nil
				} else {
					alertCh <- a
				}
			case ev, ok := <-fsCh:
				if !ok {
					fsCh = nil
				} else {
					// proc/audit file events treated as high if cron
					if ev.TargetFilename != "" {
						log.Printf("[DEBUG] fsCh file event: %s", ev.TargetFilename)
					}
				}
			case ev, ok := <-fsNotifyCh:
				if !ok {
					fsNotifyCh = nil
				} else {
					log.Printf("[ALERT] fsnotify persistence: %s %s", ev.TargetFilename, ev.FileAction)
					alertCh <- event.Alert{
						Host:     ev.Host,
						RuleID:   "fsnotify-persistence",
						RuleName: "Persistence file modified",
						Level:    "medium",
						MitreID:  "T1547",
						Event:    ev,
					}
				}
			}
			if sigmaAlerts == nil && iocAlerts == nil && fsCh == nil && fsNotifyCh == nil {
				return
			}
		}
	}()

	// Handle signals + emit
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for a := range alertCh {
			if *jsonOut {
				em.Emit(a)
			} else {
				fmt.Printf("[%s] %s %s %s %s -> %s\n", a.Level, a.RuleName, a.Event.Image, a.Event.CommandLine, a.MitreID, a.Host)
				em.Emit(a)
			}
			st.Save(a)
		}
	}()

	<-sigCh
	log.Println("[INFO] shutting down...")
	for _, fn := range stopFuncs {
		fn()
	}
}

// try helpers — isolated to allow Windows build (no cgo ebpf)

func tryAudit(stopFuncs *[]func() error) (<-chan event.Event, string) {
	p := audit.New(*auditLog)
	ch, _ := p.Start()
	*stopFuncs = append(*stopFuncs, p.Stop)
	return ch, "auditd"
}

func tryProc(stopFuncs *[]func() error) (<-chan event.Event, string) {
	p := proc.New(500 * 1000 * 1000) // 500ms for short-lived curl|bash capture
	ch, _ := p.Start()
	*stopFuncs = append(*stopFuncs, p.Stop)
	return ch, "proc"
}
