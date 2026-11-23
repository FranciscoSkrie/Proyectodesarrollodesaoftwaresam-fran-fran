const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

function getToken() {
  return localStorage.getItem('token')
}

function collectTelemetry() {
  const nav = window.navigator || {}
  const screenData = window.screen || {}
  return {
    user_agent: nav.userAgent || '',
    language: nav.language || '',
    languages: Array.isArray(nav.languages) ? nav.languages : [],
    platform: nav.platform || '',
    cookie_enabled: Boolean(nav.cookieEnabled),
    hardware_concurrency: nav.hardwareConcurrency || null,
    device_memory: nav.deviceMemory || null,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || '',
    timezone_offset_minutes: new Date().getTimezoneOffset(),
    screen: {
      width: screenData.width || null,
      height: screenData.height || null,
      color_depth: screenData.colorDepth || null,
      pixel_depth: screenData.pixelDepth || null
    },
    viewport: {
      width: window.innerWidth || null,
      height: window.innerHeight || null,
      device_pixel_ratio: window.devicePixelRatio || null
    },
    referrer: document.referrer || '',
    current_url: window.location.href
  }
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
  changePassword: (payload) => request('/me/password', { method: 'POST', body: JSON.stringify(payload) }),
  events: (params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request(`/events${qs ? `?${qs}` : ''}`)
  },
  event: (id) => request(`/events/${id}`),
  offers: (eventId) => request(`/events/${eventId}/offers`),
  buy: (offerId) => request(`/offers/${offerId}/buy`, { method: 'POST' }),
  buyEvent: (eventId) => request(`/events/${eventId}/buy`, { method: 'POST' }),
  myTickets: () => request('/me/tickets'),
  cancelTicket: (ticketId) => request(`/tickets/${ticketId}/cancel`, { method: 'POST' }),
  transferTicket: (ticketId, userId) => request(`/tickets/${ticketId}/transfer`, { method: 'POST', body: JSON.stringify({ user_id: Number(userId) }) }),
  myOffers: () => request('/me/offers'),
  createOffer: (payload) => request('/me/offers', { method: 'POST', body: JSON.stringify({ ...payload, telemetry: collectTelemetry() }) }),
  requestExchange: (offerId, payload) => request(`/offers/${offerId}/exchange`, { method: 'POST', body: JSON.stringify(payload) }),
  exchangeInbox: () => request('/me/exchanges/inbox'),
  exchangeSent: () => request('/me/exchanges/sent'),
  acceptExchange: (id) => request(`/exchanges/${id}/accept`, { method: 'POST' }),
  rejectExchange: (id) => request(`/exchanges/${id}/reject`, { method: 'POST' }),
  createEvent: (payload) => request('/admin/events', { method: 'POST', body: JSON.stringify(payload) }),
  updateEvent: (id, payload) => request(`/admin/events/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  cancelEvent: (id) => request(`/admin/events/${id}`, { method: 'DELETE' }),
  report: (id) => request(`/admin/events/${id}/report`),
  securityOffers: () => request('/admin/security/offers'),
  clearAlerts: () => request('/admin/security/alerts/clear', { method: 'POST' })
}
