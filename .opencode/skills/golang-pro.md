# golang-pro Skill — N1ceWatch

**For:** `lib/provider/**`, `lib/distributor/**`, `lib/consumer/**`, `cmd/agent/**`, `bpf/**`

Rules:
- Go 1.22+, `go vet`, `CGO_ENABLED=1` for eBPF, `without_ebpf` tag for fallback
- Use `cilium/ebpf` ringbuf, `fsnotify`, `mattn/go-sqlite3`
- All-Ubuntu: probe BTF at runtime, never panic on missing BTF
- Option A: never hide from sudo, only protect from non-sudo via systemd
