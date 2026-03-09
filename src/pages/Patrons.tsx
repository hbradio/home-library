import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useApi } from '../lib/api'

interface Patron {
  id: string
  first_name: string
  last_name: string
}

export default function Patrons() {
  const { fetchWithAuth } = useApi()
  const navigate = useNavigate()
  const [patrons, setPatrons] = useState<Patron[]>([])
  const [search, setSearch] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [loading, setLoading] = useState(true)

  const load = async (q = search) => {
    setLoading(true)
    const resp = await fetchWithAuth(`/api/patrons?q=${encodeURIComponent(q)}`)
    if (resp.ok) setPatrons(await resp.json())
    setLoading(false)
  }

  useEffect(() => {
    const timer = setTimeout(() => load(search), 300)
    return () => clearTimeout(timer)
  }, [search, fetchWithAuth])

  const addPatron = async () => {
    if (!firstName.trim() || !lastName.trim()) return
    const resp = await fetchWithAuth('/api/patrons', {
      method: 'POST',
      body: JSON.stringify({ first_name: firstName.trim(), last_name: lastName.trim() }),
    })
    if (resp.ok) {
      setFirstName('')
      setLastName('')
      load()
    }
  }

  return (
    <div>
      <h2>Patrons <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <input
        type="text"
        placeholder="Search patrons..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{ width: '100%', maxWidth: '400px', marginBottom: '1em' }}
      />

      <div className="inline-form" style={{ marginBottom: '1em' }}>
        <input placeholder="First name" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
        <input placeholder="Last name" value={lastName} onChange={(e) => setLastName(e.target.value)} />
        <button onClick={addPatron}>Add Patron</button>
      </div>

      {loading ? (
        <div className="loading">Loading...</div>
      ) : patrons.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No patrons found.</p>
      ) : (
        patrons.map((p) => (
          <div
            key={p.id}
            className="patron-list-item"
            onClick={() => navigate(`/patron/${p.id}`)}
          >
            {p.first_name} {p.last_name}
          </div>
        ))
      )}
    </div>
  )
}
