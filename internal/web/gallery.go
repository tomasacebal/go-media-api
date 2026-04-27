package web

// LoginHTML devuelve la pantalla de login.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Documento HTML completo para autenticar la web.
func LoginHTML() string {
	return `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Login | go-media-api</title>
  <style>
    :root {
      --bg: #eef3f0;
      --panel: #ffffff;
      --ink: #17211d;
      --muted: #66736d;
      --line: #d8e0dc;
      --brand: #0f766e;
      --danger: #b42318;
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 24px;
      color: var(--ink);
      background: var(--bg);
      font-family: "Segoe UI", sans-serif;
    }

    form {
      width: min(420px, 100%);
      display: grid;
      gap: 14px;
      padding: 24px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      box-shadow: 0 20px 50px rgba(23, 33, 29, 0.08);
    }

    h1 {
      margin: 0;
      font-size: 1.7rem;
      letter-spacing: 0;
    }

    label {
      display: grid;
      gap: 6px;
      color: var(--muted);
      font-size: 0.9rem;
      font-weight: 700;
    }

    input {
      width: 100%;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 11px 12px;
      color: var(--ink);
      font: inherit;
    }

    button {
      border: 0;
      border-radius: 8px;
      padding: 12px 14px;
      color: #fff;
      background: var(--brand);
      cursor: pointer;
      font: inherit;
      font-weight: 800;
    }

    .message {
      min-height: 20px;
      color: var(--danger);
      font-size: 0.92rem;
    }
  </style>
</head>
<body>
  <form method="post" action="/login" id="loginForm">
    <h1>go-media-api</h1>
    <label>
      Usuario
      <input name="username" autocomplete="username" required />
    </label>
    <label>
      Password
      <input name="password" type="password" autocomplete="current-password" required />
    </label>
    <button type="submit">Ingresar</button>
    <div class="message" id="message"></div>
  </form>
  <script>
    const form = document.querySelector("#loginForm");
    const message = document.querySelector("#message");
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      message.textContent = "";
      const response = await fetch("/login", {
        method: "POST",
        body: new FormData(form)
      });
      if (response.ok) {
        window.location.href = "/";
        return;
      }
      const payload = await response.json().catch(() => ({}));
      message.textContent = payload.error?.message || "No se pudo iniciar sesion";
    });
  </script>
</body>
</html>`
}

// GalleryHTML devuelve la interfaz web de upload, galeria y api keys.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Documento HTML completo para servir desde Fiber.
func GalleryHTML() string {
	return `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>go-media-api</title>
  <style>
    :root {
      --bg: #eef3f0;
      --panel: #ffffff;
      --ink: #17211d;
      --muted: #66736d;
      --line: #d8e0dc;
      --brand: #0f766e;
      --brand-strong: #115e59;
      --danger: #b42318;
      --soft: #e7f4f1;
      --shadow: 0 20px 50px rgba(23, 33, 29, 0.08);
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      min-height: 100vh;
      color: var(--ink);
      background: var(--bg);
      font-family: "Segoe UI", sans-serif;
    }

    button, input, select, textarea { font: inherit; }

    .shell {
      width: min(1180px, calc(100% - 32px));
      margin: 0 auto;
      padding: 24px 0 40px;
    }

    .topbar {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      margin-bottom: 18px;
    }

    h1, h2, h3 { margin: 0; letter-spacing: 0; }

    h1 { font-size: clamp(2rem, 7vw, 4rem); line-height: 0.95; }
    h2 { font-size: 1.2rem; }
    h3 { font-size: 1rem; }

    .status {
      color: var(--brand-strong);
      background: var(--soft);
      border: 1px solid #b8d9d2;
      border-radius: 999px;
      padding: 9px 14px;
      font-size: 0.92rem;
      text-align: center;
    }

    .layout {
      display: grid;
      grid-template-columns: 360px 1fr;
      gap: 18px;
      align-items: start;
    }

    .side {
      display: grid;
      gap: 14px;
      position: sticky;
      top: 16px;
    }

    .panel {
      background: rgba(255, 255, 255, 0.92);
      border: 1px solid var(--line);
      border-radius: 8px;
      box-shadow: var(--shadow);
    }

    .panel-inner { padding: 16px; }

    .dropzone {
      display: grid;
      place-items: center;
      min-height: 150px;
      border: 2px dashed #9ab5ad;
      border-radius: 8px;
      background: #f8fbf9;
      text-align: center;
      cursor: pointer;
      padding: 18px;
      transition: border-color 160ms ease, background 160ms ease;
    }

    .dropzone:hover, .dropzone.is-over {
      border-color: var(--brand);
      background: #edf8f5;
    }

    .dropzone strong { display: block; margin-bottom: 6px; overflow-wrap: anywhere; }
    .dropzone span { color: var(--muted); font-size: 0.9rem; }

    .field {
      display: grid;
      gap: 6px;
      margin-top: 12px;
    }

    .field label, .checkset legend {
      color: var(--muted);
      font-size: 0.82rem;
      font-weight: 800;
      text-transform: uppercase;
    }

    .field input, .field select, .field textarea {
      width: 100%;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 10px 11px;
      background: #fff;
      color: var(--ink);
    }

    .field textarea {
      resize: vertical;
      min-height: 78px;
    }

    .checkset {
      display: grid;
      gap: 8px;
      margin: 12px 0 0;
      padding: 0;
      border: 0;
    }

    .checks {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }

    .checks label {
      display: flex;
      gap: 6px;
      align-items: center;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 8px 10px;
      background: #fff;
    }

    .actions {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      margin-top: 14px;
    }

    .button {
      border: 0;
      border-radius: 8px;
      padding: 10px 13px;
      cursor: pointer;
      color: #fff;
      background: var(--brand);
      font-weight: 800;
    }

    .button.secondary {
      color: var(--brand-strong);
      background: var(--soft);
    }

    .button.danger { background: var(--danger); }
    .button:disabled { cursor: not-allowed; opacity: 0.6; }

    .toolbar {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      align-items: center;
      padding: 14px;
      margin-bottom: 14px;
    }

    .toolbar select {
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 9px 10px;
      background: #fff;
      color: var(--ink);
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
      gap: 14px;
    }

    .asset, .key-row {
      overflow: hidden;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
    }

    .thumb {
      display: grid;
      place-items: center;
      aspect-ratio: 4 / 3;
      background: #e8eee9;
      overflow: hidden;
      color: var(--muted);
      text-decoration: none;
      font-weight: 900;
    }

    .thumb img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }

    .meta, .key-row { padding: 12px; }
    .meta strong, .key-row strong { display: block; overflow-wrap: anywhere; margin-bottom: 6px; }
    .meta p, .key-row p { margin: 0 0 10px; color: var(--muted); font-size: 0.9rem; overflow-wrap: anywhere; }

    .asset-actions, .key-actions {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .asset-actions a, .asset-actions button, .key-actions button {
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 7px 9px;
      background: #fff;
      color: var(--ink);
      text-decoration: none;
      cursor: pointer;
      font-size: 0.86rem;
    }

    .empty {
      display: grid;
      place-items: center;
      min-height: 220px;
      border: 1px dashed #9ab5ad;
      border-radius: 8px;
      color: var(--muted);
      text-align: center;
      padding: 24px;
    }

    .message {
      min-height: 22px;
      margin-top: 12px;
      color: var(--muted);
      font-size: 0.92rem;
      overflow-wrap: anywhere;
    }

    .message.error { color: var(--danger); }
    .keys { display: grid; gap: 10px; margin-top: 12px; }

    @media (max-width: 860px) {
      .topbar { align-items: stretch; flex-direction: column; }
      .layout { grid-template-columns: 1fr; }
      .side { position: static; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div>
        <h1>Media Gallery</h1>
        <div class="status" id="status">Listo</div>
      </div>
      <button class="button secondary" id="logoutButton" type="button">Salir</button>
    </header>

    <section class="layout">
      <aside class="side">
        <form class="panel panel-inner" id="uploadForm">
          <h2>Subir archivo</h2>
          <label class="dropzone" id="dropzone">
            <input id="fileInput" name="file" type="file" hidden required />
            <span>
              <strong id="fileLabel">Elegir archivo</strong>
              Cualquier tipo de archivo
            </span>
          </label>

          <div class="field">
            <label for="visibility">Visibilidad</label>
            <select id="visibility" name="visibility">
              <option value="public">Public</option>
              <option value="private">Private</option>
            </select>
          </div>

          <div class="field">
            <label for="title">Titulo</label>
            <input id="title" name="title" maxlength="160" placeholder="Nombre visible" />
          </div>

          <div class="field">
            <label for="category">Categoria</label>
            <input id="category" name="category" maxlength="120" placeholder="Ej: documentos" />
          </div>

          <div class="field">
            <label for="description">Descripcion</label>
            <textarea id="description" name="description" maxlength="600" placeholder="Detalle breve"></textarea>
          </div>

          <div class="actions">
            <button class="button" id="uploadButton" type="submit">Subir</button>
            <button class="button secondary" id="refreshButton" type="button">Refrescar</button>
          </div>
          <div class="message" id="message"></div>
        </form>

        <section class="panel panel-inner">
          <h2>Api keys</h2>
          <form id="keyForm">
            <div class="field">
              <label for="keyName">Nombre</label>
              <input id="keyName" name="name" maxlength="120" placeholder="Integracion externa" required />
            </div>
            <fieldset class="checkset">
              <legend>Scopes</legend>
              <div class="checks">
                <label><input type="checkbox" name="scope" value="read" checked /> read</label>
                <label><input type="checkbox" name="scope" value="write" checked /> write</label>
                <label><input type="checkbox" name="scope" value="delete" /> delete</label>
              </div>
            </fieldset>
            <div class="actions">
              <button class="button" type="submit">Crear key</button>
            </div>
          </form>
          <div class="message" id="keyMessage"></div>
          <div class="keys" id="keys"></div>
        </section>
      </aside>

      <section>
        <div class="panel toolbar">
          <h2>Assets</h2>
          <select id="filter">
            <option value="all">Todos</option>
            <option value="image">Imagenes</option>
            <option value="public">Publicos</option>
            <option value="private">Privados</option>
          </select>
        </div>
        <div class="grid" id="gallery"></div>
      </section>
    </section>
  </main>

  <script>
    const state = { files: [], keys: [], filter: "all" };
    const form = document.querySelector("#uploadForm");
    const fileInput = document.querySelector("#fileInput");
    const fileLabel = document.querySelector("#fileLabel");
    const dropzone = document.querySelector("#dropzone");
    const uploadButton = document.querySelector("#uploadButton");
    const refreshButton = document.querySelector("#refreshButton");
    const logoutButton = document.querySelector("#logoutButton");
    const gallery = document.querySelector("#gallery");
    const message = document.querySelector("#message");
    const status = document.querySelector("#status");
    const filter = document.querySelector("#filter");
    const keyForm = document.querySelector("#keyForm");
    const keyName = document.querySelector("#keyName");
    const keyMessage = document.querySelector("#keyMessage");
    const keys = document.querySelector("#keys");

    fileInput.addEventListener("change", () => {
      fileLabel.textContent = fileInput.files[0] ? fileInput.files[0].name : "Elegir archivo";
    });

    dropzone.addEventListener("dragover", (event) => {
      event.preventDefault();
      dropzone.classList.add("is-over");
    });

    dropzone.addEventListener("dragleave", () => dropzone.classList.remove("is-over"));

    dropzone.addEventListener("drop", (event) => {
      event.preventDefault();
      dropzone.classList.remove("is-over");
      if (event.dataTransfer.files.length > 0) {
        fileInput.files = event.dataTransfer.files;
        fileLabel.textContent = event.dataTransfer.files[0].name;
      }
    });

    filter.addEventListener("change", () => {
      state.filter = filter.value;
      renderFiles();
    });

    refreshButton.addEventListener("click", () => {
      loadFiles();
      loadKeys();
    });

    logoutButton.addEventListener("click", async () => {
      await fetch("/logout", { method: "POST" });
      window.location.href = "/login";
    });

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      if (!fileInput.files[0]) {
        setMessage("Selecciona un archivo.", true);
        return;
      }

      uploadButton.disabled = true;
      setStatus("Subiendo");
      setMessage("");

      try {
        const response = await fetch("/web/media/upload", {
          method: "POST",
          body: new FormData(form)
        });
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.error?.message || "No se pudo subir el archivo");
        }

        form.reset();
        fileLabel.textContent = "Elegir archivo";
        setMessage("Archivo subido.");
        await loadFiles();
      } catch (error) {
        setMessage(error.message, true);
        setStatus("Error");
      } finally {
        uploadButton.disabled = false;
      }
    });

    keyForm.addEventListener("submit", async (event) => {
      event.preventDefault();
      const scopes = Array.from(keyForm.querySelectorAll("input[name='scope']:checked")).map((item) => item.value);
      if (scopes.length === 0) {
        setKeyMessage("Selecciona al menos un scope.", true);
        return;
      }

      try {
        const response = await fetch("/api/v1/api-keys/", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            name: keyName.value,
            scopes
          })
        });
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.error?.message || "No se pudo crear la api key");
        }
        keyForm.reset();
        keyForm.querySelector("input[value='read']").checked = true;
        keyForm.querySelector("input[value='write']").checked = true;
        setKeyMessage("Key creada: " + payload.data.secret);
        await loadKeys();
      } catch (error) {
        setKeyMessage(error.message, true);
      }
    });

    async function loadFiles() {
      setStatus("Cargando");
      try {
        const response = await fetch("/api/v1/media?limit=100");
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.error?.message || "No se pudo listar la galeria");
        }
        state.files = payload.data || [];
        renderFiles();
        setStatus(state.files.length + " assets");
      } catch (error) {
        gallery.innerHTML = emptyState(error.message);
        setStatus("Error");
      }
    }

    async function loadKeys() {
      try {
        const response = await fetch("/api/v1/api-keys/");
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.error?.message || "No se pudieron listar las api keys");
        }
        state.keys = payload.data || [];
        renderKeys();
      } catch (error) {
        keys.innerHTML = emptyState(error.message);
      }
    }

    function renderFiles() {
      const files = state.files.filter(matchesFilter);
      if (files.length === 0) {
        gallery.innerHTML = emptyState("No hay assets para mostrar.");
        return;
      }

      gallery.innerHTML = files.map((file) => {
        const title = escapeHTML(file.title || file.original_name || file.id);
        const subtitle = escapeHTML([file.mime_type, formatBytes(file.size_bytes), file.visibility].filter(Boolean).join(" - "));
        const description = escapeHTML(file.description || file.category || file.id);
        const downloadURL = "/api/v1/media/" + encodeURIComponent(file.id) + "/download";
        const preview = file.mime_type.startsWith("image/")
          ? "<img src=\"" + downloadURL + "\" alt=\"" + title + "\" loading=\"lazy\" />"
          : "<span>" + escapeHTML(file.extension || "file") + "</span>";

        return ""
          + "<article class=\"asset\">"
          + "<a class=\"thumb\" href=\"" + downloadURL + "\" target=\"_blank\" rel=\"noreferrer\">" + preview + "</a>"
          + "<div class=\"meta\">"
          + "<strong>" + title + "</strong>"
          + "<p>" + subtitle + "</p>"
          + "<p>" + description + "</p>"
          + "<div class=\"asset-actions\">"
          + "<a href=\"" + downloadURL + "\" target=\"_blank\" rel=\"noreferrer\">Abrir</a>"
          + "<button type=\"button\" data-delete=\"" + escapeHTML(file.id) + "\">Borrar</button>"
          + "</div>"
          + "</div>"
          + "</article>";
      }).join("");
    }

    function renderKeys() {
      if (state.keys.length === 0) {
        keys.innerHTML = emptyState("No hay api keys activas.");
        return;
      }

      keys.innerHTML = state.keys.map((key) => {
        return ""
          + "<article class=\"key-row\">"
          + "<strong>" + escapeHTML(key.name) + "</strong>"
          + "<p>" + escapeHTML(key.key_prefix + "..." + " - " + key.scopes.join(", ")) + "</p>"
          + "<div class=\"key-actions\">"
          + "<button type=\"button\" data-revoke=\"" + escapeHTML(key.id) + "\">Revocar</button>"
          + "</div>"
          + "</article>";
      }).join("");
    }

    gallery.addEventListener("click", async (event) => {
      const button = event.target.closest("[data-delete]");
      if (!button) {
        return;
      }

      const id = button.getAttribute("data-delete");
      button.disabled = true;
      setStatus("Borrando");
      try {
        const response = await fetch("/api/v1/media/" + encodeURIComponent(id), { method: "DELETE" });
        if (!response.ok) {
          const payload = await response.json();
          throw new Error(payload.error?.message || "No se pudo borrar");
        }
        state.files = state.files.filter((file) => file.id !== id);
        renderFiles();
        setStatus(state.files.length + " assets");
      } catch (error) {
        setMessage(error.message, true);
        setStatus("Error");
        button.disabled = false;
      }
    });

    keys.addEventListener("click", async (event) => {
      const button = event.target.closest("[data-revoke]");
      if (!button) {
        return;
      }

      button.disabled = true;
      try {
        const response = await fetch("/api/v1/api-keys/" + encodeURIComponent(button.getAttribute("data-revoke")), {
          method: "DELETE"
        });
        if (!response.ok) {
          const payload = await response.json();
          throw new Error(payload.error?.message || "No se pudo revocar");
        }
        await loadKeys();
      } catch (error) {
        setKeyMessage(error.message, true);
        button.disabled = false;
      }
    });

    function matchesFilter(file) {
      if (state.filter === "all") {
        return true;
      }
      if (state.filter === "image") {
        return file.mime_type.startsWith("image/");
      }
      return file.visibility === state.filter;
    }

    function setMessage(value, isError = false) {
      message.textContent = value;
      message.classList.toggle("error", isError);
    }

    function setKeyMessage(value, isError = false) {
      keyMessage.textContent = value;
      keyMessage.classList.toggle("error", isError);
    }

    function setStatus(value) {
      status.textContent = value;
    }

    function formatBytes(value) {
      if (!value) {
        return "0 B";
      }
      const units = ["B", "KB", "MB", "GB", "TB"];
      let size = value;
      let unit = 0;
      while (size >= 1024 && unit < units.length - 1) {
        size = size / 1024;
        unit++;
      }
      return size.toFixed(size >= 10 || unit === 0 ? 0 : 1) + " " + units[unit];
    }

    function escapeHTML(value) {
      return String(value).replace(/[&<>"']/g, (char) => ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        "\"": "&quot;",
        "'": "&#039;"
      })[char]);
    }

    function emptyState(text) {
      return "<div class=\"empty\">" + escapeHTML(text) + "</div>";
    }

    loadFiles();
    loadKeys();
  </script>
</body>
</html>`
}
