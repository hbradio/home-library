import { useAuth0 } from '@auth0/auth0-react'
import type { ReactNode } from 'react'

export default function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, loginWithRedirect } = useAuth0()

  if (!isAuthenticated) {
    loginWithRedirect()
    return <div className="loading">Redirecting to login...</div>
  }

  return <>{children}</>
}
