import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useApi } from '../lib/api'

interface Book {
  id: string
  isbn: string
  title: string
  author: string
  genre: string
  publisher: string
  dewey_decimal: string
  lc_classification: string
  publish_year: number | null
  loan_status: string
  patron_name?: string
}

export default function SharedBrowse() {
  const { ownerId } = useParams()
  const navigate = useNavigate()
  const { fetchWithAuth } = useApi()
  const [books, setBooks] = useState<Book[]>([])
  const [ownerEmail, setOwnerEmail] = useState('')
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [genre, setGenre] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Get owner email from the shared libraries list
    const loadOwner = async () => {
      const resp = await fetchWithAuth('/api/shared-libraries')
      if (resp.ok) {
        const libs = await resp.json()
        const lib = libs.find((l: { owner_id: string }) => l.owner_id === ownerId)
        if (lib) setOwnerEmail(lib.owner_email)
      }
    }
    loadOwner()
  }, [ownerId, fetchWithAuth])

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      const params = new URLSearchParams({ owner_id: ownerId! })
      if (title) params.set('title', title)
      if (author) params.set('author', author)
      if (genre) params.set('genre', genre)
      const resp = await fetchWithAuth(`/api/shared-libraries?${params}`)
      if (resp.ok) {
        setBooks(await resp.json())
      } else if (resp.status === 403) {
        navigate('/shared')
      }
      setLoading(false)
    }
    const timer = setTimeout(load, 300)
    return () => clearTimeout(timer)
  }, [ownerId, title, author, genre, fetchWithAuth, navigate])

  return (
    <div>
      <h2>
        {ownerEmail ? `${ownerEmail}'s Library` : 'Shared Library'}
        <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}> (Esc to go back)</span>
      </h2>
      <div className="filters">
        <input placeholder="Filter by title..." value={title} onChange={(e) => setTitle(e.target.value)} />
        <input placeholder="Filter by author..." value={author} onChange={(e) => setAuthor(e.target.value)} />
        <input placeholder="Filter by genre..." value={genre} onChange={(e) => setGenre(e.target.value)} />
      </div>
      {loading ? (
        <div className="loading">Loading books...</div>
      ) : books.length === 0 ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No books found.</p>
      ) : (
        books.map((book) => (
          <div
            key={book.id}
            className="book-card"
            onClick={() => navigate(`/shared/${ownerId}/book/${book.id}`)}
            style={{ cursor: 'pointer' }}
          >
            <div>
              <strong>{book.title}</strong>
              {book.author && <span> — {book.author}</span>}
            </div>
            <span className={`status-badge ${book.loan_status}`}>
              {book.loan_status === 'checked_out'
                ? `Out to ${book.patron_name || 'someone'}`
                : 'Available'}
            </span>
          </div>
        ))
      )}
    </div>
  )
}
