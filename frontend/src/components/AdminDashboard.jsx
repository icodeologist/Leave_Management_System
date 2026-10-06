import { useCallback, useEffect, useMemo, useState } from 'react'
// ASSUMPTION: adjust these two to match your api.js (see notes below the file).
import { getAllLeaves, updateLeaveStatus } from '../api'
import Logo from './Logo'

const LEAVE_TYPES = { CASUAL: 'Casual', SICK: 'Sick', ANNUAL: 'Annual' }
const DAY_TYPES = { FULL_DAY: 'Full day', HALF_DAY: 'Half day' }
const STATUS_STYLES = {
  PENDING: 'bg-amber-50 text-amber-800 ring-amber-600/20',
  APPROVED: 'bg-green-50 text-green-800 ring-green-600/20',
  REJECTED: 'bg-red-50 text-red-800 ring-red-600/20',
}
const FILTERS = [
  { value: 'ALL', label: 'All' },
  { value: 'PENDING', label: 'Pending' },
  { value: 'APPROVED', label: 'Approved' },
  { value: 'REJECTED', label: 'Rejected' },
]

const focusRing = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 focus-visible:ring-offset-2'
const secondaryButton = `inline-flex items-center justify-center rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60 ${focusRing}`
const approveButton = `inline-flex items-center justify-center rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60 ${focusRing}`
const rejectButton = `inline-flex items-center justify-center rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60 ${focusRing}`

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

function formatSubmitted(value) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString('en-GB', { day: 'numeric', month: 'short', hour: 'numeric', minute: '2-digit', hour12: true })
}

function isNew(value) {
  const time = new Date(value).getTime()
  return !Number.isNaN(time) && Date.now() - time < 24 * 60 * 60 * 1000
}

// Newest first: by created_at when present, otherwise by id.
function sortRecent(leaves) {
  return [...leaves].sort((a, b) => {
    const diff = new Date(b.created_at || 0) - new Date(a.created_at || 0)
    return diff !== 0 ? diff : (b.id || 0) - (a.id || 0)
  })
}

function employeeOf(leave) {
  const person = leave.user || leave.employee || {}
  return {
    name: person.name || leave.user_name || leave.employee_name || `User #${leave.user_id ?? '?'}`,
    email: person.email || leave.user_email || leave.employee_email || '',
  }
}

export default function AdminView({ token, user, onLogout }) {
  const [leaves, setLeaves] = useState([])
  const [filter, setFilter] = useState('ALL')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [actingId, setActingId] = useState(null)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')

  const load = useCallback(async () => {
    try {
      const response = await getAllLeaves(token)
      setLeaves(sortRecent(response.leaves || []))
      setError('')
    } catch (requestError) {
      setError(requestError.message)
    }
  }, [token])

  useEffect(() => {
    load().finally(() => setLoading(false))
  }, [load])

  async function refresh() {
    setRefreshing(true)
    setNotice('')
    await load()
    setRefreshing(false)
  }

  async function decide(leave, status) {
    setActingId(leave.id)
    setError('')
    setNotice('')
    try {
      await updateLeaveStatus(token, leave.id, status)
      await load()
      setNotice(`Request from ${employeeOf(leave).name} ${status === 'APPROVED' ? 'approved' : 'rejected'}.`)
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setActingId(null)
    }
  }

  const counts = useMemo(() => {
    const result = { ALL: leaves.length, PENDING: 0, APPROVED: 0, REJECTED: 0 }
    leaves.forEach((leave) => { if (result[leave.status] !== undefined) result[leave.status] += 1 })
    return result
  }, [leaves])

  const visible = filter === 'ALL' ? leaves : leaves.filter((leave) => leave.status === filter)

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
          <Logo />
          <div className="flex items-center gap-4">
            <span className="hidden text-sm text-slate-500 sm:inline">{user.email}</span>
            <button type="button" onClick={onLogout} className={secondaryButton}>Log out</button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <h1 className="text-2xl font-semibold tracking-tight">Admin dashboard</h1>
          <button type="button" onClick={refresh} disabled={refreshing || loading} className={secondaryButton}>
            {refreshing ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>

        {notice && <p role="status" className="mt-6 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">{notice}</p>}
        {error && <p role="alert" className="mt-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">{error}</p>}

        <section className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <Stat label="Total requests" value={counts.ALL} loading={loading} />
          <Stat label="Pending" value={counts.PENDING} loading={loading} />
          <Stat label="Approved" value={counts.APPROVED} loading={loading} />
          <Stat label="Rejected" value={counts.REJECTED} loading={loading} />
        </section>

        <section className="mt-6 rounded-xl border border-slate-200 bg-white">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-6 py-4">
            <h2 className="text-base font-semibold">Recent leave requests</h2>
            <div role="tablist" aria-label="Filter by status" className="flex gap-1 rounded-lg bg-slate-100 p-1">
              {FILTERS.map((item) => (
                <button
                  key={item.value}
                  type="button"
                  role="tab"
                  aria-selected={filter === item.value}
                  onClick={() => setFilter(item.value)}
                  className={`rounded-md px-3 py-1 text-sm font-medium ${focusRing} ${filter === item.value ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
                >
                  {item.label} <span className="tabular-nums text-slate-400">{counts[item.value]}</span>
                </button>
              ))}
            </div>
          </div>

          {loading ? (
            <p className="px-6 py-10 text-center text-sm text-slate-500">Loading leave requests…</p>
          ) : visible.length === 0 ? (
            <div className="px-6 py-10 text-center">
              <p className="text-sm font-medium text-slate-700">{filter === 'ALL' ? 'No leave requests yet' : `No ${filter.toLowerCase()} requests`}</p>
              <p className="mt-1 text-sm text-slate-500">New requests from employees will appear here.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[880px] text-left text-sm">
                <thead className="bg-slate-50 text-slate-500">
                  <tr>
                    <th scope="col" className="px-6 py-3 font-medium">Employee</th>
                    <th scope="col" className="px-6 py-3 font-medium">Type</th>
                    <th scope="col" className="px-6 py-3 font-medium">Dates</th>
                    <th scope="col" className="px-6 py-3 text-right font-medium">Days</th>
                    <th scope="col" className="px-6 py-3 font-medium">Submitted</th>
                    <th scope="col" className="px-6 py-3 font-medium">Status</th>
                    <th scope="col" className="px-6 py-3 text-right font-medium">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {visible.map((leave) => {
                    const employee = employeeOf(leave)
                    const busy = actingId === leave.id
                    return (
                      <tr key={leave.id} className="align-top">
                        <td className="px-6 py-4">
                          <span className="font-medium text-slate-900">{employee.name}</span>
                          {employee.email && <span className="block text-xs text-slate-500">{employee.email}</span>}
                          {leave.reason && <span title={leave.reason} className="mt-1 block max-w-xs truncate text-xs text-slate-500">{leave.reason}</span>}
                        </td>
                        <td className="px-6 py-4">
                          <span className="font-medium text-slate-900">{LEAVE_TYPES[leave.leave_type] || leave.leave_type}</span>
                          <span className="block text-xs text-slate-500">{DAY_TYPES[leave.day_type] || leave.day_type}</span>
                        </td>
                        <td className="px-6 py-4 text-slate-700">{formatRange(leave.start_date, leave.end_date)}</td>
                        <td className="px-6 py-4 text-right tabular-nums text-slate-700">{leave.number_of_days}</td>
                        <td className="px-6 py-4 text-slate-700">
                          {formatSubmitted(leave.created_at)}
                          {isNew(leave.created_at) && <span className="ml-2 rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">New</span>}
                        </td>
                        <td className="px-6 py-4"><Status status={leave.status} /></td>
                        <td className="px-6 py-4 text-right">
                          {leave.status === 'PENDING' ? (
                            <div className="flex justify-end gap-2">
                              <button type="button" disabled={busy} onClick={() => decide(leave, 'APPROVED')} className={approveButton}>Approve</button>
                              <button type="button" disabled={busy} onClick={() => decide(leave, 'REJECTED')} className={rejectButton}>Reject</button>
                            </div>
                          ) : (
                            <span className="text-slate-300">—</span>
                          )}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </main>
    </div>
  )
}

function Stat({ label, value, loading }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="mt-1 text-3xl font-semibold tabular-nums">{loading ? '–' : value}</p>
    </div>
  )
}

function Status({ status }) {
  const label = status ? status.charAt(0) + status.slice(1).toLowerCase() : ''
  const style = STATUS_STYLES[status] || 'bg-slate-100 text-slate-700 ring-slate-500/20'
  return <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset ${style}`}>{label}</span>
}
