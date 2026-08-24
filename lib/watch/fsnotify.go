package watch

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/n1cewatch/n1cewatch/lib/event"
)

// Watcher monitors persistence paths via fsnotify — GhostCatcher pattern
type Watcher struct {
	watcher *fsnotify.Watcher
	paths   []string
	debounce time.Duration
}

func New() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	paths := []string{
		"/etc/crontab",
		"/etc/cron.d",
		"/etc/cron.hourly",
		"/etc/systemd/system",
		"/etc/sudoers.d",
		"/etc/pam.d",
		"/etc/ld.so.preload",
		"/root/.ssh/authorized_keys",
		"/home",
	}
	for _, p := range paths {
		_ = w.Add(p) // best effort; non-existent is ok
		// Also try glob for home subdirs
		if p == "/home" {
			matches, _ := filepath.Glob("/home/*/ .ssh/authorized_keys")
			_ = matches
		}
	}
	return &Watcher{watcher: w, paths: paths, debounce: 500 * time.Millisecond}, nil
}

func (w *Watcher) Start() <-chan event.Event {
	ch := make(chan event.Event, 64)
	go func() {
		defer close(ch)
		defer w.watcher.Close()
		var last time.Time
		hostname, _ := os.Hostname()
		for {
			select {
			case ev, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if time.Since(last) < w.debounce {
					continue
				}
				last = time.Now()
				ch <- event.Event{
					Timestamp: time.Now().UTC(),
					Host:      hostname,
					Type:      event.TypeFSNotify,
					EventID:   11,
					TargetFilename: ev.Name,
					FileAction: ev.Op.String(),
					Image: "fsnotify",
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				_ = err
			}
		}
	}()
	return ch
}

func (w *Watcher) Close() error { return w.watcher.Close() }
