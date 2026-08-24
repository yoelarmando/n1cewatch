import useSWR from 'swr'
const fetcher = (u: string) => fetch(u).then(r => r.json())
export default function Report() {
  const { data } = useSWR('/api/stats', fetcher, { refreshInterval: 5000 })
  return (
    <div style={{ fontFamily: 'system-ui', background: '#0f172a', color: '#e2e8f0', minHeight: '100vh', padding: 24 }}>
      <h1>Compliance Report</h1>
      <p>Generated from SQLite store — CIS/NIST mapping</p>
      <div style={{ background: '#1e293b', padding: 16, borderRadius: 12 }}>
        <p>Total: {data?.total_alerts ?? 0}</p>
        <ul>
          <li>CIS 8.2 — Audit process creation: {data?.total_alerts ?? 0} events</li>
          <li>NIST SI-4 — System monitoring: active</li>
          <li>MITRE coverage: T1059, T1505.003, T1071, T1547</li>
        </ul>
        <a href="http://localhost:8081/report.html" style={{ color: '#38bdf8' }}>Download HTML (from agent)</a>
      </div>
    </div>
  )
}
