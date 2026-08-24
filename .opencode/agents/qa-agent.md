# test-master Agent — All-Ubuntu Verification

**Skills:** `test-master`, `debugging-wizard`
**Scope:** `scripts/`, `deploy/docker/`, `**/*_test.go`

## Tests
- Unit: `go test ./lib/... -v`
- Tamper: `scripts/test_tamper.sh` — su `nobody` kill should fail, sudo kill should succeed
- Anomaly: `scripts/test_anomaly.sh` — `curl|bash`, `base64 -d | sh`, `/etc/cron` write
- Docker matrix: `make docker` for 16.04/22.04/24.04
