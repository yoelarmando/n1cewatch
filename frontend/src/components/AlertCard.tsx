type Props = { alert: any }
export default function AlertCard({ alert }: Props) {
  const color = alert.level === 'critical' ? '#ef4444' : alert.level === 'high' ? '#f97316' : '#eab308'
  return (
    <div style={{ background: '#1e293b', borderLeft: `4px solid ${color}`, padding: 12, borderRadius: 8, margin: '8px 0' }}>
      <div style={{ fontWeight: 600 }}>{alert.rule_name} <span style={{ color, fontSize: 12 }}> [{alert.level}]</span></div>
      <div style={{ fontSize: 12, color: '#94a3b8' }}>{alert.mitre} • {alert.event?.image}</div>
      <div style={{ fontFamily: 'monospace', fontSize: 11, marginTop: 4 }}>{alert.event?.command_line}</div>
    </div>
  )
}
