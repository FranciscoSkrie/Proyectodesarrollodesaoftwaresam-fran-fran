# TicketGuard OpenBuy

Sistema de gestión de eventos y entradas con enfoque tipo marketplace/OpenBuy. Permite que clientes compren entradas, vendedores publiquen ofertas y administradores gestionen eventos, reportes y control general del catálogo.

## Tecnologías utilizadas

- Backend: Go, Gin, GORM, MySQL, JWT y bcrypt.
- Frontend: React + Vite.
- Base de datos: MySQL 8.
- DevOps: Docker, Dockerfile y Docker Compose.
- Seguridad extra del proyecto: cliente de análisis de links para ofertas externas, con implementación mock local y adaptación preparada para VirusTotal si se configura `VT_API_KEY`.

## Estructura del proyecto

```text
ticketguard-openbuy/
├── backend/
│   ├── clients/        # clientes externos, como análisis de links
│   ├── config/         # variables de entorno y conexión
│   ├── controllers/    # handlers HTTP
│   ├── dao/            # acceso a datos con GORM
│   ├── domain/         # entidades del negocio
│   ├── middleware/     # autenticación/autorización
│   ├── services/       # lógica de negocio
│   ├── utils/          # JWT, password y respuestas
│   └── main.go
├── frontend/
│   └── src/
├── docs/
├── docker-compose.yml
└── .env.example
```

## Cómo levantar todo

1. Copiá el archivo de variables:

```bash
cp .env.example .env
```

En Windows podés crear un archivo `.env` copiando el contenido de `.env.example`.

2. Levantá el sistema completo:

```bash
docker compose up --build
```

3. Abrí el frontend:

```text
http://localhost:5173
```

4. API backend:

```text
http://localhost:8080/api
```

5. MySQL desde DBeaver/TablePlus, si querés inspeccionar:

```text
Host: localhost
Port: 3307
User: ticketguard
Password: ticketguard123
Database: ticketguard_db
```

## Usuarios iniciales cargados por seed

```text
Admin:    admin@ticketguard.test    / Admin123!
Vendedor: seller@ticketguard.test   / Seller123!
Cliente:  cliente@ticketguard.test  / Cliente123!
```

## Flujo de prueba sugerido

1. Entrar como admin y crear o editar eventos.
2. Entrar como vendedor y crear una oferta para un evento.
3. Entrar como cliente, abrir el detalle de un evento, comprar una oferta y luego ver “Mis Entradas”.
4. Cancelar una entrada o transferirla a otro usuario por email.
5. Entrar como admin y ver el reporte de ocupación/ventas.

## Endpoints principales

### Autenticación

```text
POST /api/auth/register
POST /api/auth/login
```

### Público

```text
GET /api/events
GET /api/events/:id
GET /api/events/:id/offers
```

### Cliente autenticado

```text
POST /api/offers/:id/buy
GET  /api/me/tickets
POST /api/tickets/:id/cancel
POST /api/tickets/:id/transfer
```

### Vendedor

```text
GET  /api/seller/offers
POST /api/seller/offers
```

### Administrador

```text
POST   /api/admin/events
PUT    /api/admin/events/:id
DELETE /api/admin/events/:id
GET    /api/admin/events/:id/report
```

## Tests del backend

```bash
cd backend
go test ./...
```

También podés obtener cobertura:

```bash
go test ./... -cover
```

## Decisiones de diseño

1. Las compras, cancelaciones y transferencias se resuelven en la capa de servicios. Los controladores solo parsean HTTP, validan entrada superficial y devuelven status codes.
2. La capa DAO encapsula GORM y evita que los controladores tengan consultas o SQL.
3. Los tickets no se eliminan físicamente cuando se cancelan: pasan a estado `cancelled`. Esto conserva historial y permite que el reporte sea auditable.
4. Las contraseñas se guardan con bcrypt, no en texto plano.
5. Los links publicados por vendedores pasan por un `LinkScanner`. Sin API key funciona con un mock local; con `VT_API_KEY` queda preparado para integrar VirusTotal.

## Diagrama de base de datos

El diagrama fuente está en:

```text
docs/db-diagram.mmd
```

Podés pegarlo en un visor Mermaid para exportarlo como imagen e incrustarlo en el README final del repositorio.
