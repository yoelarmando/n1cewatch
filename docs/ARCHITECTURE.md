# N1ceWatch Architecture — All-Ubuntu EDR-Lite

## Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        Linux Kernel 4.15 → 6.8                         │
│  tracepoint/sched/sched_process_exec  openat  inet_sock_set_state      │
│  utimensat  bpf syscall        ─┐                                       │
└─────────────────────────────────┼──────────────────────────────────────┘
                                  │ CO-RE ringbuf (cilium/ebpf)
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Provider Layer (lib/provider)                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  BTF Probe at startup       │
│  │  eBPF    │→ │  auditd  │→ │  /proc   │  probeBTF() checks           │
│  │  CO-RE   │  │  tail    │  │  poll    │  /sys/kernel/btf/vmlinux    │
│  └──────────┘  └──────────┘  └──────────┘  fallback chain            │
│  Enrich: /proc/PID/{cmdline,cwd,exe} + parent LRU cache              │
└──────────────────────────────────────────────────────────────────────┘
                                  │ Event{Type, PID, Image, Cmdline, ParentImage...}
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Distributor (lib/distributor)                                         │
│  - UID → username (os/user)                                          │
│  - Parent correlator (LRU 16384)                                     │
│  - Cgroup → container/pod enrich                                     │
│  - Throttle (1.0/sec + burst 5)                                      │
└──────────────────────────────────────────────────────────────────────┘
                                  │ Normalized Event
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Consumers                                                             │
│  ┌──────────────┐ ┌──────────┐ ┌──────────────────┐                  │
│  │ Sigma Engine │ │ IOC      │ │ fsnotify Watcher │                  │
│  │ 119 proc     │ │ c2-iocs  │ │ cron, ld.so.pre  │                  │
│  │ 8 file       │ │ filename │ │ systemd, sudoers │                  │
│  │ throttled    │ │          │ │ debounced rescan │                  │
│  └──────────────┘ └──────────┘ └──────────────────┘                  │
└──────────────────────────────────────────────────────────────────────┘
                                  │ Alert{Rule, Level, Mitre, Event}
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Emitter + Store (lib/emitter, lib/store)                              │
│  JSONL: stdout + /var/log/n1cewatch/events.jsonl + UDP/HTTP sinks     │
│  SQLite: /var/log/n1cewatch.db + HTTP :8081 (/api/alerts, /report.html)│
│  Spool: /var/log/n1cewatch/spool (retry on sink failure)              │
└──────────────────────────────────────────────────────────────────────┘
                                  │ SSE / HTTP Poll
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│ Frontend (Next.js 14, TypeScript)                                     │
│  Pages: Alerts (real-time), Process Tree, Hosts, Rules, Reports      │
│  API Gateway: Go :8080 -> SQLite + NATS-lite SSE                     │
└──────────────────────────────────────────────────────────────────────┘
```

## Provider Selection (All Ubuntu)

```go
// cmd/agent/btf_probe.go
func probeBTF() bool {
  // Check /sys/kernel/btf/vmlinux exists + bpftool feature probe
  if _, err := os.Stat("/sys/kernel/btf/vmlinux"); err == nil {
    return true // 5.8+ Ubuntu 20.04+
  }
  // Also try /sys/kernel/debug/tracing/available_filter_functions
  return false // 16.04/18.04 4.15 -> fallback
}

func selectProvider(cfg Config) Provider {
  if !cfg.NoEBPF && probeBTF() {
    if p, err := ebpf.New(); err == nil { return p }
  }
  if _, err := os.Stat(cfg.AuditLog); err == nil {
    return audit.New(cfg.AuditLog)
  }
  return proc.New(2 * time.Second) // poll /proc
}
```

## Hardening (Option A)

* Non-sudo cannot `kill` -> `CAP_BPF` requires root, `kill(2)` checks `CAP_KILL` or `same UID`.
* `systemd` `ProtectSystem=strict` mounts `/usr` ro, `ReadWritePaths` only `/var/log/n1cewatch`.
* `Restart=always` resurrects if non-sudo kills child (but `sudo systemctl stop` still works - visible to admin).

## Skills & Agents Orchestration

See `.opencode/` — `golang-pro` for eBPF Go, `nextjs-developer` for dashboard, `test-master` for e2e, `code-reviewer` for Sigma.

