import { useState } from 'react'
import Logo from './Logo'

const initialForm = {
  name: '',
  email: '',
  password: '',
  role: 'EMPLOYEE',
}

export default function AuthScreen({ onSubmit }) {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState(initialForm)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function changeMode(nextMode) {
    setMode(nextMode)
    setError('')
  }

  function updateField(event) {
    setForm((current) => ({ ...current, [event.target.name]: event.target.value }))
  }

  async function handleSubmit(event) {
    event.preventDefault()
    setError('')
    setLoading(true)

    try {
      const payload = mode === 'login'
        ? { email: form.email, password: form.password }
        : form
      await onSubmit(mode, payload)
    } catch (submitError) {
      setError(submitError.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4 py-10">
      <section className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-6 shadow-sm sm:p-8">
        <Logo />

        <div className="mt-8 grid grid-cols-2 rounded-lg bg-slate-100 p-1">
          <Tab active={mode === 'login'} onClick={() => changeMode('login')}>Login</Tab>
          <Tab active={mode === 'register'} onClick={() => changeMode('register')}>Register</Tab>
        </div>

        <div className="mt-7">
          <h1 className="text-2xl font-semibold text-slate-900">
            {mode === 'login' ? 'Welcome back' : 'Create an account'}
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            {mode === 'login' ? 'Enter your account details to continue.' : 'Enter your details to register.'}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          {mode === 'register' && (
            <>
              <Field label="Full name" name="name" value={form.name} onChange={updateField} placeholder="Your name" />
              <label className="block">
                <span className="label">Role</span>
                <select name="role" value={form.role} onChange={updateField} className="field">
                  <option value="EMPLOYEE">Employee</option>
                  <option value="ADMIN">Admin</option>
                </select>
              </label>
            </>
          )}

          <Field label="Email" name="email" type="email" value={form.email} onChange={updateField} placeholder="you@example.com" />
          <Field label="Password" name="password" type="password" value={form.password} onChange={updateField} placeholder="Minimum 8 characters" minLength={8} />

          {error && <p className="rounded-lg bg-red-50 px-3 py-2.5 text-sm text-red-700">{error}</p>}

          <button type="submit" disabled={loading} className="primary-button w-full">
            {loading ? 'Please wait...' : mode === 'login' ? 'Login' : 'Register'}
          </button>
        </form>
      </section>
    </main>
  )
}

function Tab({ active, onClick, children }) {
  return (
    <button type="button" onClick={onClick} className={`rounded-md px-3 py-2 text-sm font-medium ${active ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500'}`}>
      {children}
    </button>
  )
}

function Field({ label, ...props }) {
  return (
    <label className="block">
      <span className="label">{label}</span>
      <input required className="field" {...props} />
    </label>
  )
}
