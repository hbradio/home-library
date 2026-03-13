import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useApi } from '../lib/api'

interface SharedLibrary {
  share_id: string
  owner_id: string
  owner_email: string
}

export default function SharedLibraries() {
  const { fetchWithAuth } = useApi()
  const navigate = useNavigate()
  const [libraries, setLibraries] = useState<SharedLibrary[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      const resp = await fetchWithAuth('/api/shared-libraries')
      if (resp.ok) setLibraries(await resp.json())
      setLoading(false)
    }
    load()
  }, [fetchWithAuth])

  return (
    <div>
      <h2>Shared Libraries <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>

      {loading ? (
        <div className="loading">Loading...</div>
      ) : libraries.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No one has shared their library with you yet.</p>
      ) : (
        libraries.map((lib) => (
          <div
            key={lib.share_id}
            className="book-card"
            onClick={() => navigate(`/shared/${lib.owner_id}`)}
            style={{ cursor: 'pointer' }}
          >
            <strong>{lib.owner_email}'s Library</strong>
          </div>
        ))
      )}
    </div>
  )
}
