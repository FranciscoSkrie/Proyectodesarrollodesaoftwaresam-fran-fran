# Cómo armar el proyecto — v6

1. Descomprimir `ticketguard-openbuy-v6.zip`.
2. Abrir la carpeta en VS Code.
3. Ejecutar limpieza si venís de una versión anterior:

```bash
docker compose down -v
```

4. Construir y levantar:

```bash
docker compose build --no-cache
docker compose up
```

5. Abrir:

```text
Frontend:   http://localhost:5173
Backend:    http://localhost:8080/api
phpMyAdmin: http://localhost:8081
```

6. Credenciales de base:

```text
Servidor phpMyAdmin: mysql
Usuario: ticketguard
Password: ticketguard123
Base: ticketguard_db
```

## Notas v6

- MySQL corre como contenedor independiente y phpMyAdmin permite verlo visualmente.
- El botón de compra está abajo del detalle del foro.
- Las publicaciones disparan solicitudes de intercambio.
- Las solicitudes recibidas se aceptan/rechazan desde “Mis entradas e intercambios”.
- El panel admin permite ver más información de alertas y limpiarlas sin borrar evidencia.
