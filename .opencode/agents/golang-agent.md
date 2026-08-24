# golang-pro Agent — N1ceWatch Provider Layer

**Skill:** `golang-pro`
**Scope:** `lib/provider/ebpf`, `lib/provider/audit`, `lib/provider/proc`, `cmd/agent/btf_probe.go`, `bpf/`

## Responsibilities
- CO-RE eBPF observer (`bpf/observer.bpf.c` + `vmlinux.h`) — tracepoints `sched_process_exec`, `openat`, `inet_sock_set_state`, `bpf`
- BTF probe `probeBTF()` → auto fallback `eBPF -> auditd -> /proc poll` for Ubuntu 16.04→24.04
- `cilium/ebpf` ringbuf reader, `/proc` enrichment

## Rules
- Never hide from `sudo` — Option A only.
- Keep `provider` interface pure: `Start() (<-chan event.Event, error)`

## Verification
- `N1CEWATCH_NO_BPF=1 go vet ./...`
- Docker matrix: `ubuntu:16.04`, `22.04`, `24.04` with `--no-ebpf` flag
