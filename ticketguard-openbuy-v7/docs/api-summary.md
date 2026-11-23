# Resumen de API REST v6

## Público

- `GET /api/events`: lista foros activos.
- `GET /api/events/:id`: detalle de foro.
- `GET /api/events/:id/offers`: publicaciones activas del foro.

## Autenticación

- `POST /api/auth/register`: crea usuario común.
- `POST /api/auth/login`: devuelve JWT.
- `POST /api/me/password`: actualiza contraseña del usuario autenticado.

## Usuario

- `POST /api/events/:id/buy`: compra/registración de entrada general del foro. No modifica saldo interno.
- `GET /api/me/tickets`: entradas del usuario.
- `POST /api/tickets/:id/cancel`: cancela entrada propia.
- `POST /api/tickets/:id/transfer`: transfiere entrada propia a otro usuario por ID.
- `GET /api/me/offers`: publicaciones propias.
- `POST /api/me/offers`: crea publicación de intercambio, opcionalmente asociada a una entrada propia.
- `POST /api/offers/:id/exchange`: solicita intercambio sobre una publicación.
- `GET /api/me/exchanges/inbox`: solicitudes recibidas.
- `GET /api/me/exchanges/sent`: solicitudes enviadas.
- `POST /api/exchanges/:id/accept`: acepta intercambio recibido y cambia titularidad.
- `POST /api/exchanges/:id/reject`: rechaza intercambio recibido.

## Administrador

- `POST /api/admin/events`: crea foro.
- `PUT /api/admin/events/:id`: actualiza foro.
- `DELETE /api/admin/events/:id`: cancela foro.
- `GET /api/admin/events/:id/report`: reporte de ocupación.
- `GET /api/admin/security/offers`: publicaciones sospechosas/bloqueadas no limpiadas.
- `POST /api/admin/security/alerts/clear`: limpia alertas del panel sin eliminar evidencia.
