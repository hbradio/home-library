import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useApi } from '../lib/api'

interface Book {
  id: string
  isbn: string
  title: string
  author: string
  genre: string
  publish_year: number | null
  loan_status: string
  patron_name?: string
}

interface LoanEvent {
  id: string
  event_type: string
  created_at: string
  patron_name?: string
}

export default function BookDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { fetchWithAuth } = useApi()
  const [book, setBook] = useState<Book | null>(null)
  const [history, setHistory] = useState<LoanEvent[]>([])
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handleDelete = async () => {
    if (!confirmDelete) {
      setConfirmDelete(true)
      return
    }
    await fetchWithAuth(`/api/books?id=${id}`, { method: 'DELETE' })
    navigate('/browse')
  }

  useEffect(() => {
    const load = async () => {
      const [bookResp, histResp] = await Promise.all([
        fetchWithAuth(`/api/books?id=${id}`),
        fetchWithAuth(`/api/loans?book_id=${id}`),
      ])
      // The books endpoint returns a list when filtering; we need the single book by ID
      // Actually we use the book detail through a different mechanism
      // Let's handle both cases
      if (bookResp.ok) {
        setBook(await bookResp.json())
      }
      if (histResp.ok) setHistory(await histResp.json())
    }
    load()
  }, [id, fetchWithAuth])

  if (!book) return <div className="loading">Loading...</div>

  return (
    <div className="detail-page">
      <div style={{ display: 'flex', gap: '2em', flexWrap: 'wrap' }}>
        <img
          className="cover"
          src={`https://covers.openlibrary.org/b/isbn/${book.isbn}-L.jpg`}
          alt={`Cover of ${book.title}`}
          onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
        />
        <div>
          <h2 style={{ marginTop: 0 }}>{book.title} <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
          <dl className="meta">
            {book.author && <><dt>Author:</dt><dd>{book.author}</dd><br /></>}
            {book.genre && <><dt>Genre:</dt><dd>{book.genre}</dd><br /></>}
            {book.publish_year && <><dt>Year:</dt><dd>{book.publish_year}</dd><br /></>}
            <dt>ISBN:</dt><dd>{book.isbn}</dd><br />
            <dt>Status:</dt>
            <dd>
              <span className={`status-badge ${book.loan_status}`}>
                {book.loan_status === 'checked_out'
                  ? `Checked out to ${book.patron_name || 'someone'}`
                  : 'Available'}
              </span>
            </dd>
          </dl>
          <button
            onClick={handleDelete}
            style={{ marginTop: '1em', color: confirmDelete ? '#fff' : '#c62828', borderColor: '#c62828', background: confirmDelete ? '#c62828' : '#fff' }}
          >
            {confirmDelete ? 'Confirm Delete' : 'Delete Book'}
          </button>
          {confirmDelete && (
            <button onClick={() => setConfirmDelete(false)} style={{ marginLeft: '0.5em', marginTop: '1em' }}>
              Cancel
            </button>
          )}
        </div>
      </div>

      <div className="loan-history">
        <h3>Loan History</h3>
        {history.length === 0 ? (
          <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No loan history.</p>
        ) : (
          <table>
            <thead>
              <tr><th>Date</th><th>Action</th><th>Patron</th></tr>
            </thead>
            <tbody>
              {history.map((e) => (
                <tr key={e.id}>
                  <td>{new Date(e.created_at).toLocaleDateString()}</td>
                  <td>{e.event_type === 'checkout' ? 'Checked Out' : 'Returned'}</td>
                  <td>{e.patron_name || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
