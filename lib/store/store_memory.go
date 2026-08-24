//go:build without_ebpf

package store

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

// MemoryStore for static builds without CGO/sqlite — used on Kali when headers missing
type Store struct {
	mu     sync.Mutex
	alerts []event.Alert
	path   string
}

func New(path string) (*Store, error) {
	return &Store{path: path}, nil
}

func (s *Store) Save(a event.Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, a)
	if len(s.alerts) > 10000 {
		s.alerts = s.alerts[1:]
	}
	return nil
}

func (s *Store) ServeHTTP(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		var out []map[string]interface{}
		// last 100 reverse
		for i := len(s.alerts) - 1; i >= 0 && len(out) < 100; i-- {
			a := s.alerts[i]
			out = append(out, map[string]interface{}{
				"ts": a.Timestamp.Format(time.RFC3339), "host": a.Host, "rule_id": a.RuleID, "rule_name": a.RuleName, "level": a.Level, "mitre": a.MitreID, "event": a.Event,
			})
		}
		if out == nil {
			out = []map[string]interface{}{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		c := len(s.alerts)
		s.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]int{"total_alerts": c})
	})
	mux.HandleFunc("/report.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		s.mu.Lock()
		defer s.mu.Unlock()
		fmt.Fprint(w, `<html><head><title>N1ceWatch Report</title><style>body{font-family:monospace;background:#0a0a0a;color:#00ff00;padding:20px} table{border-collapse:collapse;width:100%} th,td{border:1px solid #333;padding:8px;text-align:left} th{background:#111}</style></head><body><h1>N1ceWatch Blue - Compliance Report (Memory)</h1>`)
		fmt.Fprintf(w, `<p>Generated: %s Total: %d (fallback mode without sqlite)</p><table><tr><th>TS</th><th>Host</th><th>Rule</th><th>Level</th><th>Mitre</th></tr>`, time.Now().Format(time.RFC3339), len(s.alerts))
		for i := len(s.alerts) - 1; i >= 0 && i > len(s.alerts)-200; i-- {
			a := s.alerts[i]
			fmt.Fprintf(w, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, a.Timestamp.Format(time.RFC3339), a.Host, a.RuleName, a.Level, a.MitreID)
		}
		fmt.Fprint(w, `</table></body></html>`)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (s *Store) Close() error { return nil }
