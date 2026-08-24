# Troubleshooting — Kali / Ubuntu Headers

## Error: Unable to locate package linux-headers-6.19.14+kali-amd64

Kali rolling often has no `linux-headers-$(uname -r)` package for custom kernel.

**Fix — use fallback (no eBPF headers needed, agent still works via auditd -> /proc):**

```bash
# Do NOT require exact headers — agent auto-falls back
sudo apt update && sudo apt install -y golang-go clang llvm libbpf-dev
# Try meta package (works on Kali/Debian/Ubuntu)
sudo apt install -y linux-headers-amd64 || sudo apt install -y linux-headers-generic || echo "headers skip — fallback works"

cd /home/kali/n1cewatch
make build-static  # CGO_ENABLED=0, no kernel headers needed
sudo ./bin/n1cewatch --no-ebpf --help  # verify

# Run without eBPF
sudo ./bin/n1cewatch --json --store-path /var/log/n1cewatch.db --store-query-port 8081 --no-ebpf &
curl http://localhost:8081/api/alerts | jq
```

**Why 8081 was failing:** `deploy/install.sh` exited before `systemd` start when no GitHub Release existed. Now patched to auto-build from source. Pull latest:

```bash
cd /home/kali/n1cewatch && git pull && sudo bash deploy/install.sh
```

## Verify BTF

```bash
ls -l /sys/kernel/btf/vmlinux  # if exists -> eBPF CO-RE works, else fallback is correct
sudo auditctl -l  # check auditd rules
cat /proc/version
```

## Still 8081 Failed to Connect?

```bash
sudo systemctl status n1cewatch --no-pager
sudo journalctl -u n1cewatch -f
sudo tail -f /var/log/n1cewatch/events.jsonl
sudo netstat -tulpn | grep 8081
```
