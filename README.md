# go-media-api

API Go con Fiber para publicar y distribuir imagenes y PDFs.

## Variables de entorno

| Variable | Default | Descripcion |
| --- | --- | --- |
| `PORT` | `8080` | Puerto HTTP. |
| `MEDIA_MAX_UPLOAD_MB` | `10` | Tamaño maximo por upload en MB. |
| `MEDIA_STORAGE_PATH` | `media` | Carpeta local donde se guardan archivos y metadata SQLite. |
| `MEDIA_PUBLIC_BASE_URL` | `http://localhost:8080` | Base usada para construir `public_url`. |

## Build

`go build ./...` solo verifica que todos los paquetes compilen. Para generar el binario en Windows:

```bash
go build -o go-media-api.exe .
```

## Endpoints

- `POST /api/v1/media/upload`
  - Multipart field obligatorio: `file`.
  - Campos opcionales: `visibility`, `title`, `description`, `category`.
- `GET /api/v1/media/:id`
- `GET /api/v1/media/:id/download`
- `DELETE /api/v1/media/:id`

## Ejemplo de upload

```bash
curl -F "file=@./foto.png" -F "visibility=public" http://localhost:8080/api/v1/media/upload
```

Los archivos privados devuelven `403` al descargar hasta integrar autenticacion real.
