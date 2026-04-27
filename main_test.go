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
	"github.com/tomasacebal/go-media-api/internal/auth"
	"github.com/tomasacebal/go-media-api/internal/config"
	"github.com/tomasacebal/go-media-api/internal/media"
)

func TestRootRequiresLogin(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("root fallo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusSeeOther {
		t.Fatalf("status esperado 303, recibido %d", resp.StatusCode)
	}
	if location := resp.Header.Get("Location"); location != "/login" {
		t.Fatalf("location esperada /login, recibida %s", location)
	}
}

func TestLoginValidAndInvalid(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	badReq := loginRequest("admin", "bad-password")
	badResp, err := app.Test(badReq)
	if err != nil {
		t.Fatalf("login invalido fallo: %v", err)
	}
	defer badResp.Body.Close()
	if badResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status esperado 401, recibido %d", badResp.StatusCode)
	}

	goodReq := loginRequest("admin", "secret-password")
	goodResp, err := app.Test(goodReq)
	if err != nil {
		t.Fatalf("login valido fallo: %v", err)
	}
	defer goodResp.Body.Close()
	if goodResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", goodResp.StatusCode)
	}
	if len(goodResp.Cookies()) == 0 {
		t.Fatal("se esperaba cookie de sesion")
	}
}

func TestCreateAPIKeyReturnsSecretOnce(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	cookie := loginCookie(t, app)
	created := createAPIKeyWithCookie(t, app, cookie, []string{auth.ScopeRead, auth.ScopeWrite})
	if created.Secret == "" {
		t.Fatal("se esperaba secret al crear la key")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)
	req.Header.Set("Cookie", cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("listar api keys fallo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}

	keys := decodeAPIKeyList(t, resp.Body)
	if len(keys) != 1 {
		t.Fatalf("cantidad esperada 1, recibida %d", len(keys))
	}
	if keys[0].Secret != "" {
		t.Fatal("el secret no debe aparecer al listar")
	}
	if keys[0].KeyPrefix == "" {
		t.Fatal("se esperaba key_prefix")
	}
}

func TestAPIUploadRequiresWriteKey(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	noKeyResp := uploadFile(t, app, "/api/v1/media/upload", "archivo.bin", []byte("contenido"), nil, nil)
	defer noKeyResp.Body.Close()
	if noKeyResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status esperado 401, recibido %d", noKeyResp.StatusCode)
	}

	readKey := createAPIKey(t, app, auth.ScopeRead)
	readResp := uploadFile(t, app, "/api/v1/media/upload", "archivo.bin", []byte("contenido"), nil, map[string]string{
		"X-API-Key": readKey,
	})
	defer readResp.Body.Close()
	if readResp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status esperado 403, recibido %d", readResp.StatusCode)
	}
}

func TestAPIUploadWithWriteAcceptsAnyFile(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	writeKey := createAPIKey(t, app, auth.ScopeWrite)
	resp := uploadFile(t, app, "/api/v1/media/upload", "programa.exe", []byte("MZ fake executable"), map[string]string{
		"visibility": "public",
	}, map[string]string{
		"Authorization": "Bearer " + writeKey,
	})
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
	file := decodeMediaData(t, resp.Body)
	if file.Extension != "exe" {
		t.Fatalf("extension esperada exe, recibida %s", file.Extension)
	}
	if file.MIMEType == "" {
		t.Fatal("se esperaba mime detectado")
	}
}

func TestWebUploadWithSessionWorks(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	cookie := loginCookie(t, app)
	resp := uploadFile(t, app, "/web/media/upload", "nota.txt", []byte("hola mundo"), map[string]string{
		"visibility": "private",
	}, map[string]string{
		"Cookie": cookie,
	})
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
}

func TestPrivateDownloadRequiresReadKeyOrSession(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	writeKey := createAPIKey(t, app, auth.ScopeWrite)
	uploadResp := uploadFile(t, app, "/api/v1/media/upload", "privado.txt", []byte("secreto"), map[string]string{
		"visibility": "private",
	}, map[string]string{
		"X-API-Key": writeKey,
	})
	defer uploadResp.Body.Close()
	file := decodeMediaData(t, uploadResp.Body)

	openReq := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	openResp, err := app.Test(openReq)
	if err != nil {
		t.Fatalf("download abierto fallo: %v", err)
	}
	defer openResp.Body.Close()
	if openResp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status esperado 403, recibido %d", openResp.StatusCode)
	}

	readKey := createAPIKey(t, app, auth.ScopeRead)
	keyReq := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	keyReq.Header.Set("X-API-Key", readKey)
	keyResp, err := app.Test(keyReq)
	if err != nil {
		t.Fatalf("download con key fallo: %v", err)
	}
	defer keyResp.Body.Close()
	if keyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", keyResp.StatusCode)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	sessionReq.Header.Set("Cookie", loginCookie(t, app))
	sessionResp, err := app.Test(sessionReq)
	if err != nil {
		t.Fatalf("download con sesion fallo: %v", err)
	}
	defer sessionResp.Body.Close()
	if sessionResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", sessionResp.StatusCode)
	}
}

func TestPublicDownloadIsOpen(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	writeKey := createAPIKey(t, app, auth.ScopeWrite)
	uploadResp := uploadFile(t, app, "/api/v1/media/upload", "foto.png", validPNG(), map[string]string{
		"visibility": "public",
	}, map[string]string{
		"X-API-Key": writeKey,
	})
	defer uploadResp.Body.Close()
	file := decodeMediaData(t, uploadResp.Body)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/"+file.ID+"/download", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("download publico fallo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "image/png" {
		t.Fatalf("content-type esperado image/png, recibido %s", got)
	}
}

func TestListRequiresReadKeyOrSession(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("list abierto fallo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status esperado 401, recibido %d", resp.StatusCode)
	}

	readKey := createAPIKey(t, app, auth.ScopeRead)
	keyReq := httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
	keyReq.Header.Set("X-API-Key", readKey)
	keyResp, err := app.Test(keyReq)
	if err != nil {
		t.Fatalf("list con key fallo: %v", err)
	}
	defer keyResp.Body.Close()
	if keyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", keyResp.StatusCode)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/v1/media", nil)
	sessionReq.Header.Set("Cookie", loginCookie(t, app))
	sessionResp, err := app.Test(sessionReq)
	if err != nil {
		t.Fatalf("list con sesion fallo: %v", err)
	}
	defer sessionResp.Body.Close()
	if sessionResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", sessionResp.StatusCode)
	}
}

func TestDeleteRequiresDeleteKeyOrSession(t *testing.T) {
	app, cleanup := newTestApp(t, 0)
	defer cleanup()

	writeKey := createAPIKey(t, app, auth.ScopeWrite)
	uploadResp := uploadFile(t, app, "/api/v1/media/upload", "archivo.txt", []byte("borrar"), nil, map[string]string{
		"X-API-Key": writeKey,
	})
	defer uploadResp.Body.Close()
	file := decodeMediaData(t, uploadResp.Body)

	noAuthReq := httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+file.ID, nil)
	noAuthResp, err := app.Test(noAuthReq)
	if err != nil {
		t.Fatalf("delete sin auth fallo: %v", err)
	}
	defer noAuthResp.Body.Close()
	if noAuthResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status esperado 401, recibido %d", noAuthResp.StatusCode)
	}

	readKey := createAPIKey(t, app, auth.ScopeRead)
	readReq := httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+file.ID, nil)
	readReq.Header.Set("X-API-Key", readKey)
	readResp, err := app.Test(readReq)
	if err != nil {
		t.Fatalf("delete con read fallo: %v", err)
	}
	defer readResp.Body.Close()
	if readResp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status esperado 403, recibido %d", readResp.StatusCode)
	}

	deleteKey := createAPIKey(t, app, auth.ScopeDelete)
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/media/"+file.ID, nil)
	deleteReq.Header.Set("X-API-Key", deleteKey)
	deleteResp, err := app.Test(deleteReq)
	if err != nil {
		t.Fatalf("delete con delete fallo: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status esperado 204, recibido %d", deleteResp.StatusCode)
	}
}

func newTestApp(t *testing.T, maxUploadBytes int64) (*fiber.App, func()) {
	t.Helper()

	storagePath := t.TempDir()
	cfg := config.Config{
		Port: "0",
		Media: config.MediaConfig{
			MaxUploadMB:    0,
			MaxUploadBytes: maxUploadBytes,
			StoragePath:    storagePath,
			PublicBaseURL:  "http://example.test",
			SQLitePath:     filepath.Join(storagePath, "metadata.sqlite"),
		},
		Auth: config.AuthConfig{
			AdminUsername: "admin",
			AdminPassword: "secret-password",
			SessionSecret: "test-secret-with-at-least-thirty-two-chars",
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

func loginRequest(username string, password string) *http.Request {
	body := strings.NewReader("username=" + username + "&password=" + password)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	return req
}

func loginCookie(t *testing.T, app *fiber.App) string {
	t.Helper()

	resp, err := app.Test(loginRequest("admin", "secret-password"))
	if err != nil {
		t.Fatalf("login fallo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}

	parts := make([]string, 0, len(resp.Cookies()))
	for _, cookie := range resp.Cookies() {
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

func createAPIKey(t *testing.T, app *fiber.App, scopes ...string) string {
	t.Helper()

	key := createAPIKeyWithCookie(t, app, loginCookie(t, app), scopes)
	return key.Secret
}

func createAPIKeyWithCookie(t *testing.T, app *fiber.App, cookie string, scopes []string) auth.APIKey {
	t.Helper()

	body, err := json.Marshal(struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}{
		Name:   "test key",
		Scopes: scopes,
	})
	if err != nil {
		t.Fatalf("marshal api key fallo: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("crear api key fallo: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}

	var payload struct {
		Data auth.APIKey `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode api key fallo: %v", err)
	}
	return payload.Data
}

func uploadFile(t *testing.T, app *fiber.App, path string, filename string, content []byte, fields map[string]string, headers map[string]string) *http.Response {
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

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("upload fallo: %v", err)
	}

	return resp
}

func decodeMediaData(t *testing.T, reader io.Reader) media.File {
	t.Helper()

	var payload struct {
		Data media.File `json:"data"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		t.Fatalf("decode media fallo: %v", err)
	}
	return payload.Data
}

func decodeAPIKeyList(t *testing.T, reader io.Reader) []auth.APIKey {
	t.Helper()

	var payload struct {
		Data []auth.APIKey `json:"data"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		t.Fatalf("decode api key list fallo: %v", err)
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
