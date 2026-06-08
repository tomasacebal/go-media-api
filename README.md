# AceTransfer

API Go con Fiber para enviar y compartir archivos con links cortos, estilo WeTransfer.

La raiz `/` sirve una web protegida por login para crear envios, copiar links, administrar archivos, cuotas, usuarios y API keys.

## Variables de entorno

| Variable | Default | Descripcion |
| --- | --- | --- |
| `PORT` | `8080` | Puerto HTTP. |
| `MEDIA_MAX_UPLOAD_MB` | `0` | Limite por archivo en MB. `0` significa sin limite propio de la app. |
| `MEDIA_STORAGE_PATH` | `media` | Carpeta local donde se guardan archivos y metadata SQLite. |
| `MEDIA_PUBLIC_BASE_URL` | `http://localhost:8080` | Base usada para construir links `/s/:code`. |
| `ADMIN_USERNAME` | `admin` | Login admin heredado. Se usa si `ADMIN_EMAIL` no existe. |
| `ADMIN_EMAIL` | `ADMIN_USERNAME` | Email o usuario del admin bootstrap. |
| `ADMIN_NAME` | `Admin` | Nombre visible del admin bootstrap. |
| `ADMIN_PASSWORD` | requerido | Password inicial del admin bootstrap. |
| `SESSION_SECRET` | requerido | Secreto para firmar cookies. Debe tener al menos 32 caracteres. |
| `SESSION_TTL_HOURS` | `12` | Duracion de sesion web. |
| `COOKIE_SECURE` | `false` | Usa cookies Secure cuando corre detras de HTTPS. |
| `USER_DEFAULT_QUOTA_GB` | `10` | Cuota default por usuario. |
| `SHARE_DEFAULT_TTL_DAYS` | `30` | Vencimiento default de links cortos. |

## Build

```bash
go build ./...
go build -o AceTransfer.exe .
```

## Modelo de producto

- SaaS multiusuario por invitacion.
- El primer arranque crea un admin desde `.env` si no hay usuarios.
- Cada archivo, link y API key tiene `owner_id`.
- Los usuarios solo ven sus archivos, envios, links y API keys.
- Admin puede listar y administrar usuarios, y ver recursos globales usando `?all=true`.
- Admin puede cambiar cuota y vencimiento default de usuarios ya creados.
- `share_ttl_days = 0` significa links sin vencimiento por defecto para esa cuenta.
- Los archivos no se sirven desde `/media`; toda descarga pasa por sesion, API key o share vigente.

## Cuotas FIFO

- Cada usuario tiene una cuota en bytes.
- Si un upload entra en cuota, se guarda normal.
- Si un archivo o envio supera la cuota total, responde `file_exceeds_quota`.
- Si falta espacio, la API responde `409 quota_cleanup_required` con los archivos mas viejos que se borrarian.
- Al reenviar con `confirm_fifo=true`, se borran archivos por `created_at` ascendente hasta liberar espacio.
- Los links directos de archivos borrados se revocan. Si un envio queda sin archivos, tambien se revoca.

## API

- `GET /login`
- `POST /login`
- `POST /logout`
- `GET /api/v1/me`
- `POST /api/v1/account/password`
- `GET /api/v1/users` admin
- `POST /api/v1/users` admin
- `PATCH /api/v1/users/:id` admin
- `GET /api/v1/api-keys`
- `POST /api/v1/api-keys`
- `DELETE /api/v1/api-keys/:id`
- `POST /api/v1/files` con sesion o API key `write`
- `GET /api/v1/files` con sesion o API key `read`
- `GET /api/v1/files/:id` con sesion o API key `read`
- `GET /api/v1/files/:id/download` con sesion o API key `read`
- `DELETE /api/v1/files/:id` con sesion o API key `delete`
- `POST /api/v1/transfers` con sesion o API key `write`
- `GET /api/v1/transfers` con sesion o API key `read`
- `GET /api/v1/transfers/:id` con sesion o API key `read`
- `POST /api/v1/shares` con sesion o API key `write`
- `GET /api/v1/shares` con sesion o API key `read`
- `DELETE /api/v1/shares/:id` con sesion o API key `delete`
- `GET /api/v1/public/shares/:code`
- `GET /s/:code`
- `GET /s/:code/download`

## Ejemplos

Crear un envio con link corto:

```bash
curl -b cookies.txt \
  -F "files=@./archivo-a.txt" \
  -F "files=@./archivo-b.txt" \
  -F "title=Entrega" \
  http://localhost:8080/api/v1/transfers
```

Crear un envio sin vencimiento:

```bash
curl -b cookies.txt \
  -F "files=@./archivo.zip" \
  -F "never_expires=true" \
  http://localhost:8080/api/v1/transfers
```

Confirmar limpieza FIFO:

```bash
curl -b cookies.txt \
  -F "files=@./archivo-grande.zip" \
  -F "confirm_fifo=true" \
  http://localhost:8080/api/v1/transfers
```
