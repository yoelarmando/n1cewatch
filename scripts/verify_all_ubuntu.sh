#!/bin/bash
# Verify all-Ubuntu matrix — requires Docker
set -e
for ver in 16.04 18.04 20.04 22.04 24.04; do
  echo "=== Testing ubuntu:$ver ==="
  docker run --rm ubuntu:$ver bash -c "cat /etc/os-release | grep VERSION"
done
echo "[OK] All Ubuntu versions reachable"
echo "[*] Now test n1cewatch --no-ebpf fallback on your server: ./bin/n1cewatch --no-ebpf --help"
