import { useState, useEffect } from 'react'
import { useApi } from '../lib/api'

interface Share {
  id: string
  shared_with_email: string
  created_at: string
}

export default function Sharing() {
  const { fetchWithAuth } = useApi()
  const [shares, setShares] = useState<Share[]>([])
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const loadShares = async () => {
    const resp = await fetchWithAuth('/api/shares')
    if (resp.ok) setShares(await resp.json())
    setLoading(false)
  }

  useEffect(() => { loadShares() }, [fetchWithAuth])

  const addShare = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    const resp = await fetchWithAuth('/api/shares', {
      method: 'POST',
      body: JSON.stringify({ email: email.trim() }),
    })
    if (resp.ok) {
      setEmail('')
      loadShares()
    } else {
      const data = await resp.json()
      setError(data.error || 'Failed to share')
    }
  }

  const removeShare = async (id: string) => {
    await fetchWithAuth(`/api/shares?id=${id}`, { method: 'DELETE' })
    loadShares()
  }

  return (
    <div>
      <h2>Sharing <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <p style={{ color: '#8b7355' }}>Share your library (read-only) with others by entering their email address.</p>

      <form onSubmit={addShare} style={{ display: 'flex', gap: '0.5em', marginBottom: '1.5em' }}>
        <input
          type="email"
          placeholder="Email address..."
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          style={{ flex: 1 }}
        />
        <button type="submit">Share</button>
      </form>
      {error && <p style={{ color: '#c62828' }}>{error}</p>}

      {loading ? (
        <div className="loading">Loading...</div>
      ) : shares.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>You haven't shared your library with anyone yet.</p>
      ) : (
        <div>
          <h3>Shared with</h3>
          {shares.map((s) => (
            <div key={s.id} className="book-card" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>{s.shared_with_email}</span>
              <button
                onClick={() => removeShare(s.id)}
                style={{ color: '#b8a88a', border: 'none', background: 'transparent', fontSize: '0.8em', cursor: 'pointer', textDecoration: 'underline' }}
              >
                Revoke
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
