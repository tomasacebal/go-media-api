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
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	resp := doRequest(t, app, httptest.NewRequest(http.MethodGet, "/", nil))
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusSeeOther {
		t.Fatalf("status esperado 303, recibido %d", resp.StatusCode)
	}
	if location := resp.Header.Get("Location"); location != "/login" {
		t.Fatalf("location esperada /login, recibida %s", location)
	}
}

func TestLoginAndMe(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	badResp := doRequest(t, app, loginRequest("admin@test.local", "bad-password"))
	defer badResp.Body.Close()
	if badResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status esperado 401, recibido %d", badResp.StatusCode)
	}

	cookie := loginCookie(t, app, "admin@test.local", "secret-password")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Cookie", cookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}
	user := decodeData[auth.SessionUser](t, resp.Body)
	if user.Role != auth.RoleAdmin || user.Email != "admin@test.local" {
		t.Fatalf("usuario inesperado: %+v", user)
	}
}

func TestAdminCreatesUserAndUserCannotReadAdminFile(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	adminCookie := loginCookie(t, app, "admin@test.local", "secret-password")
	user := createUser(t, app, adminCookie, "user@test.local", "user-password")
	if user.Role != auth.RoleUser {
		t.Fatalf("rol esperado user, recibido %s", user.Role)
	}

	adminFile := uploadSingleFile(t, app, "/api/v1/files/", "admin.txt", []byte("admin"), nil, map[string]string{"Cookie": adminCookie}).Files[0]
	userCookie := loginCookie(t, app, "user@test.local", "user-password")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+adminFile.ID, nil)
	req.Header.Set("Cookie", userCookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status esperado 403, recibido %d", resp.StatusCode)
	}
}

func TestAPIKeyIsScopedToOwner(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	adminCookie := loginCookie(t, app, "admin@test.local", "secret-password")
	_ = createUser(t, app, adminCookie, "user@test.local", "user-password")
	adminFile := uploadSingleFile(t, app, "/api/v1/files/", "admin.txt", []byte("admin"), nil, map[string]string{"Cookie": adminCookie}).Files[0]

	userCookie := loginCookie(t, app, "user@test.local", "user-password")
	key := createAPIKeyWithCookie(t, app, userCookie, []string{auth.ScopeRead})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+adminFile.ID, nil)
	req.Header.Set("X-API-Key", key.Secret)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status esperado 403, recibido %d", resp.StatusCode)
	}
}

func TestTransferCreatesShortShareAndPublicDownload(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	cookie := loginCookie(t, app, "admin@test.local", "secret-password")
	result := uploadTransfer(t, app, cookie, false, false, []namedContent{
		{name: "a.txt", body: []byte("uno")},
		{name: "b.txt", body: []byte("dos")},
	})
	if len(result.Share.Code) != 8 {
		t.Fatalf("codigo corto esperado de 8 caracteres, recibido %s", result.Share.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/shares/"+result.Share.Code, nil)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}
	detail := decodeData[media.ShareDetail](t, resp.Body)
	if len(detail.Files) != 2 {
		t.Fatalf("archivos esperados 2, recibidos %d", len(detail.Files))
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/s/"+result.Share.Code+"/download", nil)
	downloadResp := doRequest(t, app, downloadReq)
	defer downloadResp.Body.Close()
	if downloadResp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", downloadResp.StatusCode)
	}
	if got := downloadResp.Header.Get("Content-Type"); !strings.Contains(got, "application/zip") {
		t.Fatalf("content-type zip esperado, recibido %s", got)
	}
}

func TestTransferCanCreateNeverExpiringShare(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	cookie := loginCookie(t, app, "admin@test.local", "secret-password")
	result := uploadTransfer(t, app, cookie, false, true, []namedContent{{name: "a.txt", body: []byte("uno")}})
	if !result.Share.NeverExpires {
		t.Fatal("se esperaba link sin vencimiento")
	}
	if result.Share.ExpiresAt != nil {
		t.Fatalf("expires_at esperado nil, recibido %v", result.Share.ExpiresAt)
	}
}

func TestAdminCanUpdateUserQuotaAndShareTTL(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	adminCookie := loginCookie(t, app, "admin@test.local", "secret-password")
	user := createUser(t, app, adminCookie, "user@test.local", "user-password")
	body := mustJSON(t, map[string]interface{}{
		"quota_bytes":    int64(20),
		"share_ttl_days": 0,
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/"+user.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", adminCookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status esperado 200, recibido %d", resp.StatusCode)
	}
	updated := decodeData[auth.User](t, resp.Body)
	if updated.QuotaBytes != 20 || updated.ShareTTLDays != 0 {
		t.Fatalf("usuario actualizado inesperado: %+v", updated)
	}

	userCookie := loginCookie(t, app, "user@test.local", "user-password")
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	meReq.Header.Set("Cookie", userCookie)
	meResp := doRequest(t, app, meReq)
	defer meResp.Body.Close()
	me := decodeData[auth.SessionUser](t, meResp.Body)
	if me.QuotaBytes != 20 || me.ShareTTLDays != 0 {
		t.Fatalf("me inesperado: %+v", me)
	}
}

func TestDownloadUsesAttachmentDisposition(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 1024)
	defer cleanup()

	cookie := loginCookie(t, app, "admin@test.local", "secret-password")
	file := uploadSingleFile(t, app, "/api/v1/files/", "download.txt", []byte("contenido"), nil, map[string]string{"Cookie": cookie}).Files[0]
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+file.ID+"/download", nil)
	req.Header.Set("Cookie", cookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Disposition"), "attachment") {
		t.Fatalf("content-disposition esperado attachment, recibido %s", resp.Header.Get("Content-Disposition"))
	}
}

func TestQuotaFIFOPreviewAndConfirm(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 10)
	defer cleanup()

	cookie := loginCookie(t, app, "admin@test.local", "secret-password")
	oldFile := uploadSingleFile(t, app, "/api/v1/files/", "old.txt", []byte("123456"), nil, map[string]string{"Cookie": cookie}).Files[0]
	share := createShare(t, app, cookie, media.ShareTargetFile, oldFile.ID)

	previewResp := uploadFileRaw(t, app, "/api/v1/files/", []namedContent{{name: "new.txt", body: []byte("1234567")}}, nil, map[string]string{"Cookie": cookie})
	defer previewResp.Body.Close()
	if previewResp.StatusCode != fiber.StatusConflict {
		t.Fatalf("status esperado 409, recibido %d", previewResp.StatusCode)
	}

	confirmed := uploadSingleFile(t, app, "/api/v1/files/", "new.txt", []byte("1234567"), map[string]string{"confirm_fifo": "true"}, map[string]string{"Cookie": cookie})
	if len(confirmed.DeletedFiles) != 1 || confirmed.DeletedFiles[0].ID != oldFile.ID {
		t.Fatalf("fifo no borro el archivo viejo: %+v", confirmed.DeletedFiles)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/shares/"+share.Code, nil)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status esperado 404, recibido %d", resp.StatusCode)
	}
}

func TestFileLargerThanQuotaIsRejected(t *testing.T) {
	app, cleanup := newTestApp(t, 0, 5)
	defer cleanup()

	cookie := loginCookie(t, app, "admin@test.local", "secret-password")
	resp := uploadFileRaw(t, app, "/api/v1/files/", []namedContent{{name: "huge.txt", body: []byte("123456")}}, nil, map[string]string{"Cookie": cookie})
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusRequestEntityTooLarge {
		t.Fatalf("status esperado 413, recibido %d", resp.StatusCode)
	}
}

func newTestApp(t *testing.T, maxUploadBytes int64, quotaBytes int64) (*fiber.App, func()) {
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
			AdminUsername:   "admin",
			AdminEmail:      "admin@test.local",
			AdminName:       "Admin",
			AdminPassword:   "secret-password",
			SessionSecret:   "test-secret-with-at-least-thirty-two-chars",
			SessionTTLHours: 12,
		},
		Product: config.ProductConfig{
			DefaultQuotaBytes: quotaBytes,
			ShareTTLDays:      30,
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

func loginRequest(email string, password string) *http.Request {
	body := strings.NewReader("email=" + email + "&password=" + password)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	return req
}

func loginCookie(t *testing.T, app *fiber.App, email string, password string) string {
	t.Helper()

	resp := doRequest(t, app, loginRequest(email, password))
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

func createUser(t *testing.T, app *fiber.App, cookie string, email string, password string) auth.User {
	t.Helper()

	body := mustJSON(t, map[string]interface{}{
		"email":    email,
		"name":     email,
		"password": password,
		"role":     auth.RoleUser,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
	return decodeData[auth.User](t, resp.Body)
}

func createAPIKeyWithCookie(t *testing.T, app *fiber.App, cookie string, scopes []string) auth.APIKey {
	t.Helper()

	body := mustJSON(t, map[string]interface{}{"name": "test key", "scopes": scopes})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
	return decodeData[auth.APIKey](t, resp.Body)
}

func createShare(t *testing.T, app *fiber.App, cookie string, targetType string, targetID string) media.Share {
	t.Helper()

	body := mustJSON(t, map[string]interface{}{"target_type": targetType, "target_id": targetID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shares/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	resp := doRequest(t, app, req)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
	return decodeData[media.Share](t, resp.Body)
}

func uploadSingleFile(t *testing.T, app *fiber.App, path string, filename string, content []byte, fields map[string]string, headers map[string]string) media.UploadResult {
	t.Helper()

	resp := uploadFileRaw(t, app, path, []namedContent{{name: filename, body: content}}, fields, headers)
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
	return decodeData[media.UploadResult](t, resp.Body)
}

func uploadTransfer(t *testing.T, app *fiber.App, cookie string, confirm bool, neverExpires bool, files []namedContent) media.TransferResult {
	t.Helper()

	fields := map[string]string{"title": "Envio test"}
	if confirm {
		fields["confirm_fifo"] = "true"
	}
	if neverExpires {
		fields["never_expires"] = "true"
	}
	resp := uploadFileRaw(t, app, "/api/v1/transfers/", files, fields, map[string]string{"Cookie": cookie})
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status esperado 201, recibido %d", resp.StatusCode)
	}
	return decodeData[media.TransferResult](t, resp.Body)
}

func uploadFileRaw(t *testing.T, app *fiber.App, path string, files []namedContent, fields map[string]string, headers map[string]string) *http.Response {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, file := range files {
		part, err := writer.CreateFormFile("files", file.name)
		if err != nil {
			t.Fatalf("crear form file fallo: %v", err)
		}
		if _, err := part.Write(file.body); err != nil {
			t.Fatalf("escribir form file fallo: %v", err)
		}
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
	return doRequest(t, app, req)
}

func doRequest(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request fallo: %v", err)
	}
	return resp
}

func decodeData[T interface{}](t *testing.T, reader io.Reader) T {
	t.Helper()

	var payload struct {
		Data T `json:"data"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		t.Fatalf("decode fallo: %v", err)
	}
	return payload.Data
}

func mustJSON(t *testing.T, value interface{}) []byte {
	t.Helper()

	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fallo: %v", err)
	}
	return body
}

type namedContent struct {
	name string
	body []byte
}
