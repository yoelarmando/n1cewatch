package distributor

import (
	"sync"
	"time"

	"github.com/n1cewatch/n1cewatch/lib/enrich"
	"github.com/n1cewatch/n1cewatch/lib/event"
)

// Distributor routes provider events -> consumers with enrichment + throttling
type Distributor struct {
	cache *enrich.Cache
	mu    sync.Mutex
	throttle map[string]time.Time
	rate     float64
	burst    int
}

func New(rate float64, burst int) *Distributor {
	return &Distributor{
		cache: enrich.NewCache(16384),
		throttle: make(map[string]time.Time),
		rate: rate,
		burst: burst,
	}
}

// Process enriches and returns false if throttled
func (d *Distributor) Process(ev *event.Event) bool {
	enrich.Enrich(ev, d.cache)
	// Per-event throttle not applied here; per-rule throttle in sigma consumer
	return true
}

// Start fans out from provider channel to fanout channels
func (d *Distributor) Start(in <-chan event.Event, outs ...chan event.Event) {
	go func() {
		for ev := range in {
			if !d.Process(&ev) {
				continue
			}
			for _, out := range outs {
				select {
				case out <- ev:
				default:
					// drop if slow consumer (ringbuf full)
				}
			}
		}
		for _, out := range outs {
			close(out)
		}
	}()
}
