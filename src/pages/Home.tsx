import { useNavigate } from 'react-router-dom'
import { useAuth0 } from '@auth0/auth0-react'

export default function Home() {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth0()

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

  return (
    <div className="home-buttons">
      <button className="home-button" onClick={() => navigate('/add')} tabIndex={1}>
        Add Book
      </button>
      <button className="home-button" onClick={() => navigate('/browse')} tabIndex={2}>
        Browse
      </button>
      <button className="home-button" onClick={() => navigate('/loan')} tabIndex={3}>
        Loan / Return
      </button>
      <button className="home-button" onClick={() => navigate('/patrons')} tabIndex={4}>
        Patrons
      </button>
    </div>
  )
}
