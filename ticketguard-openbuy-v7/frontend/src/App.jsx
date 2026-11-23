import * as React from 'react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { api } from './api.js'
const emptyEvent = {
  title: '',
  description: '',
  category: '',
  location: '',
  starts_at: '',
  duration_minutes: 120,
  capacity: 100,
  entry_price: 15000,
  image_url: ''
}

const SESSION_VERSION = 'ticketguard-openbuy-v7'

function clearStoredSession() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  localStorage.removeItem('session_version')
}

function loadStoredSession() {
  const version = localStorage.getItem('session_version')
  const token = localStorage.getItem('token')
  const user = localStorage.getItem('user')

  if (version !== SESSION_VERSION || !token || !user) {
    clearStoredSession()
    return null
  }

  try {
    return { token, user: normalizeUser(JSON.parse(user)) }
  } catch {
    clearStoredSession()
    return null
  }
}

function normalizeUser(user) {
  if (!user) return user
  return { ...user, role: user.role === 'admin' ? 'admin' : 'usuario' }
}

function App() {
  const [session, setSession] = useState(loadStoredSession)
  const [view, setView] = useState('home')
  const [events, setEvents] = useState([])
  const [selectedEvent, setSelectedEvent] = useState(null)
  const [offers, setOffers] = useState([])
  const [tickets, setTickets] = useState([])
  const [userOffers, setUserOffers] = useState([])
  const [exchangeInbox, setExchangeInbox] = useState([])
  const [exchangeSent, setExchangeSent] = useState([])
  const [exchangeOffer, setExchangeOffer] = useState(null)
  const [report, setReport] = useState(null)
  const [securityAlerts, setSecurityAlerts] = useState([])
  const [selectedEvidence, setSelectedEvidence] = useState(null)
  const [message, setMessage] = useState('')
  const [filters, setFilters] = useState({ q: '', category: '' })
  const syncingHistory = useRef(false)

  const role = session?.user?.role || 'anon'
  const isUser = role === 'usuario'
  const isAdmin = role === 'admin'

  useEffect(() => { loadEvents() }, [])

  useEffect(() => {
    if (!message) return
    const id = setTimeout(() => setMessage(''), 3500)
    return () => clearTimeout(id)
  }, [message])

  useEffect(() => {
    window.history.replaceState({ ticketguard: true, view: 'home' }, '', '#home')
    const onPopState = (event) => {
      syncingHistory.current = true
      const nextView = event.state?.view || 'home'
      setView(nextView)
    }
    window.addEventListener('popstate', onPopState)
    return () => window.removeEventListener('popstate', onPopState)
  }, [])

  useEffect(() => {
    if (syncingHistory.current) {
      syncingHistory.current = false
      return
    }
    const state = window.history.state
    if (!state?.ticketguard || state.view !== view) {
      window.history.pushState({ ticketguard: true, view }, '', `#${view}`)
    }
  }, [view])

  async function loadEvents(nextFilters = filters) {
    try {
      setEvents(await api.events(nextFilters))
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function loadOwnedTicketsOnly() {
    if (!isUser) return []
    const data = await api.myTickets()
    setTickets(data)
    return data
  }

  function saveSession(resp) {
    const normalized = { ...resp, user: normalizeUser(resp.user) }
    localStorage.setItem('session_version', SESSION_VERSION)
    localStorage.setItem('token', normalized.token)
    localStorage.setItem('user', JSON.stringify(normalized.user))
    setSession(normalized)
    setView('account')
    setMessage('Sesión iniciada correctamente')
  }

  function logout() {
    clearStoredSession()
    setSession(null)
    setTickets([])
    setUserOffers([])
    setExchangeInbox([])
    setExchangeSent([])
    setSecurityAlerts([])
    setSelectedEvidence(null)
    setExchangeOffer(null)
    setView('home')
    setMessage('Sesión cerrada')
  }

  async function openEvent(event) {
    try {
      setSelectedEvent(event)
      setOffers(await api.offers(event.id))
      setView('detail')
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function loadTickets() {
    if (!isUser) {
      setMessage('Para ver entradas tenés que iniciar sesión como usuario')
      return
    }
    try {
      await loadOwnedTicketsOnly()
      await loadExchangeBox()
      setView('tickets')
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function loadExchangeBox() {
    if (!isUser) return
    const [inbox, sent] = await Promise.all([api.exchangeInbox(), api.exchangeSent()])
    setExchangeInbox(inbox)
    setExchangeSent(sent)
  }

  async function loadUserOffers() {
    if (!isUser) {
      setMessage('Para crear publicaciones tenés que iniciar sesión como usuario')
      return
    }
    try {
      const [offersData, ticketsData] = await Promise.all([api.myOffers(), api.myTickets()])
      setUserOffers(offersData)
      setTickets(ticketsData)
      setView('publish')
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function loadReport(eventId) {
    try {
      setReport(await api.report(eventId))
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function loadSecurityAlerts() {
    try {
      setSecurityAlerts(await api.securityOffers())
    } catch (err) {
      setMessage(err.message)
    }
  }

  function goPublish() {
    if (!session) {
      setView('login')
      setMessage('Iniciá sesión como usuario para crear publicaciones')
      return
    }
    if (!isUser) {
      setMessage('Solo una cuenta de usuario puede crear publicaciones')
      return
    }
    loadUserOffers()
  }

  function goAdmin() {
    if (!isAdmin) {
      setMessage('Solo administración puede ver el panel de gestión')
      return
    }
    loadSecurityAlerts()
    setView('admin')
  }

  function goAccount() {
    if (!session) {
      setView('login')
      return
    }
    setView('account')
  }

  function requireLogin() {
    setView('login')
    setMessage('Iniciá sesión con una cuenta de usuario para operar sobre entradas')
  }

  async function openExchangeModal(offer) {
    if (!session) {
      requireLogin()
      return
    }
    if (!isUser) {
      setMessage('Solo usuarios pueden solicitar intercambios')
      return
    }
    try {
      await loadOwnedTicketsOnly()
      setExchangeOffer(offer)
    } catch (err) {
      setMessage(err.message)
    }
  }

  return (
    <div className="market-shell">
      <Sidebar
        session={session}
        view={view}
        filters={filters}
        setFilters={setFilters}
        onSearch={loadEvents}
        setView={setView}
        onAccount={goAccount}
        onTickets={loadTickets}
        onPublish={goPublish}
        onAdmin={goAdmin}
        isUser={isUser}
        isAdmin={isAdmin}
        onLogout={logout}
      />

      <main className="content">
        <div className="content-top">
          <h2>{titleFor(view)}</h2>
        </div>

        {message && <div className="message" onClick={() => setMessage('')}>{message}</div>}

        {view === 'home' && <Home events={events} onOpen={openEvent} />}
        {view === 'login' && <Login onLogin={saveSession} setMessage={setMessage} />}
        {view === 'account' && <Account session={session} onTickets={loadTickets} onPublish={goPublish} onAdmin={goAdmin} onLogout={logout} isUser={isUser} isAdmin={isAdmin} setMessage={setMessage} />}
        {view === 'detail' && <Detail event={selectedEvent} offers={offers} session={session} isUser={isUser} onRequireLogin={requireLogin} onBuyEvent={async (id) => {
          try {
            await api.buyEvent(id)
            setMessage('Entrada comprada correctamente. El saldo no se modifica en este TP.')
            await loadTickets()
          } catch (err) {
            setMessage(err.message)
          }
        }} onExchange={openExchangeModal} />}
        {view === 'tickets' && <Tickets tickets={tickets} inbox={exchangeInbox} sent={exchangeSent} reload={loadTickets} setMessage={setMessage} onCancel={async (id) => {
          try {
            await api.cancelTicket(id)
            setMessage('Entrada cancelada')
            await loadTickets()
          } catch (err) {
            setMessage(err.message)
          }
        }} onTransfer={async (id, userId) => {
          try {
            await api.transferTicket(id, userId)
            setMessage('Entrada transferida')
            await loadTickets()
          } catch (err) {
            setMessage(err.message)
          }
        }} />}
        {view === 'publish' && <Publish events={events} offers={userOffers} tickets={tickets} reload={loadUserOffers} setMessage={setMessage} />}
        {view === 'admin' && <Admin events={events} reloadEvents={loadEvents} report={report} loadReport={loadReport} alerts={securityAlerts} reloadAlerts={loadSecurityAlerts} clearAlerts={async () => {
          try {
            await api.clearAlerts()
            setMessage('Alertas limpiadas')
            await loadSecurityAlerts()
          } catch (err) {
            setMessage(err.message)
          }
        }} onMoreInfo={setSelectedEvidence} setMessage={setMessage} />}
      </main>

      {exchangeOffer && <ExchangeModal offer={exchangeOffer} tickets={tickets} onClose={() => setExchangeOffer(null)} onSubmit={async (payload) => {
        try {
          await api.requestExchange(exchangeOffer.id, payload)
          setMessage('Solicitud de intercambio enviada')
          setExchangeOffer(null)
          await loadExchangeBox()
        } catch (err) {
          setMessage(err.message)
        }
      }} />}

      {selectedEvidence && <EvidenceModal evidence={selectedEvidence} onClose={() => setSelectedEvidence(null)} />}
    </div>
  )
}

function titleFor(view) {
  const titles = {
    home: 'Foros de eventos destacados',
    login: 'Tu cuenta',
    account: 'Tu cuenta',
    detail: 'Publicaciones del evento',
    tickets: 'Mis entradas e intercambios',
    publish: 'Crear publicación',
    admin: 'Panel de gestión y alertas'
  }
  return titles[view] || 'Marketplace'
}

function Sidebar({ session, view, filters, setFilters, onSearch, setView, onAccount, onTickets, onPublish, onAdmin, isUser, isAdmin, onLogout }) {
  function submit(e) {
    e.preventDefault()
    onSearch()
    setView('home')
  }

  function setCategory(category) {
    const next = { ...filters, category }
    setFilters(next)
    onSearch(next)
    setView('home')
  }

  return (
    <aside className="sidebar">
      <div className="brand-row">
        <h1>Marketplace</h1>
        <button className="round-btn" title="Configuración de cuenta" aria-label="Abrir cuenta o iniciar sesión" onClick={onAccount}>⚙</button>
      </div>

      <form className="search-box" onSubmit={submit}>
        <span>⌕</span>
        <input placeholder="Buscar en Marketplace" value={filters.q} onChange={e => setFilters({ ...filters, q: e.target.value })} />
      </form>

      <div className="side-menu">
        <button className={view === 'home' ? 'side-active' : ''} onClick={() => { setView('home'); onSearch() }}>🏪 Explorar todo</button>
        <button className={view === 'account' || view === 'login' ? 'side-active' : ''} onClick={onAccount}>👤 Tu cuenta</button>
        {isUser && <button className="create-btn" onClick={onPublish}>＋ Crear publicación</button>}
        {isUser && <button className={view === 'tickets' ? 'side-active' : ''} onClick={onTickets}>🎟️ Mis entradas</button>}
        {isAdmin && <button className={view === 'admin' ? 'side-active' : ''} onClick={onAdmin}>🛡️ Gestión y alertas</button>}
        {session && <button onClick={onLogout}>↩ Salir</button>}
      </div>

      <hr />

      <section className="sidebar-section">
        <h3>Filtros</h3>
        <input placeholder="Categoría" value={filters.category} onChange={e => setFilters({ ...filters, category: e.target.value })} />
        <button onClick={() => { onSearch(); setView('home') }}>Aplicar filtros</button>
      </section>

      <section className="sidebar-section categories">
        <h3>Categorías</h3>
        <button onClick={() => setCategory('Tecnología')}>💻 Tecnología</button>
        <button onClick={() => setCategory('Música')}>🎵 Música</button>
        <button onClick={() => setCategory('Deportes')}>⚽ Deportes</button>
        <button onClick={() => setCategory('Teatro')}>🎭 Teatro</button>
        <button onClick={() => setCategory('Gaming')}>🎮 Gaming</button>
        <button onClick={() => setCategory('Conferencias')}>🎤 Conferencias</button>
      </section>
    </aside>
  )
}

function Home({ events, onOpen }) {
  return (
    <section className="market-grid">
      {events.length === 0 && <div className="empty-state">No se encontraron foros con esos filtros.</div>}
      {events.map(event => <article className="listing-card" key={event.id} onClick={() => onOpen(event)}>
        <div className="thumb">
          {event.image_url ? <img src={event.image_url} alt={event.title} /> : <span>{event.category || 'Foro'}</span>}
        </div>
        <div className="listing-body">
          <strong>{event.title}</strong>
          <p>{event.location}</p>
          <small>{new Date(event.starts_at).toLocaleString()} · {event.category}</small>
          <small>Valor entrada general: ${formatMoney(event.entry_price)}</small>
        </div>
      </article>)}
    </section>
  )
}

function Login({ onLogin, setMessage }) {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({ name: '', email: 'usuario@ticketguard.test', password: 'Usuario123!' })
  async function submit(e) {
    e.preventDefault()
    try {
      const resp = mode === 'login' ? await api.login(form) : await api.register(form)
      onLogin(resp)
    } catch (err) {
      setMessage(err.message)
    }
  }
  return <section className="panel narrow-panel">
    <h2>{mode === 'login' ? 'Iniciar sesión' : 'Crear cuenta'}</h2>
    <p className="hint">Sin iniciar sesión solo se puede explorar el catálogo público. Publicar, comprar, intercambiar, transferir y gestionar alertas queda bloqueado hasta autenticar la cuenta correcta.</p>
    <form onSubmit={submit}>
      {mode === 'register' && <input placeholder="Nombre" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />}
      <input placeholder="Email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} />
      <input placeholder="Password" type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} />
      <button>{mode === 'login' ? 'Entrar' : 'Registrarme'}</button>
    </form>
    <button className="text-btn" onClick={() => setMode(mode === 'login' ? 'register' : 'login')}>{mode === 'login' ? 'Crear una cuenta' : 'Ya tengo cuenta'}</button>
    <div className="demo-box">
      <p><strong>Demo</strong></p>
      <p>admin@ticketguard.test / Admin123!</p>
      <p>usuario@ticketguard.test / Usuario123!</p>
      <p>intercambio@ticketguard.test / Usuario123!</p>
    </div>
  </section>
}

function Account({ session, onTickets, onPublish, onAdmin, onLogout, isUser, isAdmin, setMessage }) {
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  if (!session) return null
  return <section className="account-grid">
    <div className="panel account-main">
      <h2>Cuenta conectada</h2>
      <p className="hint">Desde acá se accede solo a las acciones habilitadas para esta sesión.</p>
      <div className="account-summary"><span>ID de usuario</span><strong>{session.user.id}</strong></div>
      <div className="account-summary"><span>Correo</span><strong>{session.user.email}</strong></div>
      <div className="account-actions">
        {isUser && <button onClick={onTickets}>Ver mis entradas</button>}
        {isUser && <button onClick={onPublish}>Crear publicación</button>}
        {isAdmin && <button onClick={onAdmin}>Ver gestión y alertas</button>}
        <button onClick={() => setShowPasswordModal(true)}>Actualizar contraseña</button>
        <button className="secondary-btn" onClick={onLogout}>Salir</button>
      </div>
    </div>
    <div className="panel account-help">
      <h2>Accesos por sesión</h2>
      <p>La sesión sin login solo puede explorar foros. Las publicaciones, entradas, transferencias, intercambios y alertas se validan contra el backend.</p>
      <p>Cada publicación funciona como parte del honeypot: el link externo se analiza y queda registrado como evidencia administrativa si resulta sospechoso.</p>
    </div>
    {showPasswordModal && <PasswordModal onClose={() => setShowPasswordModal(false)} setMessage={setMessage} />}
  </section>
}

function PasswordModal({ onClose, setMessage }) {
  const [form, setForm] = useState({ current_password: '', new_password: '' })
  async function submit(e) {
    e.preventDefault()
    try {
      await api.changePassword(form)
      setMessage('Contraseña actualizada')
      onClose()
    } catch (err) {
      setMessage(err.message)
    }
  }
  return <Modal title="Actualizar contraseña" onClose={onClose}>
    <form className="password-form" onSubmit={submit}>
      <input type="password" placeholder="Contraseña actual" value={form.current_password} onChange={e => setForm({ ...form, current_password: e.target.value })} />
      <input type="password" placeholder="Nueva contraseña" value={form.new_password} onChange={e => setForm({ ...form, new_password: e.target.value })} />
      <div className="modal-actions"><button>Guardar contraseña</button><button type="button" className="secondary-btn" onClick={onClose}>Cancelar</button></div>
    </form>
  </Modal>
}

function Detail({ event, offers, session, isUser, onRequireLogin, onBuyEvent, onExchange }) {
  if (!event) return null
  return <div className="detail-layout">
    <section className="panel detail-panel">
      {event.image_url && <img className="detail-image" src={event.image_url} alt={event.title} />}
      <h2>{event.title}</h2>
      <p>{event.description}</p>
      <div className="meta-list two-cols">
        <span><strong>Lugar</strong>{event.location}</span>
        <span><strong>Fecha</strong>{new Date(event.starts_at).toLocaleString()}</span>
        <span><strong>Valor entrada</strong>${formatMoney(event.entry_price)}</span>
      </div>
      <button className="buy-event-btn" onClick={() => session && isUser ? onBuyEvent(event.id) : onRequireLogin()}>Comprar entrada</button>
      <p className="hint">La compra registra una entrada a tu nombre. El saldo interno no se modifica en este trabajo práctico.</p>
    </section>
    <section className="panel">
      <h2>Publicaciones de usuarios</h2>
      {offers.length === 0 && <p>No hay publicaciones activas.</p>}
      {offers.map(offer => <OfferRow offer={offer} key={offer.id} session={session} isUser={isUser} onRequireLogin={onRequireLogin} onExchange={onExchange} />)}
      {!session && <p className="hint">Para comprar o solicitar intercambio necesitás iniciar sesión.</p>}
      {session && !isUser && <p className="hint">Esta cuenta puede explorar y administrar, pero las publicaciones y entradas pertenecen a cuentas de usuario.</p>}
    </section>
  </div>
}

function OfferRow({ offer, session, isUser, onRequireLogin, onExchange }) {
  const riskClass = offer.scan_status === 'malicious' ? 'danger' : offer.scan_status === 'suspicious' ? 'warning' : offer.scan_status === 'pending' ? 'pending' : 'safe'
  const blocked = offer.status === 'blocked' || offer.quantity <= 0
  const ownOffer = session?.user?.id && offer.publisher_id === session.user.id
  const canExchange = Boolean(session && isUser && !blocked && !ownOffer)
  const label = offer.status === 'blocked' ? 'Publicación bloqueada' : ownOffer ? 'Tu publicación' : !session ? 'Iniciar sesión' : !isUser ? 'Solo usuarios' : offer.quantity <= 0 ? 'No disponible' : 'Solicitar intercambio'
  const disabled = Boolean(session && (!isUser || blocked || ownOffer))
  return <div className="offer-row">
    <div>
      <strong>{offer.title}</strong>
      <p>Valor de referencia: ${formatMoney(offer.price)}</p>
      {offer.exchange_details && <small>Motivo / asiento / día: {offer.exchange_details}</small>}
      <span className={`scan-badge ${riskClass}`}>{offer.scan_status}</span>
      {offer.scan_status !== 'safe' && <small>{offer.scan_verdict}</small>}
    </div>
    <button disabled={disabled} onClick={() => canExchange ? onExchange(offer) : onRequireLogin()}>{label}</button>
  </div>
}

function ExchangeModal({ offer, tickets, onClose, onSubmit }) {
  const eventTickets = tickets.filter(t => t.status === 'active' && t.event_id === offer.event_id && t.id !== offer.ticket_id)
  const [form, setForm] = useState({ offered_ticket_id: eventTickets[0]?.id || '', message: '', contact_info: '' })
  const selectedTicket = eventTickets.find(t => String(t.id) === String(form.offered_ticket_id))
  const difference = selectedTicket ? Number(offer.price || 0) - Number(selectedTicket.price || 0) : 0
  return <Modal title="Solicitar intercambio" onClose={onClose}>
    <p className="hint">Vas a enviar una solicitud al usuario de la publicación. Si acepta, el sistema cambia la titularidad de ambas entradas.</p>
    <div className="exchange-target">
      <strong>{offer.title}</strong>
      <p>Entrada solicitada: #{offer.ticket_id || 'sin ID'} · Valor: ${formatMoney(offer.price)}</p>
      <small>{offer.exchange_details || 'Sin detalle de asiento/día'}</small>
    </div>
    {eventTickets.length === 0 ? <p>No tenés entradas activas de este foro para ofrecer. Primero comprá una entrada desde el botón general del foro.</p> : <form onSubmit={(e) => { e.preventDefault(); onSubmit({ ...form, offered_ticket_id: Number(form.offered_ticket_id) }) }}>
      <select value={form.offered_ticket_id} onChange={e => setForm({ ...form, offered_ticket_id: e.target.value })}>
        {eventTickets.map(t => <option key={t.id} value={t.id}>#{t.id} · {t.seat_info || t.code} · ${formatMoney(t.price)}</option>)}
      </select>
      <textarea placeholder="Mensaje: asiento, día, motivo del cambio" value={form.message} onChange={e => setForm({ ...form, message: e.target.value })} />
      <input placeholder="Contacto para coordinar diferencia de valor" value={form.contact_info} onChange={e => setForm({ ...form, contact_info: e.target.value })} />
      {difference !== 0 && <p className="payment-note">Diferencia estimada: ${formatMoney(Math.abs(difference))}. Contacto para coordinar pago: {form.contact_info || 'completá un contacto'}.</p>}
      <div className="modal-actions"><button>Enviar solicitud</button><button type="button" className="secondary-btn" onClick={onClose}>Cancelar</button></div>
    </form>}
  </Modal>
}

function Tickets({ tickets, inbox, sent, reload, setMessage, onCancel, onTransfer }) {
  const [userIdByTicket, setUserIdByTicket] = useState({})
  return <div className="tickets-layout">
    <section className="panel">
      <h2>Mis entradas</h2>
      {tickets.length === 0 && <p>Todavía no tenés entradas.</p>}
      {tickets.map(ticket => <div className="table-row" key={ticket.id}>
        <div>
          <strong>{ticket.event?.title}</strong>
          <p>ID: #{ticket.id} · Código: {ticket.code} · Estado: {ticket.status} · Valor: ${formatMoney(ticket.price)}</p>
          <small>{ticket.seat_info || 'Sin asiento cargado'} {ticket.description ? `· ${ticket.description}` : ''}</small>
        </div>
        <div className="ticket-actions">
          <button onClick={() => onCancel(ticket.id)}>Cancelar</button>
          <input type="number" min="1" placeholder="ID usuario destino" value={userIdByTicket[ticket.id] || ''} onChange={e => setUserIdByTicket({ ...userIdByTicket, [ticket.id]: e.target.value })} />
          <button onClick={() => onTransfer(ticket.id, userIdByTicket[ticket.id])}>Transferir</button>
        </div>
      </div>)}
    </section>

    <ExchangeBox inbox={inbox} sent={sent} reload={reload} setMessage={setMessage} />
  </div>
}

function ExchangeBox({ inbox, sent, reload, setMessage }) {
  async function accept(id) {
    try {
      await api.acceptExchange(id)
      setMessage('Intercambio aceptado. Las entradas cambiaron de titular.')
      await reload()
    } catch (err) {
      setMessage(err.message)
    }
  }
  async function reject(id) {
    try {
      await api.rejectExchange(id)
      setMessage('Intercambio rechazado')
      await reload()
    } catch (err) {
      setMessage(err.message)
    }
  }
  return <section className="panel exchange-box">
    <h2>Bandeja de intercambios</h2>
    <h3>Recibidas</h3>
    {inbox.length === 0 && <p>No tenés solicitudes recibidas.</p>}
    {inbox.map(x => <ExchangeRow key={x.id} exchange={x} incoming onAccept={accept} onReject={reject} />)}
    <h3>Enviadas</h3>
    {sent.length === 0 && <p>No tenés solicitudes enviadas.</p>}
    {sent.map(x => <ExchangeRow key={x.id} exchange={x} />)}
  </section>
}

function ExchangeRow({ exchange, incoming, onAccept, onReject }) {
  const diff = Number(exchange.price_difference || 0)
  return <div className="exchange-row">
    <div>
      <strong>{exchange.event?.title || exchange.offer?.title}</strong>
      <p>Estado: {exchange.status} · Solicitante: {exchange.requester?.email}</p>
      <small>Ofrece entrada #{exchange.offered_ticket_id} por entrada #{exchange.target_ticket_id}</small>
      {exchange.message && <small>Mensaje: {exchange.message}</small>}
      {diff !== 0 && <small className="payment-note">Diferencia de valor: ${formatMoney(Math.abs(diff))}. Contacto para coordinar pago: {exchange.contact_info || 'sin contacto'}.</small>}
    </div>
    {incoming && exchange.status === 'pending' && <div className="inline-actions"><button onClick={() => onAccept(exchange.id)}>Aceptar</button><button className="secondary-btn" onClick={() => onReject(exchange.id)}>Rechazar</button></div>}
  </div>
}

function Publish({ events, offers, tickets, reload, setMessage }) {
  const firstEvent = useMemo(() => events[0]?.id ? String(events[0].id) : '', [events])
  const [form, setForm] = useState({ event_id: firstEvent, ticket_id: '', title: '', price: 10000, quantity: 1, external_url: '', contact_info: '', exchange_details: '' })

  useEffect(() => {
    if (!form.event_id && firstEvent) setForm(prev => ({ ...prev, event_id: firstEvent }))
  }, [firstEvent, form.event_id])

  const availableTickets = tickets.filter(t => t.status === 'active' && (!form.event_id || String(t.event_id) === String(form.event_id)))

  async function submit(e) {
    e.preventDefault()
    try {
      const payload = { ...form, event_id: Number(form.event_id), ticket_id: Number(form.ticket_id || 0), price: Number(form.price), quantity: Number(form.ticket_id ? 1 : form.quantity) }
      const offer = await api.createOffer(payload)
      if (offer.status === 'blocked') setMessage('Publicación bloqueada')
      else if (offer.scan_status === 'pending') setMessage('Publicación recibida. Queda pendiente de revisión')
      else setMessage('Publicación creada correctamente')
      setForm(prev => ({ ...prev, ticket_id: '', title: '', price: 10000, quantity: 1, external_url: '', contact_info: '', exchange_details: '' }))
      await reload()
    } catch (err) {
      setMessage(err.message)
    }
  }
  return <div className="detail-layout">
    <section className="panel publish-panel">
      <h2>Crear publicación</h2>
      <p className="hint">La publicación sirve para solicitar intercambios. Asociá una entrada propia para que el sistema pueda cambiar titularidad y mantener trazabilidad.</p>
      <form onSubmit={submit}>
        <select value={form.event_id} onChange={e => setForm({ ...form, event_id: e.target.value, ticket_id: '' })}>
          <option value="">Seleccionar foro</option>
          {events.map(e => <option key={e.id} value={e.id}>{e.title}</option>)}
        </select>
        <select value={form.ticket_id} onChange={e => {
          const selected = availableTickets.find(t => String(t.id) === e.target.value)
          setForm({ ...form, ticket_id: e.target.value, price: selected?.price || form.price })
        }}>
          <option value="">Sin entrada asociada</option>
          {availableTickets.map(t => <option key={t.id} value={t.id}>Entrada #{t.id} · {t.seat_info || t.code} · ${formatMoney(t.price)}</option>)}
        </select>
        <input placeholder="Título de publicación" value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} />
        <input type="number" placeholder="Valor de referencia" value={form.price} onChange={e => setForm({ ...form, price: e.target.value })} />
        <input type="number" placeholder="Cantidad de entradas" value={form.quantity} disabled={Boolean(form.ticket_id)} onChange={e => setForm({ ...form, quantity: e.target.value })} />
        <input placeholder="Contacto para coordinar diferencia de valor" value={form.contact_info} onChange={e => setForm({ ...form, contact_info: e.target.value })} />
        <textarea placeholder="Descripción para intercambio: asiento, día, sector o motivo" value={form.exchange_details} onChange={e => setForm({ ...form, exchange_details: e.target.value })} />
        <textarea className="long-url" placeholder="Link externo completo para analizar. Puede ser un link largo de phishing para pruebas controladas del honeypot." value={form.external_url} onChange={e => setForm({ ...form, external_url: e.target.value })} spellCheck="false" />
        <button>Publicar</button>
      </form>
      <p className="hint">Si el link es malicioso, al usuario solo se le informa que la publicación fue bloqueada.</p>
    </section>
    <section className="panel">
      <h2>Mis publicaciones</h2>
      {offers.length === 0 && <p>No hay publicaciones cargadas.</p>}
      {offers.map(o => <MyOfferRow offer={o} key={o.id} />)}
    </section>
  </div>
}

function MyOfferRow({ offer }) {
  if (offer.status === 'blocked') {
    return <div className="table-row blocked-only"><strong>Publicación bloqueada</strong></div>
  }
  return <div className="table-row">
    <div>
      <strong>{offer.title}</strong>
      <p>{publicOfferStatus(offer)} · Entrada asociada: {offer.ticket_id ? `#${offer.ticket_id}` : 'sin asociar'}</p>
      <small>{offer.event?.title}</small>
    </div>
    <strong>${formatMoney(offer.price)}</strong>
  </div>
}

function publicOfferStatus(offer) {
  if (offer.status === 'blocked') return 'Publicación bloqueada'
  if (offer.status === 'paused') return 'Publicación pausada'
  return 'Publicación activa'
}

function Admin({ events, reloadEvents, report, loadReport, alerts, reloadAlerts, clearAlerts, onMoreInfo, setMessage }) {
  const [form, setForm] = useState(emptyEvent)
  const [editingId, setEditingId] = useState(null)
  const [saving, setSaving] = useState(false)

  function beginEdit(event) {
    setEditingId(event.id)
    setForm({
      title: event.title || '', description: event.description || '', category: event.category || '',
      location: event.location || '', starts_at: event.starts_at ? new Date(event.starts_at).toISOString().slice(0, 16) : '',
      duration_minutes: event.duration_minutes || 120, capacity: event.capacity || 1,
      entry_price: event.entry_price || 1, image_url: event.image_url || ''
    })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  function resetForm() { setEditingId(null); setForm(emptyEvent) }

  async function submit(e) {
    e.preventDefault()
    setSaving(true)
    try {
      const payload = { ...form, starts_at: new Date(form.starts_at).toISOString(), capacity: Number(form.capacity), duration_minutes: Number(form.duration_minutes), entry_price: Number(form.entry_price) }
      if (editingId) {
        await api.updateEvent(editingId, payload)
        setMessage('Foro actualizado correctamente')
      } else {
        await api.createEvent(payload)
        setMessage('Foro creado correctamente')
      }
      resetForm()
      await reloadEvents()
    } catch (err) { setMessage(err.message) }
    finally { setSaving(false) }
  }

  async function cancelEvent(id) {
    if (!window.confirm('¿Seguro que querés dar de baja este foro?')) return
    try { await api.cancelEvent(id); setMessage('Foro cancelado'); await reloadEvents() }
    catch (err) { setMessage(err.message) }
  }

  return <div className="admin-layout">
    <section className="panel">
      <h2>{editingId ? 'Editar foro' : 'Crear foro'}</h2>
      <form onSubmit={submit}>
        <input required placeholder="Título" value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} />
        <textarea placeholder="Descripción" value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} />
        <input placeholder="Categoría" value={form.category} onChange={e => setForm({ ...form, category: e.target.value })} />
        <input required placeholder="Lugar" value={form.location} onChange={e => setForm({ ...form, location: e.target.value })} />
        <input required type="datetime-local" value={form.starts_at} onChange={e => setForm({ ...form, starts_at: e.target.value })} />
        <input required min="1" type="number" placeholder="Duración (minutos)" value={form.duration_minutes} onChange={e => setForm({ ...form, duration_minutes: e.target.value })} />
        <input required min="1" type="number" placeholder="Capacidad" value={form.capacity} onChange={e => setForm({ ...form, capacity: e.target.value })} />
        <input required min="0.01" step="0.01" type="number" placeholder="Valor entrada general" value={form.entry_price} onChange={e => setForm({ ...form, entry_price: e.target.value })} />
        <input type="url" placeholder="Imagen URL (opcional)" value={form.image_url} onChange={e => setForm({ ...form, image_url: e.target.value })} />
        <div className="inline-actions"><button disabled={saving}>{saving ? 'Guardando…' : editingId ? 'Guardar cambios' : 'Crear foro'}</button>{editingId && <button type="button" className="secondary-btn" onClick={resetForm}>Cancelar edición</button>}</div>
      </form>
    </section>

    <section className="panel">
      <h2>Foros y reportes</h2>
      {events.map(e => <div className="table-row" key={e.id}>
        <div><strong>{e.title}</strong><small>{e.status === 'cancelled' ? ' · Cancelado' : ''} · ${formatMoney(e.entry_price)}</small></div>
        <div className="inline-actions"><button onClick={() => beginEdit(e)}>Editar</button><button onClick={() => loadReport(e.id)}>Reporte</button><button className="secondary-btn" disabled={e.status === 'cancelled'} onClick={() => cancelEvent(e.id)}>Dar de baja</button></div>
      </div>)}
      {report && <div className="report">
        <h3>Reporte: {report.event.title}</h3>
        <p>Entradas activas: {report.sold}</p><p>Capacidad total: {report.capacity}</p><p>Disponibles: {report.available}</p><p>Ocupación: {Number(report.occupation_pct || 0).toFixed(2)}%</p>
        <h4>Listado de compradores</h4>
        {!report.buyers?.length ? <p>Todavía no hay entradas activas.</p> : report.buyers.map(ticket => <div className="table-row" key={ticket.id}><span>{ticket.owner?.name || `Usuario #${ticket.owner_id}`} · {ticket.owner?.email || 'correo no disponible'}</span><small>Entrada #{ticket.id} · {ticket.code}</small></div>)}
      </div>}
    </section>

    <section className="panel admin-alerts">
      <div className="panel-title-row"><h2>Alertas del honeypot</h2><div className="inline-actions"><button onClick={reloadAlerts}>Actualizar</button><button className="secondary-btn" onClick={clearAlerts}>Limpiar alertas</button></div></div>
      {alerts.length === 0 && <p>No hay publicaciones sospechosas o bloqueadas.</p>}
      {alerts.map(o => <div className="alert-row" key={o.id}><div><strong>{o.title}</strong><p>{o.event?.title} · usuario: {o.request_user_name || o.publisher?.name || 'sin nombre'} ({o.request_user_email || o.publisher?.email})</p><p>IP capturada: {o.request_ip || 'sin dato'}</p><p>Navegador: {o.browser_data || o.user_agent || 'sin dato'}</p><span className={`scan-badge ${o.scan_status === 'malicious' ? 'danger' : 'warning'}`}>{o.scan_status}</span><small>{o.scan_verdict}</small><small className="evidence-url">{o.external_url}</small></div><div className="alert-actions"><strong>{o.status}</strong><button onClick={() => onMoreInfo(o)}>Más información</button></div></div>)}
    </section>
  </div>
}

function EvidenceModal({ evidence, onClose }) {
  return <Modal title="Más información del honeypot" onClose={onClose}>
    <div className="evidence-modal">
      <p><strong>Usuario:</strong> {evidence.request_user_name || evidence.publisher?.name} · {evidence.request_user_email || evidence.publisher?.email}</p>
      <p><strong>IP:</strong> {evidence.request_ip || 'sin dato'}</p>
      <p><strong>User-Agent:</strong> {evidence.user_agent || 'sin dato'}</p>
      <p><strong>URL publicada:</strong></p>
      <pre>{evidence.external_url || 'sin URL'}</pre>
      <p><strong>Headers capturados por el servidor:</strong></p>
      <pre>{evidence.request_headers || 'sin headers'}</pre>
      <p><strong>Telemetría enviada por el navegador:</strong></p>
      <pre>{evidence.telemetry_json || evidence.browser_data || 'sin telemetría'}</pre>
      <p className="hint">Estos datos son metadatos defensivos obtenidos por la aplicación y el navegador. No se accede a archivos, credenciales ni información privada del dispositivo.</p>
    </div>
  </Modal>
}

function Modal({ title, onClose, children }) {
  return <div className="modal-backdrop" onMouseDown={onClose}>
    <div className="modal-card" onMouseDown={e => e.stopPropagation()}>
      <div className="modal-title"><h2>{title}</h2><button className="round-btn" onClick={onClose}>×</button></div>
      {children}
    </div>
  </div>
}

function formatMoney(value) {
  return Number(value || 0).toLocaleString('es-AR', { minimumFractionDigits: 0, maximumFractionDigits: 0 })
}

export default App
