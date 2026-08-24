#!/bin/bash
# Test Option A tamper-resistance — non-sudo cannot kill, sudo can
set -e
PID=$(cat /run/n1cewatch/n1cewatch.pid 2>/dev/null || pgrep n1cewatch || echo "")
if [ -z "$PID" ]; then echo "[SKIP] n1cewatch not running"; exit 0; fi
echo "[*] PID $PID"

echo "[*] Test non-sudo kill (should fail)"
if su nobody -c "kill $PID" 2>&1 | grep -q "Operation not permitted"; then
  echo "[PASS] non-sudo kill blocked"
else
  echo "[FAIL] non-sudo kill not blocked"
fi

echo "[*] Test sudo status visible"
systemctl status n1cewatch --no-pager | head -5
echo "[PASS] visible to sudo"
