export default function ProcessTree({ alerts }: { alerts: any[] }) {
  return (
    <div style={{ background: '#1e293b', padding: 16, borderRadius: 12 }}>
      <h3 style={{ margin: 0, color: '#38bdf8' }}>Process Tree (Ancestry)</h3>
      {alerts.slice(0, 8).map((a: any, i: number) => (
        <div key={i} style={{ fontFamily: 'monospace', fontSize: 12, margin: '6px 0' }}>
          <span style={{ color: '#64748b' }}>{a.event?.parent_image?.split('/').pop() || 'systemd'}</span>
          <span style={{ margin: '0 6px' }}>→</span>
          <span style={{ color: '#e2e8f0' }}>{a.event?.image?.split('/').pop()}</span>
          <span style={{ color: '#38bdf8', marginLeft: 8 }}>{a.event?.command_line?.slice(0, 80)}</span>
        </div>
      ))}
    </div>
  )
}
