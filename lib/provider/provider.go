package provider

import "github.com/n1cewatch/n1cewatch/lib/event"

// Provider abstracts eBPF / auditd / proc polling
type Provider interface {
	Name() string
	Start() (<-chan event.Event, error)
	Stop() error
}
