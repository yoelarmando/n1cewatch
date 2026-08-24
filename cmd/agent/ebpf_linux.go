//go:build linux

package main

import (
	"github.com/n1cewatch/n1cewatch/lib/event"
	"github.com/n1cewatch/n1cewatch/lib/provider/ebpf"
)

func tryEBPF(ch *<-chan event.Event, name *string, stopFuncs *[]func() error) error {
	p, err := ebpf.New()
	if err != nil {
		return err
	}
	c, err := p.Start()
	if err != nil {
		return err
	}
	*ch = c
	*name = "ebpf"
	*stopFuncs = append(*stopFuncs, p.Stop)
	return nil
}
