package ioc

import (
	"bufio"
	"os"
	"strings"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

type Engine struct {
	c2IPs    map[string]bool
	filenames map[string]bool
}

func New(c2Path, filePath string) *Engine {
	e := &Engine{c2IPs: make(map[string]bool), filenames: make(map[string]bool)}
	load := func(path string, m map[string]bool) {
		f, err := os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			m[line] = true
		}
	}
	load(c2Path, e.c2IPs)
	load(filePath, e.filenames)
	return e
}

func (e *Engine) Evaluate(ev event.Event) *event.Alert {
	if e.c2IPs[ev.DestinationIP] {
		return &event.Alert{
			Host: ev.Host, RuleID: "ioc-c2", RuleName: "C2 IP match", Level: "high", MitreID: "T1071", Event: ev,
		}
	}
	for fn := range e.filenames {
		if strings.Contains(ev.TargetFilename, fn) || strings.Contains(ev.Image, fn) {
			return &event.Alert{
				Host: ev.Host, RuleID: "ioc-file", RuleName: "IOC filename match", Level: "high", Event: ev,
			}
		}
	}
	return nil
}

func (e *Engine) Start(in <-chan event.Event) <-chan event.Alert {
	out := make(chan event.Alert, 128)
	go func() {
		defer close(out)
		for ev := range in {
			if a := e.Evaluate(ev); a != nil {
				out <- *a
			}
		}
	}()
	return out
}
