const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

function getToken() {
  return localStorage.getItem('token')
}

async function request(path, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`

  const response = await fetch(`${API_URL}${path}`, { ...options, headers })
  const contentType = response.headers.get('content-type') || ''
  const data = contentType.includes('application/json') ? await response.json() : null
  if (!response.ok) {
    throw new Error(data?.error || `HTTP ${response.status}`)
  }
  return data
}

export const api = {
  login: (payload) => request('/auth/login', { method: 'POST', body: JSON.stringify(payload) }),
  register: (payload) => request('/auth/register', { method: 'POST', body: JSON.stringify(payload) }),
  events: (params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request(`/events${qs ? `?${qs}` : ''}`)
  },
  event: (id) => request(`/events/${id}`),
  offers: (eventId) => request(`/events/${eventId}/offers`),
  buy: (offerId) => request(`/offers/${offerId}/buy`, { method: 'POST' }),
  myTickets: () => request('/me/tickets'),
  cancelTicket: (ticketId) => request(`/tickets/${ticketId}/cancel`, { method: 'POST' }),
  transferTicket: (ticketId, email) => request(`/tickets/${ticketId}/transfer`, { method: 'POST', body: JSON.stringify({ email }) }),
  sellerOffers: () => request('/seller/offers'),
  createOffer: (payload) => request('/seller/offers', { method: 'POST', body: JSON.stringify(payload) }),
  createEvent: (payload) => request('/admin/events', { method: 'POST', body: JSON.stringify(payload) }),
  updateEvent: (id, payload) => request(`/admin/events/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  cancelEvent: (id) => request(`/admin/events/${id}`, { method: 'DELETE' }),
  report: (id) => request(`/admin/events/${id}/report`),
}
