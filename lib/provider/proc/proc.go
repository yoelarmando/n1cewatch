package proc

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

// Provider polls /proc — final fallback for Ubuntu 16.04 without auditd
// Inspired by GhostCatcherEDR /proc poll.
type Provider struct {
	interval time.Duration
	done     chan struct{}
	seen     map[int]bool
}

func New(interval time.Duration) *Provider {
	if interval == 0 {
		interval = 2 * time.Second
	}
	return &Provider{interval: interval, done: make(chan struct{}), seen: make(map[int]bool)}
}

func (p *Provider) Name() string { return "proc" }

func (p *Provider) Start() (<-chan event.Event, error) {
	ch := make(chan event.Event, 512)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		hostname, _ := os.Hostname()
		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
				entries, err := os.ReadDir("/proc")
				if err != nil {
					continue
				}
				for _, e := range entries {
					if !e.IsDir() {
						continue
					}
					pid, err := strconv.Atoi(e.Name())
					if err != nil {
						continue
					}
					if p.seen[pid] {
						continue
					}
					p.seen[pid] = true
					// Trim seen to avoid leak
					if len(p.seen) > 16384 {
						p.seen = make(map[int]bool)
					}
					ev := event.Event{
						Timestamp: time.Now().UTC(),
						Host:      hostname,
						Type:      event.TypeProcessCreation,
						EventID:   1,
						ProcessID: pid,
					}
					if link, err := os.Readlink(filepath.Join("/proc", e.Name(), "exe")); err == nil {
						ev.Image = link
					}
					if data, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline")); err == nil {
						ev.CommandLine = strings.ReplaceAll(string(data), "\x00", " ")
					}
					ch <- ev
				}
			}
		}
	}()
	return ch, nil
}

func (p *Provider) Stop() error { close(p.done); return nil }
