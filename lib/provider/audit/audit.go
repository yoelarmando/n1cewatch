package audit

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

// Provider tails /var/log/audit/audit.log — for Ubuntu 16.04/18.04 4.15 without BTF
type Provider struct {
	path string
	done chan struct{}
}

func New(path string) *Provider {
	if path == "" {
		path = "/var/log/audit/audit.log"
	}
	return &Provider{path: path, done: make(chan struct{})}
}

func (p *Provider) Name() string { return "auditd" }

func (p *Provider) Start() (<-chan event.Event, error) {
	ch := make(chan event.Event, 512)
	go func() {
		defer close(ch)
		f, err := os.Open(p.path)
		if err != nil {
			// No audit.log — emit nothing, allow /proc fallback
			<-p.done
			return
		}
		defer f.Close()
		// Seek end for real-time tail
		f.Seek(0, 2)
		scanner := bufio.NewScanner(f)
		// Poll for new data
		for {
			select {
			case <-p.done:
				return
			default:
			}
			for scanner.Scan() {
				line := scanner.Text()
				if ev := parseAuditLine(line); ev != nil {
					ch <- *ev
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	return ch, nil
}

func (p *Provider) Stop() error { close(p.done); return nil }

func parseAuditLine(line string) *event.Event {
	// Minimal: look for syscall key, exe, pid
	if !strings.Contains(line, "type=SYSCALL") && !strings.Contains(line, "type=EXECVE") {
		return nil
	}
	hostname, _ := os.Hostname()
	ev := &event.Event{
		Timestamp: time.Now().UTC(),
		Host:      hostname,
		Type:      event.TypeProcessCreation,
		EventID:   1,
		RawAudit:  map[string]string{"raw": line},
	}
	// Extract exe="..."  pid=...
	if idx := strings.Index(line, `exe="`); idx != -1 {
		end := strings.Index(line[idx+5:], `"`)
		if end != -1 {
			ev.Image = line[idx+5 : idx+5+end]
		}
	}
	return ev
}
