# go-media-api

API Go con Fiber para publicar y distribuir archivos desde storage local.

La raiz `/` sirve una web protegida por login para subir, listar, abrir, borrar assets y administrar api keys.

## Variables de entorno

| Variable | Default | Descripcion |
| --- | --- | --- |
| `PORT` | `8080` | Puerto HTTP. |
| `MEDIA_MAX_UPLOAD_MB` | `0` | Limite por upload en MB. `0` significa sin limite propio de la app. |
| `MEDIA_STORAGE_PATH` | `media` | Carpeta local donde se guardan archivos y metadata SQLite. |
| `MEDIA_PUBLIC_BASE_URL` | `http://localhost:8080` | Base usada para construir `public_url`. |
| `ADMIN_USERNAME` | `admin` | Usuario admin para la web. |
| `ADMIN_PASSWORD` | requerido | Password admin para la web. |
| `SESSION_SECRET` | requerido | Secreto para firmar cookies. Debe tener al menos 32 caracteres. |

## Build

`go build ./...` verifica que todos los paquetes compilen. Para generar el binario en Windows:

```bash
go build -o go-media-api.exe .
```

## Auth

- La web requiere login con `ADMIN_USERNAME` y `ADMIN_PASSWORD`.
- Las sesiones se guardan en cookie HTTP-only firmada con `SESSION_SECRET`.
- Las api keys se crean desde la web autenticada.
- El secreto completo de una api key solo aparece al crearla.
- Clientes externos deben enviar `X-API-Key: <key>` o `Authorization: Bearer <key>`.

## Scopes

- `read`: listar metadata, leer metadata y descargar privados.
- `write`: subir archivos por API.
- `delete`: borrar archivos por API.

## Endpoints

- `GET /login`
- `POST /login`
- `POST /logout`
- `GET /api/v1/api-keys`
- `POST /api/v1/api-keys`
- `DELETE /api/v1/api-keys/:id`
- `POST /api/v1/media/upload` con api key `write`
- `POST /web/media/upload` con sesion web
- `GET /api/v1/media` con sesion web o api key `read`
- `GET /api/v1/media/:id` con sesion web o api key `read`
- `GET /api/v1/media/:id/download`
- `DELETE /api/v1/media/:id` con sesion web o api key `delete`

Los archivos publicos se descargan sin credenciales. Los privados requieren sesion web o api key `read`.

## Ejemplo de upload por API

```bash
curl -H "X-API-Key: $MEDIA_API_KEY" -F "file=@./archivo.bin" -F "visibility=private" http://localhost:8080/api/v1/media/upload
```
