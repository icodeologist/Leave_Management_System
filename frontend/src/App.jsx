import { useEffect, useState } from 'react'
import { getCurrentUser, login, register } from './api'
import AuthScreen from './components/AuthScreen'
import EmployeeView from './components/EmployeeView'
import Logo from './components/Logo'
import AdminView from './components/AdminDashboard'

const TOKEN_KEY = 'leave_management_token'

function App() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(Boolean(token))

  useEffect(() => {
    if (!token) return

    getCurrentUser(token)
      .then(setUser)
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
        setToken(null)
      })
      .finally(() => setLoading(false))
  }, [token])

  async function handleAuth(mode, form) {
    const response = mode === 'login' ? await login(form) : await register(form)
    localStorage.setItem(TOKEN_KEY, response.token)
    setToken(response.token)
    setUser(response.user)
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY)
    setToken(null)
    setUser(null)
  }

  if (loading) {
    return (
      <main className="grid min-h-screen place-items-center bg-slate-50">
        <p className="text-sm text-slate-500">Loading...</p>
      </main>
    )
  }

  if (!token || !user) return <AuthScreen onSubmit={handleAuth} />

  if (user.role === 'EMPLOYEE') {
    return <EmployeeView token={token} user={user} onUserChange={setUser} onLogout={logout} />
  }
  if (user.role === 'ADMIN') {
    return <AdminView token={token} user={user} onLogout={logout} />
  }

  return (
    <main className="grid min-h-screen place-items-center bg-slate-50 px-4">
      <section className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
        <Logo />
        <div className="mt-8 border-t border-slate-100 pt-6">
          <p className="text-sm text-slate-500">Signed in as</p>
          <h1 className="mt-1 text-xl font-semibold text-slate-900">{user.name}</h1>
          <p className="mt-1 text-sm text-slate-500">{user.email}</p>
          <span className="mt-4 inline-flex rounded-md bg-blue-50 px-2.5 py-1 text-xs font-semibold text-blue-700">
            {user.role}
          </span>
        </div>
        <button type="button" onClick={logout} className="secondary-button mt-8 w-full">
          Log out
        </button>
      </section>
    </main>
  )
}

export default App
