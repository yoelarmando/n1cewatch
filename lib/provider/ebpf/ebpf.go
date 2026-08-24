//go:build linux

package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/n1cewatch/n1cewatch/lib/event"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -type event bpf ../../bpf/observer.bpf.c -- -I../../bpf

// Provider implements ringbuf -> Event
type Provider struct {
	coll *ebpf.Collection
	rd   *ringbuf.Reader
	done chan struct{}
}

func New() (*Provider, error) {
	if _, err := os.Stat("/sys/kernel/btf/vmlinux"); err != nil {
		return nil, fmt.Errorf("BTF not available, need auditd fallback: %w", err)
	}
	// In CI without actual BPF object, return mock-friendly
	if os.Getenv("N1CEWATCH_NO_BPF") == "1" {
		return &Provider{done: make(chan struct{})}, nil
	}
	// Real load would use bpf2go generated objs:
	// objs := bpfObjects{}
	// if err := loadBpfObjects(&objs, nil); err != nil { return nil, err }
	return &Provider{done: make(chan struct{})}, nil
}

func (p *Provider) Name() string { return "ebpf" }

func (p *Provider) Start() (<-chan event.Event, error) {
	ch := make(chan event.Event, 1024)
	if p.coll == nil && p.rd == nil {
		// Mock mode for Windows dev / all-ubuntu CI without kernel
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-p.done:
					close(ch)
					return
				case <-ticker.C:
					// emit heartbeat like Aurora provider
				}
			}
		}()
		return ch, nil
	}
	// Real ringbuf reader:
	// p.rd, _ = ringbuf.NewReader(p.coll.Maps["events"])
	go func() {
		defer close(ch)
		var raw struct {
			PID  uint32
			PPID uint32
			UID  uint32
			Type uint32
		}
		for {
			select {
			case <-p.done:
				return
			default:
			}
			if p.rd == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			rec, err := p.rd.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					return
				}
				continue
			}
			if err := binary.Read(bytes.NewBuffer(rec.RawSample), binary.LittleEndian, &raw); err != nil {
				continue
			}
			hostname, _ := os.Hostname()
			ev := event.Event{
				Timestamp:       time.Now().UTC(),
				Host:            hostname,
				Type:            event.TypeProcessCreation,
				EventID:         1,
				ProcessID:       int(raw.PID),
				ParentProcessID: int(raw.PID),
				User:            fmt.Sprintf("%d", raw.UID),
			}
			// Enrich from /proc
			enrichProc(&ev)
			ch <- ev
		}
	}()
	return ch, nil
}

func (p *Provider) Stop() error {
	close(p.done)
	if p.rd != nil {
		p.rd.Close()
	}
	if p.coll != nil {
		p.coll.Close()
	}
	return nil
}

func enrichProc(ev *event.Event) {
	// Minimal /proc enrichment, distributor will do full
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", ev.ProcessID)); err == nil {
		ev.CommandLine = string(bytes.ReplaceAll(data, []byte{0}, []byte{' '}))
	}
	if link, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", ev.ProcessID)); err == nil {
		ev.Image = link
	}
}
