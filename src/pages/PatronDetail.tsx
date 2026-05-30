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

interface EditFields {
  first_name: string
  last_name: string
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
  const [editing, setEditing] = useState(false)
  const [edit, setEdit] = useState<EditFields>({ first_name: '', last_name: '' })

  const startEdit = () => {
    if (!patron) return
    setEdit({
      first_name: patron.first_name,
      last_name: patron.last_name,
    })
    setEditing(true)
  }

  const saveEdit = async () => {
    const resp = await fetchWithAuth(`/api/patrons?id=${id}`, {
      method: 'PUT',
      body: JSON.stringify({
        first_name: edit.first_name,
        last_name: edit.last_name,
      }),
    })
    if (resp.ok) {
      const updated = await resp.json()
      setPatron(updated)
      setEditing(false)
    }
  }

  const cancelEdit = () => {
    setEditing(false)
  }

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
      <h2>
        {editing ? (
          <>
            <input 
              value={edit.first_name} 
              onChange={(e) => setEdit({ ...edit, first_name: e.target.value })} 
              style={{ font: 'inherit', width: 'auto', marginRight: '0.5em' }} 
            />
            <input 
              value={edit.last_name} 
              onChange={(e) => setEdit({ ...edit, last_name: e.target.value })} 
              style={{ font: 'inherit', width: 'auto' }} 
            />
          </>
        ) : (
          <>
            {patron.first_name} {patron.last_name} <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span>
          </>
        )}
      </h2>
      <p style={{ color: '#8b7355' }}>Patron since {new Date(patron.created_at).toLocaleDateString()}</p>

      {editing ? (
        <div style={{ marginBottom: '1em' }}>
          <button onClick={saveEdit} style={{ marginRight: '0.5em' }}>Save</button>
          <button onClick={cancelEdit}>Cancel</button>
        </div>
      ) : (
        <button onClick={startEdit} style={{ marginBottom: '1em' }}>Edit</button>
      )}

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
