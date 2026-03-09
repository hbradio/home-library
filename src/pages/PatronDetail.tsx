import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useApi } from '../lib/api'
import BookCard from '../components/BookCard'

interface PatronData {
  id: string
  first_name: string
  last_name: string
  created_at: string
  checked_out_books: Array<{
    id: string
    isbn: string
    title: string
    author: string
    loan_status: string
  }>
}

interface LoanEvent {
  id: string
  event_type: string
  created_at: string
  book_title?: string
}

export default function PatronDetail() {
  const { id } = useParams()
  const { fetchWithAuth } = useApi()
  const [patron, setPatron] = useState<PatronData | null>(null)
  const [history, setHistory] = useState<LoanEvent[]>([])

  useEffect(() => {
    const load = async () => {
      const [patronResp, histResp] = await Promise.all([
        fetchWithAuth(`/api/patrons?id=${id}`),
        fetchWithAuth(`/api/loans?patron_id=${id}`),
      ])
      if (patronResp.ok) setPatron(await patronResp.json())
      if (histResp.ok) setHistory(await histResp.json())
    }
    load()
  }, [id, fetchWithAuth])

  if (!patron) return <div className="loading">Loading...</div>

  return (
    <div className="detail-page">
      <h2>{patron.first_name} {patron.last_name}</h2>
      <p style={{ color: '#8b7355' }}>Patron since {new Date(patron.created_at).toLocaleDateString()}</p>

      <h3>Currently Checked Out</h3>
      {(!patron.checked_out_books || patron.checked_out_books.length === 0) ? (
        <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No books checked out.</p>
      ) : (
        patron.checked_out_books.map((book) => (
          <BookCard key={book.id} book={book} />
        ))
      )}

      <div className="loan-history">
        <h3>Loan History</h3>
        {history.length === 0 ? (
          <p style={{ color: '#8b7355', fontStyle: 'italic' }}>No loan history.</p>
        ) : (
          <table>
            <thead>
              <tr><th>Date</th><th>Action</th><th>Book</th></tr>
            </thead>
            <tbody>
              {history.map((e) => (
                <tr key={e.id}>
                  <td>{new Date(e.created_at).toLocaleDateString()}</td>
                  <td>{e.event_type === 'checkout' ? 'Checked Out' : 'Returned'}</td>
                  <td>{e.book_title || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
