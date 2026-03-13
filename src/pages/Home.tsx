import { useNavigate } from 'react-router-dom'
import { useAuth0 } from '@auth0/auth0-react'
import { useRef, useEffect } from 'react'

export default function Home() {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth0()
  const buttonsRef = useRef<(HTMLButtonElement | null)[]>([])

  useEffect(() => {
    if (isAuthenticated) {
      buttonsRef.current[0]?.focus()
    }
  }, [isAuthenticated])

  const handleKeyDown = (e: React.KeyboardEvent, index: number) => {
    const count = buttonsRef.current.length
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      buttonsRef.current[(index + 1) % count]?.focus()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      buttonsRef.current[(index - 1 + count) % count]?.focus()
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="home-buttons">
        <p style={{ fontSize: '1.2em', textAlign: 'center' }}>
          Manage your personal book collection.<br />
          Log in to get started.
        </p>
      </div>
    )
  }

  const buttons = [
    { label: 'Add Book', path: '/add' },
    { label: 'Browse', path: '/browse' },
    { label: 'Loan / Return', path: '/loan' },
    { label: 'Patrons', path: '/patrons' },
  ]

  const secondaryButtons = [
    { label: 'Sharing', path: '/sharing' },
    { label: 'Shared Libraries', path: '/shared' },
  ]

  return (
    <div className="home-buttons">
      {buttons.map((btn, i) => (
        <button
          key={btn.path}
          className="home-button"
          ref={(el) => { buttonsRef.current[i] = el }}
          onClick={() => navigate(btn.path)}
          onKeyDown={(e) => handleKeyDown(e, i)}
        >
          {btn.label}
        </button>
      ))}
      <div style={{ display: 'flex', gap: '0.5em', justifyContent: 'center', marginTop: '0.5em' }}>
        {secondaryButtons.map((btn, i) => (
          <button
            key={btn.path}
            ref={(el) => { buttonsRef.current[buttons.length + i] = el }}
            onClick={() => navigate(btn.path)}
            onKeyDown={(e) => handleKeyDown(e, buttons.length + i)}
            style={{ fontSize: '0.8em', padding: '0.4em 1em', color: '#8b7355', background: 'transparent', border: '1px solid #d4c9b8', cursor: 'pointer' }}
          >
            {btn.label}
          </button>
        ))}
      </div>
    </div>
  )
}
