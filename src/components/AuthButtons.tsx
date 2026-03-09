import { useAuth0 } from '@auth0/auth0-react'

export default function AuthButtons() {
  const { isAuthenticated, loginWithRedirect, logout, user } = useAuth0()

  if (isAuthenticated) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.8em' }}>
        <span style={{ fontSize: '0.9em' }}>{user?.email}</span>
        <button onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}>
          Log out
        </button>
      </div>
    )
  }

  return <button onClick={() => loginWithRedirect()}>Log in</button>
}
