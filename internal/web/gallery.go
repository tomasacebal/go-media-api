package web

// GalleryHTML devuelve la interfaz web de upload y galeria.
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
      color-scheme: light;
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

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      min-height: 100vh;
      font-family: Georgia, "Times New Roman", serif;
      color: var(--ink);
      background:
        linear-gradient(135deg, rgba(15, 118, 110, 0.14), transparent 34%),
        linear-gradient(315deg, rgba(151, 126, 68, 0.12), transparent 30%),
        var(--bg);
    }

    button,
    input,
    select,
    textarea {
      font: inherit;
    }

    .shell {
      width: min(1180px, calc(100% - 32px));
      margin: 0 auto;
      padding: 28px 0 40px;
    }

    .topbar {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: flex-end;
      margin-bottom: 22px;
    }

    h1 {
      margin: 0;
      font-size: clamp(2rem, 7vw, 4.6rem);
      line-height: 0.92;
      letter-spacing: 0;
    }

    .lede {
      max-width: 560px;
      margin: 10px 0 0;
      color: var(--muted);
      font-family: "Segoe UI", sans-serif;
      font-size: 1rem;
    }

    .status {
      min-width: 190px;
      color: var(--brand-strong);
      background: var(--soft);
      border: 1px solid #b8d9d2;
      border-radius: 999px;
      padding: 9px 14px;
      font-family: "Segoe UI", sans-serif;
      font-size: 0.92rem;
      text-align: center;
    }

    .layout {
      display: grid;
      grid-template-columns: 360px 1fr;
      gap: 18px;
      align-items: start;
    }

    .panel {
      background: rgba(255, 255, 255, 0.86);
      border: 1px solid var(--line);
      border-radius: 8px;
      box-shadow: var(--shadow);
    }

    .upload {
      position: sticky;
      top: 18px;
      padding: 18px;
    }

    .dropzone {
      display: grid;
      place-items: center;
      min-height: 190px;
      border: 2px dashed #9ab5ad;
      border-radius: 8px;
      background: #f8fbf9;
      text-align: center;
      cursor: pointer;
      padding: 18px;
      transition: border-color 160ms ease, background 160ms ease;
    }

    .dropzone:hover,
    .dropzone.is-over {
      border-color: var(--brand);
      background: #edf8f5;
    }

    .dropzone strong {
      display: block;
      margin-bottom: 6px;
      font-size: 1.1rem;
    }

    .dropzone span {
      color: var(--muted);
      font-family: "Segoe UI", sans-serif;
      font-size: 0.92rem;
    }

    .field {
      display: grid;
      gap: 6px;
      margin-top: 12px;
      font-family: "Segoe UI", sans-serif;
    }

    .field label {
      color: var(--muted);
      font-size: 0.84rem;
      font-weight: 700;
      text-transform: uppercase;
    }

    .field input,
    .field select,
    .field textarea {
      width: 100%;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 10px 11px;
      background: #fff;
      color: var(--ink);
    }

    .field textarea {
      resize: vertical;
      min-height: 86px;
    }

    .actions {
      display: flex;
      gap: 10px;
      margin-top: 14px;
    }

    .button {
      border: 0;
      border-radius: 8px;
      padding: 11px 14px;
      cursor: pointer;
      color: #fff;
      background: var(--brand);
      font-family: "Segoe UI", sans-serif;
      font-weight: 800;
      transition: transform 140ms ease, background 140ms ease;
    }

    .button:hover {
      background: var(--brand-strong);
      transform: translateY(-1px);
    }

    .button.secondary {
      color: var(--brand-strong);
      background: var(--soft);
    }

    .button.danger {
      background: var(--danger);
    }

    .button:disabled {
      cursor: not-allowed;
      opacity: 0.58;
      transform: none;
    }

    .toolbar {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      align-items: center;
      padding: 14px;
      margin-bottom: 14px;
    }

    .toolbar h2 {
      margin: 0;
      font-size: 1.25rem;
    }

    .toolbar select {
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 9px 10px;
      background: #fff;
      color: var(--ink);
      font-family: "Segoe UI", sans-serif;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
      gap: 14px;
    }

    .asset {
      overflow: hidden;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      box-shadow: 0 12px 26px rgba(23, 33, 29, 0.07);
    }

    .thumb {
      display: grid;
      place-items: center;
      aspect-ratio: 4 / 3;
      background: #e8eee9;
      overflow: hidden;
    }

    .thumb img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }

    .pdf {
      width: 76px;
      height: 96px;
      border-radius: 6px;
      background: #fff;
      border: 1px solid #cfd8d3;
      display: grid;
      place-items: center;
      color: var(--danger);
      font-family: "Segoe UI", sans-serif;
      font-weight: 900;
      box-shadow: 0 10px 18px rgba(23, 33, 29, 0.1);
    }

    .meta {
      padding: 12px;
      font-family: "Segoe UI", sans-serif;
    }

    .meta strong {
      display: block;
      overflow-wrap: anywhere;
      margin-bottom: 6px;
    }

    .meta p {
      margin: 0 0 10px;
      color: var(--muted);
      font-size: 0.9rem;
      overflow-wrap: anywhere;
    }

    .asset-actions {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .asset-actions a,
    .asset-actions button {
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 7px 9px;
      background: #fff;
      color: var(--ink);
      text-decoration: none;
      cursor: pointer;
      font-family: "Segoe UI", sans-serif;
      font-size: 0.86rem;
    }

    .empty {
      display: grid;
      place-items: center;
      min-height: 260px;
      border: 1px dashed #9ab5ad;
      border-radius: 8px;
      color: var(--muted);
      font-family: "Segoe UI", sans-serif;
      text-align: center;
      padding: 24px;
    }

    .message {
      min-height: 22px;
      margin-top: 12px;
      color: var(--muted);
      font-family: "Segoe UI", sans-serif;
      font-size: 0.92rem;
    }

    .message.error {
      color: var(--danger);
    }

    @media (max-width: 820px) {
      .topbar {
        align-items: stretch;
        flex-direction: column;
      }

      .layout {
        grid-template-columns: 1fr;
      }

      .upload {
        position: static;
      }
    }
  </style>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div>
        <h1>Media Gallery</h1>
        <p class="lede">Subi imagenes y PDFs, revisa metadata y administra assets publicos o privados desde el navegador.</p>
      </div>
      <div class="status" id="status">Listo</div>
    </header>

    <section class="layout">
      <form class="panel upload" id="uploadForm">
        <label class="dropzone" id="dropzone">
          <input id="fileInput" name="file" type="file" accept=".jpg,.jpeg,.png,.webp,.pdf" hidden required />
          <span>
            <strong id="fileLabel">Elegir archivo</strong>
            JPG, JPEG, PNG, WEBP o PDF
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

      <section>
        <div class="panel toolbar">
          <h2>Assets</h2>
          <select id="filter">
            <option value="all">Todos</option>
            <option value="image">Imagenes</option>
            <option value="pdf">PDF</option>
            <option value="public">Publicos</option>
            <option value="private">Privados</option>
          </select>
        </div>
        <div class="grid" id="gallery"></div>
      </section>
    </section>
  </main>

  <script>
    const state = {
      files: [],
      filter: "all"
    };

    const form = document.querySelector("#uploadForm");
    const fileInput = document.querySelector("#fileInput");
    const fileLabel = document.querySelector("#fileLabel");
    const dropzone = document.querySelector("#dropzone");
    const uploadButton = document.querySelector("#uploadButton");
    const refreshButton = document.querySelector("#refreshButton");
    const gallery = document.querySelector("#gallery");
    const message = document.querySelector("#message");
    const status = document.querySelector("#status");
    const filter = document.querySelector("#filter");

    fileInput.addEventListener("change", () => {
      fileLabel.textContent = fileInput.files[0] ? fileInput.files[0].name : "Elegir archivo";
    });

    dropzone.addEventListener("dragover", (event) => {
      event.preventDefault();
      dropzone.classList.add("is-over");
    });

    dropzone.addEventListener("dragleave", () => {
      dropzone.classList.remove("is-over");
    });

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
      render();
    });

    refreshButton.addEventListener("click", () => {
      loadFiles();
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
        const response = await fetch("/api/v1/media/upload", {
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

    async function loadFiles() {
      setStatus("Cargando");
      try {
        const response = await fetch("/api/v1/media?limit=100");
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.error?.message || "No se pudo listar la galeria");
        }
        state.files = payload.data || [];
        render();
        setStatus(state.files.length + " assets");
      } catch (error) {
        gallery.innerHTML = emptyState(error.message);
        setStatus("Error");
      }
    }

    function render() {
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
          : "<div class=\"pdf\">PDF</div>";

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

    gallery.addEventListener("click", async (event) => {
      const button = event.target.closest("[data-delete]");
      if (!button) {
        return;
      }

      const id = button.getAttribute("data-delete");
      button.disabled = true;
      setStatus("Borrando");
      try {
        const response = await fetch("/api/v1/media/" + encodeURIComponent(id), {
          method: "DELETE"
        });
        if (!response.ok) {
          const payload = await response.json();
          throw new Error(payload.error?.message || "No se pudo borrar");
        }
        state.files = state.files.filter((file) => file.id !== id);
        render();
        setStatus(state.files.length + " assets");
      } catch (error) {
        setMessage(error.message, true);
        setStatus("Error");
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
      if (state.filter === "pdf") {
        return file.mime_type === "application/pdf";
      }
      return file.visibility === state.filter;
    }

    function setMessage(value, isError = false) {
      message.textContent = value;
      message.classList.toggle("error", isError);
    }

    function setStatus(value) {
      status.textContent = value;
    }

    function formatBytes(value) {
      if (!value) {
        return "0 B";
      }
      const units = ["B", "KB", "MB", "GB"];
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
  </script>
</body>
</html>`
}
