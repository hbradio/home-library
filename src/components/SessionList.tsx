interface SessionItem {
  text: string
  type: 'success' | 'error' | 'info'
}

export default function SessionList({ items }: { items: SessionItem[] }) {
  if (items.length === 0) return null

  return (
    <div className="session-list">
      <h3>This Session</h3>
      {items.map((item, i) => (
        <div key={i} className={`session-item ${item.type}`}>
          {item.text}
        </div>
      ))}
    </div>
  )
}
