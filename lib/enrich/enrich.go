package enrich

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/n1cewatch/n1cewatch/lib/event"
)

// LRU parent cache — distributor style from Aurora
type Cache struct {
	mu   sync.Mutex
	data map[int]event.Event
	order []int
	max   int
}

func NewCache(max int) *Cache {
	return &Cache{data: make(map[int]event.Event), max: max}
}

func (c *Cache) Put(pid int, ev event.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.data[pid]; !ok {
		c.order = append(c.order, pid)
		if len(c.order) > c.max {
			old := c.order[0]
			c.order = c.order[1:]
			delete(c.data, old)
		}
	}
	c.data[pid] = ev
}

func (c *Cache) Get(pid int) (event.Event, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ev, ok := c.data[pid]
	return ev, ok
}

// Enrich fills User, CurrentDirectory, ParentImage, cgroup/container
func Enrich(ev *event.Event, cache *Cache) {
	// UID -> username
	if ev.User != "" {
		if uid, err := strconv.Atoi(ev.User); err == nil {
			if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
				ev.User = u.Username
			}
		}
	}
	// /proc enrich
	if ev.ProcessID != 0 {
		if cwd, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(ev.ProcessID), "cwd")); err == nil {
			ev.CurrentDirectory = cwd
		}
		if data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(ev.ProcessID), "cgroup")); err == nil {
			ev.Cgroup = strings.TrimSpace(string(data))
			// container detection: docker/k8s
			if strings.Contains(ev.Cgroup, "docker") || strings.Contains(ev.Cgroup, "kubepods") {
				ev.Container = "container"
			}
		}
		if ppidData, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(ev.ProcessID), "stat")); err == nil {
			fields := strings.Fields(string(ppidData))
			if len(fields) > 3 {
				if ppid, err := strconv.Atoi(fields[3]); err == nil {
					ev.ParentProcessID = ppid
					if parent, ok := cache.Get(ppid); ok {
						ev.ParentImage = parent.Image
						ev.ParentCommandLine = parent.CommandLine
					}
				}
			}
		}
		cache.Put(ev.ProcessID, *ev)
	}
}
