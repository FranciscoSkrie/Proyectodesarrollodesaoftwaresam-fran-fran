# Resumen de API REST

La API usa JSON y JWT Bearer Token.

## Status codes usados

- `200 OK`: lectura o acción exitosa.
- `201 Created`: recurso creado.
- `204 No Content`: eliminación/cancelación sin body cuando aplique.
- `400 Bad Request`: entrada inválida.
- `401 Unauthorized`: falta token o token inválido.
- `403 Forbidden`: rol insuficiente.
- `404 Not Found`: recurso inexistente o no visible para el usuario.
- `409 Conflict`: conflicto de negocio, por ejemplo sin stock/cupo.
- `500 Internal Server Error`: error inesperado.

## Auth

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "cliente@ticketguard.test",
  "password": "Cliente123!"
}
```

Respuesta:

```json
{
  "token": "...",
  "user": { "id": 3, "name": "Cliente Demo", "email": "cliente@ticketguard.test", "role": "cliente" }
}
```
