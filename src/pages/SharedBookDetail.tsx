import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
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

export default function SharedBookDetail() {
  const { ownerId, bookId } = useParams()
  const { fetchWithAuth } = useApi()
  const [book, setBook] = useState<Book | null>(null)
  const [ownerEmail, setOwnerEmail] = useState('')

  useEffect(() => {
    const load = async () => {
      const [bookResp, libsResp] = await Promise.all([
        fetchWithAuth(`/api/shared-libraries?owner_id=${ownerId}&book_id=${bookId}`),
        fetchWithAuth('/api/shared-libraries'),
      ])
      if (bookResp.ok) setBook(await bookResp.json())
      if (libsResp.ok) {
        const libs = await libsResp.json()
        const lib = libs.find((l: { owner_id: string }) => l.owner_id === ownerId)
        if (lib) setOwnerEmail(lib.owner_email)
      }
    }
    load()
  }, [ownerId, bookId, fetchWithAuth])

  const isManual = book?.isbn?.startsWith('MANUAL-')

  if (!book) return <div className="loading">Loading...</div>

  return (
    <div className="detail-page">
      <div style={{ display: 'flex', gap: '2em', flexWrap: 'wrap' }}>
        {!isManual && (
          <img
            className="cover"
            src={`https://covers.openlibrary.org/b/isbn/${book.isbn}-L.jpg`}
            alt={`Cover of ${book.title}`}
            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
          />
        )}
        <div>
          <h2 style={{ marginTop: 0 }}>
            {book.title}
            <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}> (Esc to go back)</span>
          </h2>
          {ownerEmail && <p style={{ color: '#8b7355', fontSize: '0.9em' }}>From {ownerEmail}'s library</p>}
          <dl className="meta">
            {book.author && <><dt>Author:</dt><dd>{book.author}</dd><br /></>}
            {book.genre && <><dt>Genre:</dt><dd>{book.genre}</dd><br /></>}
            {book.publisher && <><dt>Publisher:</dt><dd>{book.publisher}</dd><br /></>}
            {book.publish_year && <><dt>Year:</dt><dd>{book.publish_year}</dd><br /></>}
            {book.dewey_decimal && <><dt>Dewey Decimal:</dt><dd>{book.dewey_decimal}</dd><br /></>}
            {book.lc_classification && <><dt>LoC:</dt><dd>{book.lc_classification}</dd><br /></>}
            <dt>ISBN:</dt><dd>{isManual ? <span style={{ color: '#8b7355', fontStyle: 'italic' }}>Manual entry</span> : book.isbn}</dd><br />
            <dt>Status:</dt>
            <dd>
              <span className={`status-badge ${book.loan_status}`}>
                {book.loan_status === 'checked_out'
                  ? `Checked out to ${book.patron_name || 'someone'}`
                  : 'Available'}
              </span>
            </dd>
          </dl>
        </div>
      </div>
    </div>
  )
}
