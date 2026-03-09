import { useState, useCallback } from 'react'
import { useApi } from '../lib/api'
import IsbnInput from '../components/IsbnInput'
import Scanner from '../components/Scanner'
import SessionList from '../components/SessionList'

interface SessionItem {
  text: string
  type: 'success' | 'error' | 'info'
}

export default function AddBook() {
  const { fetchWithAuth } = useApi()
  const [useCamera, setUseCamera] = useState(false)
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

  return (
    <div>
      <h2>Add Book <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <div className="scan-area">
        <IsbnInput onScan={handleScan} />
        <button
          className="toggle-camera"
          onClick={() => setUseCamera(!useCamera)}
        >
          {useCamera ? 'Use Keyboard' : 'Use Camera'}
        </button>
      </div>

      {pending > 0 && (
        <div className="message info">Looking up {pending} book{pending > 1 ? 's' : ''}...</div>
      )}
      {message && (
        <div className={`message ${message.type}`}>{message.text}</div>
      )}

      <Scanner active={useCamera} onScan={handleScan} />

      <SessionList items={sessionItems} />
    </div>
  )
}
