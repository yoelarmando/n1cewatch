# Install — All Ubuntu Versions

## One-liner (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/yoelarmando/n1cewatch/main/deploy/install.sh | sudo bash
```

## Manual Build — Any Ubuntu

### Ubuntu 22.04 / 24.04 (eBPF)
```bash
sudo apt update && sudo apt install -y clang llvm libbpf-dev linux-headers-$(uname -r) golang-go sqlite3
git clone https://github.com/yoelarmando/n1cewatch && cd n1cewatch
make build
sudo ./bin/n1cewatch --json --store-path /var/log/n1cewatch.db --store-query-port 8081
```

### Ubuntu 16.04 / 18.04 (auditd fallback, no BTF)
```bash
sudo apt update && sudo apt install -y auditd sqlite3
git clone https://github.com/yoelarmando/n1cewatch && cd n1cewatch
make build-static  # or ./bin/n1cewatch --no-ebpf
sudo ./bin/n1cewatch --no-ebpf --audit-log /var/log/audit/audit.log --json
```

## Frontend Dashboard

```bash
cd frontend && npm install && npm run dev
# -> http://localhost:3000  (proxies to :8081)
```

## Verify Tamper-Resistance (Option A)

```bash
# non-sudo cannot kill
su nobody -c "kill $(cat /run/n1cewatch/n1cewatch.pid)"
# -> Operation not permitted (PASS)

# sudo can manage
sudo systemctl status n1cewatch
sudo bpftool prog show | grep n1cewatch
curl http://localhost:8081/api/alerts | jq
```

## Docker Matrix

```bash
make docker  # builds ubuntu:16.04 + 22.04 images
docker run --rm -v /var/log/audit:/var/log/audit n1cewatch:22.04 --no-ebpf
```
