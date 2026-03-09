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
  const [processing, setProcessing] = useState(false)
  const [message, setMessage] = useState<{ text: string; type: string } | null>(null)
  const [sessionItems, setSessionItems] = useState<SessionItem[]>([])

  const handleScan = useCallback(async (isbn: string) => {
    setProcessing(true)
    setMessage({ text: `Looking up ISBN ${isbn}...`, type: 'info' })

    try {
      // Look up book metadata
      const lookupResp = await fetchWithAuth(`/api/book-lookup?isbn=${isbn}`)
      if (!lookupResp.ok) {
        const err = await lookupResp.json()
        setMessage({ text: err.error || 'Book not found', type: 'error' })
        setSessionItems((prev) => [{ text: `${isbn} — Not found`, type: 'error' }, ...prev])
        setProcessing(false)
        return
      }
      const bookData = await lookupResp.json()

      // Save to library
      const saveResp = await fetchWithAuth('/api/books', {
        method: 'POST',
        body: JSON.stringify(bookData),
      })

      if (saveResp.status === 409) {
        setMessage({ text: `"${bookData.title}" is already in your library`, type: 'info' })
        setSessionItems((prev) => [{ text: `${bookData.title} — Already in library`, type: 'info' }, ...prev])
      } else if (saveResp.ok) {
        setMessage({ text: `Added "${bookData.title}" by ${bookData.author || 'Unknown'}`, type: 'success' })
        setSessionItems((prev) => [{ text: `${bookData.title} — Added`, type: 'success' }, ...prev])
      } else {
        const err = await saveResp.json()
        setMessage({ text: err.error || 'Failed to save book', type: 'error' })
        setSessionItems((prev) => [{ text: `${isbn} — Error saving`, type: 'error' }, ...prev])
      }
    } catch {
      setMessage({ text: 'Network error', type: 'error' })
    }
    setProcessing(false)
  }, [fetchWithAuth])

  return (
    <div>
      <h2>Add Book</h2>
      <div className="scan-area">
        <IsbnInput onScan={handleScan} disabled={processing} />
        <button
          className="toggle-camera"
          onClick={() => setUseCamera(!useCamera)}
        >
          {useCamera ? 'Use Keyboard' : 'Use Camera'}
        </button>
      </div>

      <Scanner active={useCamera} onScan={handleScan} />

      {message && (
        <div className={`message ${message.type}`}>{message.text}</div>
      )}

      <SessionList items={sessionItems} />
    </div>
  )
}
