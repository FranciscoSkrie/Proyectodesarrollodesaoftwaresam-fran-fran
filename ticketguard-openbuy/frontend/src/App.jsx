import { useEffect, useState } from 'react'
import { api } from './api.js'

const emptyEvent = {
  title: '', description: '', category: '', location: '', starts_at: '', duration_minutes: 120, capacity: 100, image_url: ''
}

function App() {
  const [session, setSession] = useState(() => {
    const token = localStorage.getItem('token')
    const user = localStorage.getItem('user')
    return token && user ? { token, user: JSON.parse(user) } : null
  })
  const [view, setView] = useState('home')
  const [events, setEvents] = useState([])
  const [selectedEvent, setSelectedEvent] = useState(null)
  const [offers, setOffers] = useState([])
  const [tickets, setTickets] = useState([])
  const [sellerOffers, setSellerOffers] = useState([])
  const [report, setReport] = useState(null)
  const [message, setMessage] = useState('')
  const [filters, setFilters] = useState({ q: '', category: '' })

  async function loadEvents() {
    try {
      setEvents(await api.events(filters))
    } catch (err) { setMessage(err.message) }
  }

  useEffect(() => { loadEvents() }, [])

  function saveSession(resp) {
    localStorage.setItem('token', resp.token)
    localStorage.setItem('user', JSON.stringify(resp.user))
    setSession(resp)
    setView('home')
    setMessage('Sesión iniciada correctamente')
  }

  function logout() {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setSession(null)
    setView('home')
  }

  async function openEvent(event) {
    setSelectedEvent(event)
    setOffers(await api.offers(event.id))
    setView('detail')
  }

  async function loadTickets() {
    setTickets(await api.myTickets())
    setView('tickets')
  }

  async function loadSellerOffers() {
    setSellerOffers(await api.sellerOffers())
    setView('seller')
  }

  async function loadReport(eventId) {
    setReport(await api.report(eventId))
  }

  return (
    <div className="app">
      <header className="topbar">
        <div>
          <h1>TicketGuard OpenBuy</h1>
          <p>Marketplace de entradas con control de links sospechosos.</p>
        </div>
        <nav>
          <button onClick={() => setView('home')}>Eventos</button>
          {session && <button onClick={loadTickets}>Mis Entradas</button>}
          {session?.user?.role === 'vendedor' && <button onClick={loadSellerOffers}>Vendedor</button>}
          {session?.user?.role === 'admin' && <button onClick={() => setView('admin')}>Admin</button>}
          {!session ? <button onClick={() => setView('login')}>Login</button> : <button onClick={logout}>Salir</button>}
        </nav>
      </header>

      {session && <p className="session">Usuario: {session.user.name} · rol: {session.user.role}</p>}
      {message && <div className="message" onClick={() => setMessage('')}>{message}</div>}

      {view === 'home' && <Home events={events} filters={filters} setFilters={setFilters} onSearch={loadEvents} onOpen={openEvent} />}
      {view === 'login' && <Login onLogin={saveSession} setMessage={setMessage} />}
      {view === 'detail' && <Detail event={selectedEvent} offers={offers} session={session} onBuy={async (id) => {
        try { await api.buy(id); setMessage('Congrats: compra realizada correctamente'); await loadTickets() } catch (err) { setMessage(err.message) }
      }} />}
      {view === 'tickets' && <Tickets tickets={tickets} onCancel={async (id) => { await api.cancelTicket(id); setMessage('Entrada cancelada'); await loadTickets() }} onTransfer={async (id, email) => { await api.transferTicket(id, email); setMessage('Entrada transferida'); await loadTickets() }} />}
      {view === 'seller' && <Seller events={events} offers={sellerOffers} reload={loadSellerOffers} setMessage={setMessage} />}
      {view === 'admin' && <Admin events={events} reloadEvents={loadEvents} report={report} loadReport={loadReport} setMessage={setMessage} />}
    </div>
  )
}

function Home({ events, filters, setFilters, onSearch, onOpen }) {
  return <main>
    <section className="card hero">
      <h2>Catálogo de Eventos</h2>
      <div className="filters">
        <input placeholder="Buscar por título, descripción o lugar" value={filters.q} onChange={e => setFilters({ ...filters, q: e.target.value })} />
        <input placeholder="Categoría" value={filters.category} onChange={e => setFilters({ ...filters, category: e.target.value })} />
        <button onClick={onSearch}>Filtrar</button>
      </div>
    </section>
    <section className="grid">
      {events.map(event => <article className="card event" key={event.id}>
        {event.image_url && <img src={event.image_url} alt={event.title} />}
        <h3>{event.title}</h3>
        <p>{event.location} · {event.category}</p>
        <p>{new Date(event.starts_at).toLocaleString()}</p>
        <button onClick={() => onOpen(event)}>Ver detalle</button>
      </article>)}
    </section>
  </main>
}

function Login({ onLogin, setMessage }) {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({ name: '', email: 'cliente@ticketguard.test', password: 'Cliente123!', role: 'cliente' })
  async function submit(e) {
    e.preventDefault()
    try {
      const resp = mode === 'login' ? await api.login(form) : await api.register(form)
      onLogin(resp)
    } catch (err) { setMessage(err.message) }
  }
  return <main className="card narrow">
    <h2>{mode === 'login' ? 'Iniciar sesión' : 'Crear cuenta'}</h2>
    <form onSubmit={submit}>
      {mode === 'register' && <input placeholder="Nombre" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />}
      <input placeholder="Email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} />
      <input placeholder="Password" type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} />
      {mode === 'register' && <select value={form.role} onChange={e => setForm({ ...form, role: e.target.value })}>
        <option value="cliente">Cliente</option>
        <option value="vendedor">Vendedor</option>
      </select>}
      <button>{mode === 'login' ? 'Entrar' : 'Registrarme'}</button>
    </form>
    <button className="link" onClick={() => setMode(mode === 'login' ? 'register' : 'login')}>{mode === 'login' ? 'Crear una cuenta' : 'Ya tengo cuenta'}</button>
    <div className="demo">
      <p><strong>Demo:</strong></p>
      <p>admin@ticketguard.test / Admin123!</p>
      <p>seller@ticketguard.test / Seller123!</p>
      <p>cliente@ticketguard.test / Cliente123!</p>
    </div>
  </main>
}

function Detail({ event, offers, session, onBuy }) {
  if (!event) return null
  return <main className="detail">
    <section className="card">
      <h2>{event.title}</h2>
      <p>{event.description}</p>
      <p><strong>Lugar:</strong> {event.location}</p>
      <p><strong>Fecha:</strong> {new Date(event.starts_at).toLocaleString()}</p>
      <p><strong>Cupo:</strong> {event.capacity}</p>
    </section>
    <section className="card">
      <h2>Ofertas disponibles</h2>
      {offers.length === 0 && <p>No hay ofertas activas.</p>}
      {offers.map(offer => <div className="row" key={offer.id}>
        <div>
          <strong>{offer.title}</strong>
          <p>${offer.price} · disponibles: {offer.quantity}</p>
          <small>Scan: {offer.scan_status} · {offer.scan_verdict}</small>
        </div>
        <button disabled={!session || session.user.role === 'vendedor'} onClick={() => onBuy(offer.id)}>Comprar</button>
      </div>)}
      {!session && <p className="hint">Para comprar necesitás iniciar sesión.</p>}
    </section>
  </main>
}

function Tickets({ tickets, onCancel, onTransfer }) {
  const [emailByTicket, setEmailByTicket] = useState({})
  return <main className="card">
    <h2>Mis Entradas</h2>
    {tickets.length === 0 && <p>Todavía no tenés entradas.</p>}
    {tickets.map(ticket => <div className="row" key={ticket.id}>
      <div>
        <strong>{ticket.event?.title}</strong>
        <p>Código: {ticket.code} · Estado: {ticket.status} · ${ticket.price}</p>
      </div>
      <div className="actions">
        <button onClick={() => onCancel(ticket.id)}>Cancelar</button>
        <input placeholder="email destino" value={emailByTicket[ticket.id] || ''} onChange={e => setEmailByTicket({ ...emailByTicket, [ticket.id]: e.target.value })} />
        <button onClick={() => onTransfer(ticket.id, emailByTicket[ticket.id])}>Transferir</button>
      </div>
    </div>)}
  </main>
}

function Seller({ events, offers, reload, setMessage }) {
  const [form, setForm] = useState({ event_id: '', title: '', price: 10000, quantity: 1, external_url: '' })
  async function submit(e) {
    e.preventDefault()
    try {
      await api.createOffer({ ...form, event_id: Number(form.event_id), price: Number(form.price), quantity: Number(form.quantity) })
      setMessage('Oferta creada y analizada')
      await reload()
    } catch (err) { setMessage(err.message) }
  }
  return <main className="split">
    <section className="card">
      <h2>Crear oferta</h2>
      <form onSubmit={submit}>
        <select value={form.event_id} onChange={e => setForm({ ...form, event_id: e.target.value })}>
          <option value="">Seleccionar evento</option>
          {events.map(e => <option key={e.id} value={e.id}>{e.title}</option>)}
        </select>
        <input placeholder="Título de oferta" value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} />
        <input type="number" placeholder="Precio" value={form.price} onChange={e => setForm({ ...form, price: e.target.value })} />
        <input type="number" placeholder="Cantidad" value={form.quantity} onChange={e => setForm({ ...form, quantity: e.target.value })} />
        <input placeholder="Link externo" value={form.external_url} onChange={e => setForm({ ...form, external_url: e.target.value })} />
        <button>Publicar</button>
      </form>
    </section>
    <section className="card">
      <h2>Mis ofertas</h2>
      {offers.map(o => <div className="row" key={o.id}><span>{o.title} · {o.status} · scan {o.scan_status}</span><strong>${o.price}</strong></div>)}
    </section>
  </main>
}

function Admin({ events, reloadEvents, report, loadReport, setMessage }) {
  const [form, setForm] = useState(emptyEvent)
  async function submit(e) {
    e.preventDefault()
    try {
      await api.createEvent({ ...form, starts_at: new Date(form.starts_at).toISOString(), capacity: Number(form.capacity), duration_minutes: Number(form.duration_minutes) })
      setMessage('Evento creado')
      setForm(emptyEvent)
      await reloadEvents()
    } catch (err) { setMessage(err.message) }
  }
  return <main className="split">
    <section className="card">
      <h2>Crear evento</h2>
      <form onSubmit={submit}>
        <input placeholder="Título" value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} />
        <textarea placeholder="Descripción" value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} />
        <input placeholder="Categoría" value={form.category} onChange={e => setForm({ ...form, category: e.target.value })} />
        <input placeholder="Lugar" value={form.location} onChange={e => setForm({ ...form, location: e.target.value })} />
        <input type="datetime-local" value={form.starts_at} onChange={e => setForm({ ...form, starts_at: e.target.value })} />
        <input type="number" placeholder="Duración" value={form.duration_minutes} onChange={e => setForm({ ...form, duration_minutes: e.target.value })} />
        <input type="number" placeholder="Cupo" value={form.capacity} onChange={e => setForm({ ...form, capacity: e.target.value })} />
        <input placeholder="Imagen URL" value={form.image_url} onChange={e => setForm({ ...form, image_url: e.target.value })} />
        <button>Guardar evento</button>
      </form>
    </section>
    <section className="card">
      <h2>Eventos y reportes</h2>
      {events.map(e => <div className="row" key={e.id}>
        <span>{e.title}</span>
        <div><button onClick={() => loadReport(e.id)}>Reporte</button><button onClick={async () => { await api.cancelEvent(e.id); setMessage('Evento cancelado'); await reloadEvents() }}>Cancelar</button></div>
      </div>)}
      {report && <div className="report">
        <h3>Reporte: {report.event.title}</h3>
        <p>Vendidas: {report.sold} / {report.capacity}</p>
        <p>Disponibles: {report.available}</p>
        <p>Ocupación: {report.occupation_pct.toFixed(2)}%</p>
      </div>}
    </section>
  </main>
}

export default App
