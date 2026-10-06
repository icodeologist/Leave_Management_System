import { useEffect, useState } from 'react'
import { createLeave, getCurrentUser, getMyLeaves } from '../api'
import Logo from './Logo'

const initialForm = { leave_type: 'CASUAL', day_type: 'FULL_DAY', start_date: '', end_date: '', reason: '' }

const LEAVE_TYPES = { CASUAL: 'Casual', SICK: 'Sick', ANNUAL: 'Annual' }
const DAY_TYPES = { FULL_DAY: 'Full day', HALF_DAY: 'Half day' }
const STATUS_STYLES = {
  PENDING: 'bg-amber-50 text-amber-800 ring-amber-600/20',
  APPROVED: 'bg-green-50 text-green-800 ring-green-600/20',
  REJECTED: 'bg-red-50 text-red-800 ring-red-600/20',
}

const focusRing = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 focus-visible:ring-offset-2'
const primaryButton = `inline-flex items-center justify-center rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60 ${focusRing}`
const secondaryButton = `inline-flex items-center justify-center rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 ${focusRing}`
const fieldClass = 'mt-1.5 block w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-600/30 disabled:bg-slate-100 disabled:text-slate-500'
const labelClass = 'block text-sm font-medium text-slate-700'

// Parse "YYYY-MM-DD" as a local date so it never shifts a day with timezones.
function formatDate(value) {
  if (!value) return ''
  const [year, month, day] = String(value).slice(0, 10).split('-').map(Number)
  if (!year || !month || !day) return value
  return new Date(year, month - 1, day).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

function formatRange(start, end) {
  return !end || start === end ? formatDate(start) : `${formatDate(start)} – ${formatDate(end)}`
}

export default function EmployeeView({ token, user, onUserChange, onLogout }) {
  const [page, setPage] = useState('overview')
  const [leaves, setLeaves] = useState([])
  const [form, setForm] = useState(initialForm)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    getMyLeaves(token)
      .then((response) => setLeaves(response.leaves || []))
      .catch((requestError) => setError(requestError.message))
      .finally(() => setLoading(false))
  }, [token])

  function updateField(event) {
    const { name, value } = event.target
    setForm((current) => {
      const next = { ...current, [name]: value }
      // A half day is a single date; otherwise the end date can never precede the start date.
      if (next.day_type === 'HALF_DAY') {
        next.end_date = next.start_date
      } else if (next.end_date && next.start_date && next.end_date < next.start_date) {
        next.end_date = next.start_date
      }
      return next
    })
  }

  function openRequestPage() {
    setError('')
    setSuccess('')
    setPage('request')
  }

  function closeRequestPage() {
    setError('')
    setPage('overview')
  }

  async function submitLeave(event) {
    event.preventDefault()
    setError('')
    setSuccess('')

    if (form.end_date < form.start_date) {
      setError('End date cannot be before the start date.')
      return
    }

    setSubmitting(true)
    try {
      await createLeave(token, form)
      const [history, currentUser] = await Promise.all([getMyLeaves(token), getCurrentUser(token)])
      setLeaves(history.leaves || [])
      onUserChange(currentUser)
      setForm(initialForm)
      setSuccess('Leave request submitted.')
      setPage('overview')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex h-16 max-w-4xl items-center justify-between px-4 sm:px-6">
          <Logo />
          <div className="flex items-center gap-4">
            <span className="hidden text-sm text-slate-500 sm:inline">{user.email}</span>
            <button type="button" onClick={onLogout} className={secondaryButton}>Log out</button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6">
        {page === 'overview' ? (
          <Overview user={user} leaves={leaves} loading={loading} success={success} error={error} onRequestLeave={openRequestPage} />
        ) : (
          <RequestPage form={form} error={error} submitting={submitting} onChange={updateField} onSubmit={submitLeave} onCancel={closeRequestPage} />
        )}
      </main>
    </div>
  )
}

function Overview({ user, leaves, loading, success, error, onRequestLeave }) {
  return (
    <>
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold tracking-tight">Hello, {user.name}</h1>
        <button type="button" onClick={onRequestLeave} className={primaryButton}>Request leave</button>
      </div>

      {success && <p role="status" className="mt-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">{success}</p>}
      {error && <p role="alert" className="mt-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">{error}</p>}

      <section className="mt-6 rounded-xl border border-slate-200 bg-white p-6">
        <p className="text-sm text-slate-500">Leave balance</p>
        <p className="mt-1 text-4xl font-semibold tabular-nums">
          {user.leave_balance} <span className="text-base font-normal text-slate-500">days remaining</span>
        </p>
      </section>

      <section className="mt-6 rounded-xl border border-slate-200 bg-white">
        <div className="flex items-center justify-between border-b border-slate-200 px-6 py-4">
          <h2 className="text-base font-semibold">Leave history</h2>
          {!loading && <span className="text-sm text-slate-500">{leaves.length} {leaves.length === 1 ? 'request' : 'requests'}</span>}
        </div>

        {loading ? (
          <p className="px-6 py-10 text-center text-sm text-slate-500">Loading leave history…</p>
        ) : leaves.length === 0 ? (
          <div className="px-6 py-10 text-center">
            <p className="text-sm font-medium text-slate-700">No leave requests yet</p>
            <p className="mt-1 text-sm text-slate-500">Select Request leave to submit your first one.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[520px] text-left text-sm">
              <thead className="bg-slate-50 text-slate-500">
                <tr>
                  <th scope="col" className="px-6 py-3 font-medium">Type</th>
                  <th scope="col" className="px-6 py-3 font-medium">Dates</th>
                  <th scope="col" className="px-6 py-3 text-right font-medium">Days</th>
                  <th scope="col" className="px-6 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {leaves.map((leave) => (
                  <tr key={leave.id}>
                    <td className="px-6 py-4">
                      <span className="font-medium text-slate-900">{LEAVE_TYPES[leave.leave_type] || leave.leave_type}</span>
                      <span className="block text-xs text-slate-500">{DAY_TYPES[leave.day_type] || leave.day_type}</span>
                    </td>
                    <td className="px-6 py-4 text-slate-700">{formatRange(leave.start_date, leave.end_date)}</td>
                    <td className="px-6 py-4 text-right tabular-nums text-slate-700">{leave.number_of_days}</td>
                    <td className="px-6 py-4"><Status status={leave.status} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </>
  )
}

function RequestPage({ form, error, submitting, onChange, onSubmit, onCancel }) {
  const isHalfDay = form.day_type === 'HALF_DAY'

  return (
    <section className="mx-auto max-w-xl">
      <button type="button" onClick={onCancel} className={`text-sm font-medium text-blue-600 hover:text-blue-700 ${focusRing} rounded`}>
        ← Back to overview
      </button>

      <h1 className="mt-4 text-2xl font-semibold tracking-tight">Request leave</h1>

      <form onSubmit={onSubmit} className="mt-6 space-y-5 rounded-xl border border-slate-200 bg-white p-6">
        <div className="grid gap-5 sm:grid-cols-2">
          <label className="block">
            <span className={labelClass}>Leave type</span>
            <select name="leave_type" value={form.leave_type} onChange={onChange} className={fieldClass}>
              {Object.entries(LEAVE_TYPES).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
            </select>
          </label>

          <label className="block">
            <span className={labelClass}>Duration</span>
            <select name="day_type" value={form.day_type} onChange={onChange} className={fieldClass}>
              {Object.entries(DAY_TYPES).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
            </select>
          </label>

          <label className="block">
            <span className={labelClass}>{isHalfDay ? 'Date' : 'Start date'}</span>
            <input required type="date" name="start_date" value={form.start_date} onChange={onChange} className={fieldClass} />
          </label>

          {!isHalfDay && (
            <label className="block">
              <span className={labelClass}>End date</span>
              <input required type="date" name="end_date" value={form.end_date} min={form.start_date || undefined} onChange={onChange} className={fieldClass} />
            </label>
          )}
        </div>

        <label className="block">
          <span className={labelClass}>Reason</span>
          <textarea required name="reason" rows={4} value={form.reason} onChange={onChange} className={`${fieldClass} resize-y`} placeholder="Briefly explain why you need leave" />
        </label>

        {error && <p role="alert" className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">{error}</p>}

        <div className="flex justify-end gap-3 border-t border-slate-100 pt-5">
          <button type="button" onClick={onCancel} className={secondaryButton}>Cancel</button>
          <button type="submit" disabled={submitting} className={primaryButton}>{submitting ? 'Submitting…' : 'Submit request'}</button>
        </div>
      </form>
    </section>
  )
}

function Status({ status }) {
  const label = status ? status.charAt(0) + status.slice(1).toLowerCase() : ''
  const style = STATUS_STYLES[status] || 'bg-slate-100 text-slate-700 ring-slate-500/20'
  return <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset ${style}`}>{label}</span>
}
