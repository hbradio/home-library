import { useState, useEffect } from 'react'
import { useApi } from '../lib/api'

interface Patron {
  id: string
  first_name: string
  last_name: string
}

interface Props {
  onSelect: (patronId: string, patronName: string) => void
}

export default function PatronSearch({ onSelect }: Props) {
  const { fetchWithAuth } = useApi()
  const [query, setQuery] = useState('')
  const [patrons, setPatrons] = useState<Patron[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')

  useEffect(() => {
    const search = async () => {
      const resp = await fetchWithAuth(`/api/patrons?q=${encodeURIComponent(query)}`)
      if (resp.ok) setPatrons(await resp.json())
    }
    search()
  }, [query, fetchWithAuth])

  const addPatron = async () => {
    if (!firstName.trim() || !lastName.trim()) return
    const resp = await fetchWithAuth('/api/patrons', {
      method: 'POST',
      body: JSON.stringify({ first_name: firstName.trim(), last_name: lastName.trim() }),
    })
    if (resp.ok) {
      const patron = await resp.json()
      onSelect(patron.id, `${patron.first_name} ${patron.last_name}`)
    }
  }

  return (
    <div className="patron-search">
      <h3 style={{ margin: '0 0 0.5em' }}>Select Patron</h3>
      <input
        type="text"
        placeholder="Search patrons..."
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        style={{ width: '100%', marginBottom: '0.5em' }}
      />
      {patrons.map((p) => (
        <div
          key={p.id}
          className="patron-list-item"
          onClick={() => onSelect(p.id, `${p.first_name} ${p.last_name}`)}
        >
          {p.first_name} {p.last_name}
        </div>
      ))}
      {patrons.length === 0 && query && (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No patrons found.</p>
      )}
      <button
        onClick={() => setShowAdd(!showAdd)}
        style={{ marginTop: '0.5em', fontSize: '0.9em' }}
      >
        + Add New Patron
      </button>
      {showAdd && (
        <div className="inline-form">
          <input
            placeholder="First name"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
          />
          <input
            placeholder="Last name"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
          />
          <button onClick={addPatron}>Add</button>
        </div>
      )}
    </div>
  )
}
