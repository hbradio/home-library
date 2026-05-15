import { useState, useEffect } from 'react'
import { useApi } from '../lib/api'

interface PublicLinkData {
  id: string
  slug: string
  created_at: string
}

export default function Sharing() {
  const { fetchWithAuth } = useApi()
  const [link, setLink] = useState<PublicLinkData | null>(null)
  const [slugInput, setSlugInput] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const loadLink = async () => {
    const resp = await fetchWithAuth('/api/public-link')
    if (resp.ok) {
      const data = await resp.json()
      setLink(data)
      setSlugInput(data.slug)
    } else if (resp.status === 404) {
      setLink(null)
    }
    setLoading(false)
  }

  useEffect(() => { loadLink() }, [fetchWithAuth])

  const createLink = async () => {
    setError('')
    setSaving(true)
    const resp = await fetchWithAuth('/api/public-link', { method: 'POST' })
    if (resp.ok) {
      const data = await resp.json()
      setLink(data)
      setSlugInput(data.slug)
    } else {
      const data = await resp.json()
      setError(data.error || 'Failed to create link')
    }
    setSaving(false)
  }

  const saveSlug = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    const trimmed = slugInput.trim().toLowerCase()
    if (trimmed === link?.slug) return
    setSaving(true)
    const resp = await fetchWithAuth('/api/public-link', {
      method: 'PUT',
      body: JSON.stringify({ slug: trimmed }),
    })
    if (resp.ok) {
      const data = await resp.json()
      setLink(data)
      setSlugInput(data.slug)
      setSuccess('Slug updated!')
      setTimeout(() => setSuccess(''), 3000)
    } else {
      const data = await resp.json()
      setError(data.error || 'Failed to update slug')
    }
    setSaving(false)
  }

  const revokeLink = async () => {
    setError('')
    const resp = await fetchWithAuth('/api/public-link', { method: 'DELETE' })
    if (resp.ok) {
      setLink(null)
      setSlugInput('')
    } else {
      setError('Failed to revoke link')
    }
  }

  const publicUrl = link ? `${window.location.origin}/lib/${link.slug}` : ''

  return (
    <div>
      <h2>Sharing <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <p style={{ color: '#8b7355' }}>Anyone with this link has read-only access to view your library.</p>

      {error && <div className="message error">{error}</div>}
      {success && <div className="message success">{success}</div>}

      {loading ? (
        <div className="loading">Loading...</div>
      ) : !link ? (
        <button onClick={createLink} disabled={saving}>
          {saving ? 'Creating...' : 'Create public link'}
        </button>
      ) : (
        <div>
          <div style={{ margin: '1em 0', padding: '1em', background: '#faf7f2', border: '1px solid #c4b5a0' }}>
            <div style={{ fontSize: '0.85em', color: '#8b7355', marginBottom: '0.3em' }}>Your public link:</div>
            <a href={publicUrl} target="_blank" rel="noopener noreferrer" style={{ wordBreak: 'break-all', fontWeight: 600 }}>
              {publicUrl}
            </a>
          </div>

          <form onSubmit={saveSlug} style={{ display: 'flex', gap: '0.5em', alignItems: 'end', marginBottom: '1.5em', flexWrap: 'wrap' }}>
            <div style={{ flex: 1, minWidth: '200px' }}>
              <label style={{ display: 'block', fontSize: '0.85em', color: '#8b7355', marginBottom: '0.3em' }}>
                Customize your link ID:
              </label>
              <input
                value={slugInput}
                onChange={(e) => { setSlugInput(e.target.value); setError(''); setSuccess('') }}
                placeholder="e.g. bradys-books"
                style={{ width: '100%' }}
              />
            </div>
            <button type="submit" disabled={saving || slugInput.trim().toLowerCase() === link.slug}>
              {saving ? 'Saving...' : 'Save'}
            </button>
          </form>
          <p style={{ fontSize: '0.8em', color: '#8b7355', marginTop: '-1em' }}>
            Lowercase letters, numbers, and hyphens. 3–30 characters.
          </p>

          <button
            onClick={revokeLink}
            style={{ color: '#b8a88a', border: 'none', background: 'transparent', fontSize: '0.8em', cursor: 'pointer', textDecoration: 'underline', padding: 0 }}
          >
            Revoke link
          </button>
        </div>
      )}
    </div>
  )
}
