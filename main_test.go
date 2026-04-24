package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/config"
	"github.com/tomasacebal/go-media-api/internal/media"
)

func TestRootServesGallery(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("root fallo: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("leer body fallo: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), "Media Gallery") {
		t.Fatal("root deberia servir la galeria")
	}
}

func TestUploadValidImageReturnsMetadata(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	resp := uploadFile(t, app, "foto.png", validPNG(), map[string]string{"visibility": "public"})
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}

	payload := decodeData(t, resp.Body)
	if payload.MIMEType != "image/png" {
		t.Fatalf("mime esperado image/png, recibido %s", payload.MIMEType)
	}
	if payload.SizeBytes == 0 {
		t.Fatal("size_bytes esperado mayor a cero")
	}
	if payload.PublicURL == "" {
		t.Fatal("public_url esperado")
	}
}

func TestUploadValidPDFReturnsMetadata(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	resp := uploadFile(t, app, "archivo.pdf", validPDF(), map[string]string{"visibility": "public"})
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}

	payload := decodeData(t, resp.Body)
	if payload.MIMEType != "application/pdf" {
		t.Fatalf("mime esperado application/pdf, recibido %s", payload.MIMEType)
	}
}

func TestListReturnsUploadedFiles(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	resp := uploadFile(t, app, "foto.png", validPNG(), map[string]string{"visibility": "public"})
	defer resp.Body.Close()
	uploaded := decodeData(t, resp.Body)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
	listResp, err := app.Test(req)
	if err != nil {
		t.Fatalf("list fallo: %v", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", listResp.StatusCode)
	}

	files := decodeList(t, listResp.Body)
	if len(files) != 1 {
		t.Fatalf("cantidad esperada 1, recibida %d", len(files))
	}
	if files[0].ID != uploaded.ID {
		t.Fatalf("id esperado %s, recibido %s", uploaded.ID, files[0].ID)
	}
}

func TestUploadRenamedExecutableIsRejected(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	resp := uploadFile(t, app, "malicioso.jpg", []byte("MZ fake executable"), nil)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnsupportedMediaType {
		t.Fatalf("status esperado 415, recibido %d", resp.StatusCode)
	}
}

func TestUploadTooLargeIsRejected(t *testing.T) {
	app, cleanup := newTestApp(t, 8)
	defer cleanup()

	resp := uploadFile(t, app, "foto.png", validPNG(), nil)
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusRequestEntityTooLarge {
		t.Fatalf("status esperado 413, recibido %d", resp.StatusCode)
	}
}

func TestDownloadUsesStoredContentType(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	uploadResp := uploadFile(t, app, "foto.png", validPNG(), nil)
	defer uploadResp.Body.Close()
	file := decodeData(t, uploadResp.Body)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("download fallo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "image/png" {
		t.Fatalf("content-type esperado image/png, recibido %s", got)
	}
}

func TestDeletePreventsDownload(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	uploadResp := uploadFile(t, app, "foto.png", validPNG(), nil)
	defer uploadResp.Body.Close()
	file := decodeData(t, uploadResp.Body)

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+file.ID, nil)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete fallo: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status esperado 204, recibido %d", deleteResp.StatusCode)
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	downloadResp, err := app.Test(downloadReq)
	if err != nil {
		t.Fatalf("download fallo: %v", err)
	}
	defer downloadResp.Body.Close()
	if downloadResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status esperado 404, recibido %d", downloadResp.StatusCode)
	}
}

func TestPrivateDownloadIsForbiddenWithoutAuth(t *testing.T) {
	app, cleanup := newTestApp(t, 1024*1024)
	defer cleanup()

	uploadResp := uploadFile(t, app, "foto.png", validPNG(), map[string]string{"visibility": "private"})
	defer uploadResp.Body.Close()
	file := decodeData(t, uploadResp.Body)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("download fallo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status esperado 403, recibido %d", resp.StatusCode)
	}
}

func newTestApp(t *testing.T, maxUploadBytes int64) (*fiber.App, func()) {
	t.Helper()

	storagePath := t.TempDir()
	cfg := config.Config{
		Port: "0",
		Media: config.MediaConfig{
			MaxUploadMB:    1,
			MaxUploadBytes: maxUploadBytes,
			StoragePath:    storagePath,
			PublicBaseURL:  "http://example.test",
			SQLitePath:     filepath.Join(storagePath, "metadata.sqlite"),
		},
	}

	app, cleanup, err := buildApp(cfg, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatalf("buildApp fallo: %v", err)
	}

	return app, func() {
		if err := cleanup(); err != nil {
			t.Fatalf("cleanup fallo: %v", err)
		}
	}
}

func uploadFile(t *testing.T, app *fiber.App, filename string, content []byte, fields map[string]string) *http.Response {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("crear form file fallo: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("escribir form file fallo: %v", err)
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("escribir field fallo: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("cerrar writer fallo: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("upload fallo: %v", err)
	}

	return resp
}

func decodeData(t *testing.T, reader io.Reader) media.File {
	t.Helper()

	var payload struct {
		Data media.File `json:"data"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		t.Fatalf("decode fallo: %v", err)
	}
	return payload.Data
}

func decodeList(t *testing.T, reader io.Reader) []media.File {
	t.Helper()

	var payload struct {
		Data []media.File `json:"data"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		t.Fatalf("decode list fallo: %v", err)
	}
	return payload.Data
}

func validPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xff, 0xff, 0x3f,
		0x00, 0x05, 0xfe, 0x02, 0xfe, 0xdc, 0xcc, 0x59,
		0xe7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

func validPDF() []byte {
	return []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF")
}
