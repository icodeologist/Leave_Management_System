const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

async function request(path, { token, ...options } = {}) {
  const headers = { ...options.headers }
  if (options.body) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`

  const response = await fetch(`${API_URL}${path}`, { ...options, headers })
  const text = await response.text()
  const data = text ? JSON.parse(text) : null

  if (!response.ok) throw new Error(data?.error || 'Something went wrong. Please try again.')
  return data
}

export function login(credentials) {
  return request('/api/auth/login', { method: 'POST', body: JSON.stringify(credentials) })
}

export function register(details) {
  return request('/api/auth/register', { method: 'POST', body: JSON.stringify(details) })
}

export function getCurrentUser(token) {
  return request('/api/me', { token })
}

export function createLeave(token, leave) {
  return request('/api/leaves', {
    token,
    method: 'POST',
    body: JSON.stringify(leave),
  })
}

export function getMyLeaves(token) {
  return request('/api/leaves/my', { token })
}

export function getAllLeaves(token) {
  return request('/api/admin/leaves', { token })
}

export function updateLeaveStatus(token, leaveID, status) {
  const action = status === 'APPROVED' ? 'approve' : 'reject'
  return request(`/api/admin/leaves/${leaveID}/${action}`, {
    token,
    method: 'PATCH',
  })
}
