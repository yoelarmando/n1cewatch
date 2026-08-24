//go:build !without_ebpf

package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/n1cewatch/n1cewatch/lib/event"
)

// Store persists alerts to SQLite + serves HTTP query API
type Store struct {
	db   *sql.DB
	path string
}

func New(path string) (*Store, error) {
	if path == "" {
		path = "/var/log/n1cewatch.db"
	}
	db, err := sql.Open("sqlite3", path+"?cache=shared&mode=rwc")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ts TEXT,
		host TEXT,
		rule_id TEXT,
		rule_name TEXT,
		level TEXT,
		mitre TEXT,
		event_json TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_ts ON alerts(ts);
	CREATE INDEX IF NOT EXISTS idx_level ON alerts(level);
	`)
	return err
}

func (s *Store) Save(a event.Alert) error {
	j, _ := json.Marshal(a.Event)
	_, err := s.db.Exec(`INSERT INTO alerts(ts,host,rule_id,rule_name,level,mitre,event_json) VALUES(?,?,?,?,?,?,?)`,
		a.Timestamp.Format(time.RFC3339), a.Host, a.RuleID, a.RuleName, a.Level, a.MitreID, string(j))
	return err
}

func (s *Store) ServeHTTP(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		rows, err := s.db.Query(`SELECT ts,host,rule_id,rule_name,level,mitre,event_json FROM alerts ORDER BY id DESC LIMIT 100`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var out []map[string]interface{}
		for rows.Next() {
			var ts, host, ruleID, ruleName, level, mitre, ej string
			rows.Scan(&ts, &host, &ruleID, &ruleName, &level, &mitre, &ej)
			var ev map[string]interface{}
			json.Unmarshal([]byte(ej), &ev)
			out = append(out, map[string]interface{}{
				"ts": ts, "host": host, "rule_id": ruleID, "rule_name": ruleName, "level": level, "mitre": mitre, "event": ev,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		var cnt int
		s.db.QueryRow(`SELECT COUNT(*) FROM alerts`).Scan(&cnt)
		json.NewEncoder(w).Encode(map[string]int{"total_alerts": cnt})
	})
	mux.HandleFunc("/report.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><head><title>N1ceWatch Report</title><style>body{font-family:monospace;background:#0a0a0a;color:#00ff00;padding:20px} table{border-collapse:collapse;width:100%} th,td{border:1px solid #333;padding:8px;text-align:left} th{background:#111}</style></head><body><h1>N1ceWatch Blue - Compliance Report</h1>`)
		fmt.Fprint(w, `<p>Generated: `+time.Now().Format(time.RFC3339)+`</p><table><tr><th>TS</th><th>Host</th><th>Rule</th><th>Level</th><th>Mitre</th></tr>`)
		rows, _ := s.db.Query(`SELECT ts,host,rule_name,level,mitre FROM alerts ORDER BY id DESC LIMIT 200`)
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var ts, host, name, lvl, mitre string
				rows.Scan(&ts, &host, &name, &lvl, &mitre)
				fmt.Fprintf(w, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, ts, host, name, lvl, mitre)
			}
		}
		fmt.Fprint(w, `</table></body></html>`)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (s *Store) Close() error { return s.db.Close() }
