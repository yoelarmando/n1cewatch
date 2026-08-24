package sigma

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/n1cewatch/n1cewatch/lib/event"
	"gopkg.in/yaml.v3"
)

// Rule mirrors Sigma YAML minimal
type Rule struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Level       string   `yaml:"level"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	MitreID     string   `yaml:"mitre"`
	Detection   struct {
		Selection map[string]interface{} `yaml:"selection"`
		Condition string                 `yaml:"condition"`
	} `yaml:"detection"`
}

// Engine evaluates events against YAML packs
type Engine struct {
	rules    []Rule
	mu       sync.Mutex
	hits     map[string]time.Time
	rate     float64
	burst    int
}

func New(rulesPath string, rate float64, burst int) (*Engine, error) {
	e := &Engine{hits: make(map[string]time.Time), rate: rate, burst: burst}
	if rulesPath == "" {
		return e, nil
	}
	err := filepath.Walk(rulesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !(strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var r Rule
		if err := yaml.Unmarshal(data, &r); err == nil {
			e.rules = append(e.rules, r)
		}
		return nil
	})
	return e, err
}

func (e *Engine) Evaluate(ev event.Event) []event.Alert {
	var alerts []event.Alert
	for _, r := range e.rules {
		if !e.matchRule(r, ev) {
			continue
		}
		// throttle per rule
		key := r.ID
		e.mu.Lock()
		last, ok := e.hits[key]
		if ok && time.Since(last) < time.Duration(1/e.rate*float64(time.Second)) {
			e.mu.Unlock()
			continue
		}
		e.hits[key] = time.Now()
		e.mu.Unlock()

		alerts = append(alerts, event.Alert{
			Timestamp: time.Now().UTC(),
			Host:      ev.Host,
			RuleID:    r.ID,
			RuleName:  r.Title,
			Level:     r.Level,
			MitreID:   r.MitreID,
			Tags:      r.Tags,
			Event:     ev,
		})
	}
	return alerts
}

func (e *Engine) matchRule(r Rule, ev event.Event) bool {
	// Simplified matcher: check selection keywords against Image/CommandLine
	// Real Aurora uses go-sigma-rule-engine with field mapping; this is portable stub
	for k, v := range r.Detection.Selection {
		needle := ""
		switch val := v.(type) {
		case string:
			needle = strings.ToLower(val)
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok && contains(ev, strings.ToLower(s)) {
					needle = "" // matched one of list
					break
				}
				needle = "notfound"
			}
			if needle == "" {
				continue
			}
			return false
		default:
			continue
		}
		_ = k
		if needle != "" && !contains(ev, needle) {
			return false
		}
	}
	return len(r.Detection.Selection) > 0
}

func contains(ev event.Event, needle string) bool {
	hay := strings.ToLower(ev.Image + " " + ev.CommandLine + " " + ev.ParentImage + " " + ev.TargetFilename)
	return strings.Contains(hay, needle)
}

func (e *Engine) Start(in <-chan event.Event) <-chan event.Alert {
	out := make(chan event.Alert, 512)
	go func() {
		defer close(out)
		for ev := range in {
			for _, a := range e.Evaluate(ev) {
				out <- a
			}
		}
	}()
	return out
}
