//go:build without_ebpf

package ebpf

import (
	"time"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

type Provider struct{ done chan struct{} }

func New() (*Provider, error) { return &Provider{done: make(chan struct{})}, nil }
func (p *Provider) Name() string { return "ebpf-stub" }
func (p *Provider) Start() (<-chan event.Event, error) {
	ch := make(chan event.Event, 1024)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
			}
		}
	}()
	return ch, nil
}
func (p *Provider) Stop() error { close(p.done); return nil }
