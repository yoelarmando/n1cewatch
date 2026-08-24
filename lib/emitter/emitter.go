package emitter

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

// Emitter writes JSONL to stdout + file + UDP/Syslog sinks with spool
type Emitter struct {
	filePath   string
	udpTarget  string
	mu         sync.Mutex
	file       *os.File
	spoolDir   string
}

func New(filePath, udpTarget string) (*Emitter, error) {
	e := &Emitter{filePath: filePath, udpTarget: udpTarget, spoolDir: filepath.Join(filepath.Dir(filePath), "spool")}
	os.MkdirAll(filepath.Dir(filePath), 0750)
	os.MkdirAll(e.spoolDir, 0750)
	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		e.file = f
	}
	return e, nil
}

func (e *Emitter) Emit(a event.Alert) error {
	data, _ := json.Marshal(a)
	line := string(data) + "\n"

	// stdout
	fmt.Println(line[:len(line)-1])

	// file
	if e.file != nil {
		e.mu.Lock()
		e.file.WriteString(line)
		e.mu.Unlock()
	}

	// UDP sink
	if e.udpTarget != "" {
		if err := e.sendUDP(line); err != nil {
			// spool to disk like GhostCatcher
			spoolPath := filepath.Join(e.spoolDir, fmt.Sprintf("%d.jsonl", len(line)))
			os.WriteFile(spoolPath, []byte(line), 0640)
		}
	}
	return nil
}

func (e *Emitter) sendUDP(line string) error {
	conn, err := net.Dial("udp", e.udpTarget)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(line))
	return err
}

func (e *Emitter) Close() error {
	if e.file != nil {
		return e.file.Close()
	}
	return nil
}
