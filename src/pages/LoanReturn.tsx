import { useState, useCallback } from 'react'
import { useApi } from '../lib/api'
import IsbnInput from '../components/IsbnInput'
import Scanner from '../components/Scanner'
import PatronSearch from '../components/PatronSearch'
import SessionList from '../components/SessionList'

interface SessionItem {
  text: string
  type: 'success' | 'error' | 'info'
}

export default function LoanReturn() {
  const { fetchWithAuth } = useApi()
  const [useCamera, setUseCamera] = useState(false)
  const [processing, setProcessing] = useState(false)
  const [message, setMessage] = useState<{ text: string; type: string } | null>(null)
  const [sessionItems, setSessionItems] = useState<SessionItem[]>([])
  const [pendingISBN, setPendingISBN] = useState<string | null>(null)
  const [pendingBookTitle, setPendingBookTitle] = useState<string>('')

  const handleScan = useCallback(async (isbn: string) => {
    setProcessing(true)
    setPendingISBN(null)
    setMessage({ text: `Scanning ISBN ${isbn}...`, type: 'info' })

    try {
      const resp = await fetchWithAuth('/api/loans', {
        method: 'POST',
        body: JSON.stringify({ isbn }),
      })

      if (!resp.ok) {
        const err = await resp.json()
        setMessage({ text: err.error || 'Error processing scan', type: 'error' })
        setSessionItems((prev) => [{ text: `${isbn} — ${err.error || 'Error'}`, type: 'error' }, ...prev])
        setProcessing(false)
        return
      }

      const result = await resp.json()

      if (result.action === 'return') {
        const title = result.book?.title || isbn
        const patron = result.patron_name || 'patron'
        setMessage({ text: `Returned "${title}" from ${patron}`, type: 'success' })
        setSessionItems((prev) => [{ text: `${title} — Returned from ${patron}`, type: 'success' }, ...prev])
      } else if (result.action === 'checkout') {
        const title = result.book?.title || isbn
        const patron = result.patron_name || 'patron'
        setMessage({ text: `Checked out "${title}" to ${patron}`, type: 'success' })
        setSessionItems((prev) => [{ text: `${title} — Checked out to ${patron}`, type: 'success' }, ...prev])
      } else if (result.action === 'needs_patron') {
        setPendingISBN(isbn)
        setPendingBookTitle(result.book?.title || isbn)
        setMessage({ text: `Select a patron for "${result.book?.title || isbn}"`, type: 'info' })
      }
    } catch {
      setMessage({ text: 'Network error', type: 'error' })
    }
    setProcessing(false)
  }, [fetchWithAuth])

  const handlePatronSelect = useCallback(async (patronId: string, patronName: string) => {
    if (!pendingISBN) return
    setProcessing(true)

    try {
      const resp = await fetchWithAuth('/api/loans', {
        method: 'POST',
        body: JSON.stringify({ isbn: pendingISBN, patron_id: patronId }),
      })

      if (resp.ok) {
        const result = await resp.json()
        const title = result.book?.title || pendingBookTitle
        setMessage({ text: `Checked out "${title}" to ${patronName}`, type: 'success' })
        setSessionItems((prev) => [{ text: `${title} — Checked out to ${patronName}`, type: 'success' }, ...prev])
      } else {
        setMessage({ text: 'Failed to checkout', type: 'error' })
      }
    } catch {
      setMessage({ text: 'Network error', type: 'error' })
    }
    setPendingISBN(null)
    setProcessing(false)
  }, [pendingISBN, pendingBookTitle, fetchWithAuth])

  return (
    <div>
      <h2>Loan / Return <span style={{ fontSize: '0.5em', fontWeight: 'normal', color: '#8b7355' }}>(Esc to go back)</span></h2>
      <p style={{ color: '#8b7355' }}>Scan a book to automatically check it out or return it.</p>

      <div className="scan-area">
        <IsbnInput onScan={handleScan} disabled={processing || !!pendingISBN} />
        <button
          className="toggle-camera"
          onClick={() => setUseCamera(!useCamera)}
        >
          {useCamera ? 'Use Keyboard' : 'Use Camera'}
        </button>
      </div>

      {message && (
        <div className={`message ${message.type}`}>{message.text}</div>
      )}

      {pendingISBN && (
        <PatronSearch onSelect={handlePatronSelect} />
      )}

      <Scanner active={useCamera && !pendingISBN} onScan={handleScan} />

      <SessionList items={sessionItems} />
    </div>
  )
}
