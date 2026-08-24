package report

import (
	"fmt"
	"os"
	"time"
)

// Generator writes periodic HTML compliance reports
type Generator struct {
	outPath string
}

func New(outPath string) *Generator {
	return &Generator{outPath: outPath}
}

func (g *Generator) Generate(stats map[string]int) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>N1ceWatch Compliance</title>
<style>body{font-family:system-ui;background:#0f172a;color:#e2e8f0;padding:24px} .card{background:#1e293b;padding:16px;border-radius:12px;margin:12px 0} h1{color:#38bdf8}</style>
</head><body>
<h1>N1ceWatch Blue — Periodic Report</h1>
<p>Generated: %s</p>
<div class="card"><h3>Summary</h3><p>Total Alerts: %d</p><p>Host: %s</p></div>
<div class="card"><h3>Compliance Mapping</h3><ul><li>CIS 8.2: Process creation audit</li><li>NIST 800-53 SI-4: System monitoring</li><li>MITRE T1059, T1505.003, T1071</li></ul></div>
<div class="card"><h3>Recommendation</h3><p>Review high/critical alerts in dashboard at :8081</p></div>
</body></html>`, time.Now().Format(time.RFC3339), stats["total"], getHostname())
	return os.WriteFile(g.outPath, []byte(html), 0644)
}

func getHostname() string { h, _ := os.Hostname(); return h }
