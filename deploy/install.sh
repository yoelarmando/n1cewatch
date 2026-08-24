#!/bin/bash
# N1ceWatch Blue — All-Ubuntu installer (16.04 -> 24.04)
# Usage: curl -fsSL https://raw.githubusercontent.com/n1cewatch/n1cewatch/main/deploy/install.sh | sudo bash
set -e

REPO="https://github.com/n1cewatch/n1cewatch"
INSTALL_DIR="/opt/n1cewatch"
BIN="$INSTALL_DIR/bin/n1cewatch"
LOG_DIR="/var/log/n1cewatch"
DB="/var/log/n1cewatch.db"

echo "[*] N1ceWatch installer — Ubuntu all-versions"
if [ "$EUID" -ne 0 ]; then echo "[!] Run as sudo"; exit 1; fi

# 1. Detect Ubuntu version + kernel + BTF
UBU=$(lsb_release -rs 2>/dev/null || echo "unknown")
KERN=$(uname -r)
echo "[*] Ubuntu $UBU kernel $KERN"

if [ -f /sys/kernel/btf/vmlinux ]; then
  echo "[*] BTF available — eBPF CO-RE will be used"
  MODE="ebpf"
else
  echo "[*] BTF not found (Ubuntu 16.04/18.04 4.15) — will use auditd -> /proc fallback"
  MODE="fallback"
fi

# 2. Deps — all Ubuntu
apt-get update -y
apt-get install -y curl ca-certificates sqlite3 auditd 2>/dev/null || true
if [ "$MODE" = "ebpf" ]; then
  apt-get install -y clang llvm libbpf-dev linux-headers-$(uname -r) 2>/dev/null || echo "[WARN] eBPF headers not found, fallback still works"
fi

# 3. Create dirs (0750/0640 Option A)
mkdir -p "$INSTALL_DIR/bin" "$INSTALL_DIR/packs/anomaly" "$LOG_DIR"
chmod 0750 "$INSTALL_DIR" "$INSTALL_DIR/bin" "$LOG_DIR"
touch "$DB" && chmod 0640 "$DB" || true

# 4. Download binary (or build if no release)
if [ ! -f "$BIN" ]; then
  echo "[*] Fetching latest release..."
  URL=$(curl -s https://api.github.com/repos/n1cewatch/n1cewatch/releases/latest | grep browser_download_url | grep linux_amd64 | cut -d'"' -f4 | head -1)
  if [ -n "$URL" ]; then
    curl -fsSL "$URL" -o "$BIN" && chmod 0750 "$BIN"
  else
    echo "[!] No release binary — build from source: make build && sudo make install"
    exit 0
  fi
fi
chmod 0750 "$BIN"
chown root:root "$BIN"

# 5. Install packs
cp -r packs/anomaly/* "$INSTALL_DIR/packs/anomaly/" 2>/dev/null || echo "[WARN] packs not found, using defaults"
chmod 0640 "$INSTALL_DIR/packs/anomaly"/* 2>/dev/null || true

# 6. Systemd
cp deploy/systemd/n1cewatch.service /etc/systemd/system/n1cewatch.service
chmod 0644 /etc/systemd/system/n1cewatch.service
systemctl daemon-reload
systemctl enable n1cewatch
systemctl restart n1cewatch

# 7. Verify
sleep 2
systemctl status n1cewatch --no-pager || true
echo "[*] Check: tail -f /var/log/n1cewatch/events.jsonl"
echo "[*] API: curl http://localhost:8081/api/alerts | jq"
echo "[*] Dashboard: http://localhost:8081/report.html"
echo "[*] Non-sudo test: su nobody -c 'kill \$(cat /run/n1cewatch/n1cewatch.pid)' -> should fail"
echo "[OK] N1ceWatch installed (mode: $MODE)"
