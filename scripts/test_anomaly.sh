#!/bin/bash
set -e
echo "[*] Trigger generic anomalies (on your own server only)"
echo "test" | base64 | base64 -d | sh -c "echo pwned" &
sleep 1
echo "[*] Check JSONL"
tail -n 20 /var/log/n1cewatch/events.jsonl | jq -s '.[] | select(.rule_id=="nws-003")' || echo "no match yet, check SIGMA packs"
echo "[*] Check API"
curl -s http://localhost:8081/api/alerts | jq '.[0]'
