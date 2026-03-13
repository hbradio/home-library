import { useState, useCallback } from 'react'
import { useApi } from '../lib/api'
import IsbnInput from '../components/IsbnInput'
import Scanner from '../components/Scanner'
import SessionList from '../components/SessionList'

interface SessionItem {
  text: string
  type: 'success' | 'error' | 'info'
}

interface ManualFields {
  title: string
  author: string
  genre: string
  publisher: string
  publish_year: string
}

const emptyManual: ManualFields = { title: '', author: '', genre: '', publisher: '', publish_year: '' }

export default function AddBook() {
  const { fetchWithAuth } = useApi()
  const [useCamera, setUseCamera] = useState(false)
  const [manualMode, setManualMode] = useState(false)
  const [manual, setManual] = useState<ManualFields>({ ...emptyManual })
  const [pending, setPending] = useState(0)
  const [message, setMessage] = useState<{ text: string; type: string } | null>(null)
  const [sessionItems, setSessionItems] = useState<SessionItem[]>([])

  const handleScan = useCallback(async (isbn: string) => {
    setPending((n) => n + 1)
    setSessionItems((prev) => [{ text: `${isbn} — Looking up...`, type: 'info' }, ...prev])

    try {
      const lookupResp = await fetchWithAuth(`/api/book-lookup?isbn=${isbn}`)
      if (!lookupResp.ok) {
        const err = await lookupResp.json()
        setMessage({ text: err.error || 'Book not found', type: 'error' })
        setSessionItems((prev) => prev.map((item) =>
          item.text === `${isbn} — Looking up...` ? { text: `${isbn} — Not found`, type: 'error' } : item
        ))
        setPending((n) => n - 1)
        return
      }
      const bookData = await lookupResp.json()

      const saveResp = await fetchWithAuth('/api/books', {
        method: 'POST',
        body: JSON.stringify(bookData),
      })

      if (saveResp.status === 409) {
        setMessage({ text: `"${bookData.title}" is already in your library`, type: 'info' })
        setSessionItems((prev) => prev.map((item) =>
          item.text === `${isbn} — Looking up...` ? { text: `${bookData.title} — Already in library`, type: 'info' } : item
        ))
      } else if (saveResp.ok) {
        setMessage({ text: `Added "${bookData.title}" by ${bookData.author || 'Unknown'}`, type: 'success' })
        setSessionItems((prev) => prev.map((item) =>
          item.text === `${isbn} — Looking up...` ? { text: `${bookData.title} — Added`, type: 'success' } : item
        ))
      } else {
        const err = await saveResp.json()
        setMessage({ text: err.error || 'Failed to save book', type: 'error' })
        setSessionItems((prev) => prev.map((item) =>
          item.text === `${isbn} — Looking up...` ? { text: `${isbn} — Error saving`, type: 'error' } : item
        ))
      }
    } catch {
      setMessage({ text: 'Network error', type: 'error' })
      setSessionItems((prev) => prev.map((item) =>
        item.text === `${isbn} — Looking up...` ? { text: `${isbn} — Network error`, type: 'error' } : item
      ))
    }
    setPending((n) => n - 1)
  }, [fetchWithAuth])

  const handleManualSave = async () => {
    if (!manual.title.trim()) {
      setMessage({ text: 'Title is required', type: 'error' })
      return
    }
    const isbn = `MANUAL-${Date.now()}`
    const year = manual.publish_year ? parseInt(manual.publish_year, 10) : null
    const body = {
      isbn,
      title: manual.title.trim(),
      author: manual.author.trim(),
      genre: manual.genre.trim(),
      publisher: manual.publisher.trim(),
      publish_year: isNaN(year as number) ? null : year,
    }
    try {
      const resp = await fetchWithAuth('/api/books', {
        method: 'POST',
        body: JSON.stringify(body),
      })
      if (resp.ok) {
        setMessage({ text: `Added "${body.title}"`, type: 'success' })
        setSessionItems((prev) => [{ text: `${body.title} — Added (manual)`, type: 'success' }, ...prev])
        setManual({ ...emptyManual })
      } else {
        const err = await resp.json()
        setMessage({ text: err.error || 'Failed to save book', type: 'error' })
      }
    } catch {
      setMessage({ text: 'Network error', type: 'error' })
    }
  }

  return (
    <div>
      <h2>Add Book <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>

      {manualMode ? (
        <div>
          <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '0.4em 0.8em', alignItems: 'center', maxWidth: '500px' }}>
            <label>Title *</label>
            <input value={manual.title} onChange={(e) => setManual({ ...manual, title: e.target.value })} />
            <label>Author</label>
            <input value={manual.author} onChange={(e) => setManual({ ...manual, author: e.target.value })} />
            <label>Genre</label>
            <input value={manual.genre} onChange={(e) => setManual({ ...manual, genre: e.target.value })} />
            <label>Publisher</label>
            <input value={manual.publisher} onChange={(e) => setManual({ ...manual, publisher: e.target.value })} />
            <label>Year</label>
            <input value={manual.publish_year} onChange={(e) => setManual({ ...manual, publish_year: e.target.value })} type="number" />
            <div style={{ gridColumn: '1 / -1', marginTop: '0.5em', display: 'flex', gap: '0.5em' }}>
              <button onClick={handleManualSave}>Add Book</button>
              <button onClick={() => setManualMode(false)}>Back to Scanner</button>
            </div>
          </div>
        </div>
      ) : (
        <>
          <div className="scan-area">
            <IsbnInput onScan={handleScan} />
            <button
              className="toggle-camera"
              onClick={() => setUseCamera(!useCamera)}
            >
              {useCamera ? 'Use Keyboard' : 'Use Camera'}
            </button>
          </div>
          <button
            onClick={() => { setManualMode(true); setUseCamera(false) }}
            style={{ marginTop: '0.5em', fontSize: '0.9em' }}
          >
            No Barcode? Enter Manually
          </button>
        </>
      )}

      {pending > 0 && (
        <div className="message info">Looking up {pending} book{pending > 1 ? 's' : ''}...</div>
      )}
      {message && (
        <div className={`message ${message.type}`}>{message.text}</div>
      )}

      <Scanner active={useCamera && !manualMode} onScan={handleScan} />

      <SessionList items={sessionItems} />
    </div>
  )
}
