import { Routes, Route, useNavigate } from 'react-router-dom'
import { useAuth0 } from '@auth0/auth0-react'
import { useEffect } from 'react'
import Home from './pages/Home'
import AddBook from './pages/AddBook'
import Browse from './pages/Browse'
import BookDetail from './pages/BookDetail'
import LoanReturn from './pages/LoanReturn'
import Patrons from './pages/Patrons'
import PatronDetail from './pages/PatronDetail'
import AuthButtons from './components/AuthButtons'
import ProtectedRoute from './components/ProtectedRoute'

function App() {
  const { isLoading } = useAuth0()
  const navigate = useNavigate()

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        navigate('/')
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [navigate])

  if (isLoading) {
    return <div className="app"><div className="loading">Loading...</div></div>
  }

  return (
    <div className="app">
      <div className="header">
        <h1><a href="/" style={{ color: 'inherit', textDecoration: 'none' }}>Home Library</a></h1>
        <AuthButtons />
      </div>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/add" element={<ProtectedRoute><AddBook /></ProtectedRoute>} />
        <Route path="/browse" element={<ProtectedRoute><Browse /></ProtectedRoute>} />
        <Route path="/book/:id" element={<ProtectedRoute><BookDetail /></ProtectedRoute>} />
        <Route path="/loan" element={<ProtectedRoute><LoanReturn /></ProtectedRoute>} />
        <Route path="/patrons" element={<ProtectedRoute><Patrons /></ProtectedRoute>} />
        <Route path="/patron/:id" element={<ProtectedRoute><PatronDetail /></ProtectedRoute>} />
      </Routes>
    </div>
  )
}

export default App
