//go:build !linux

package main

import (
	"errors"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

func tryEBPF(ch *<-chan event.Event, name *string, stopFuncs *[]func() error) error {
	return errors.New("eBPF only on Linux")
}
