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
  <title>Ingresar | AceTransfer</title>
  <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
  <link rel="stylesheet" href="/assets/app.css?v=3" />
</head>
<body class="login-body">
  <form class="login-card" method="post" action="/login" id="loginForm">
    <div class="brand-container">
      <svg class="sd-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 6V18A2 2 0 0 0 6 20H18A2 2 0 0 0 20 18V9L15 4H6A2 2 0 0 0 4 6Z" />
        <path d="M9 10V6" />
        <path d="M12 10V6" />
        <path d="M15 10V6" />
      </svg>
      <strong class="brand">AceTransfer</strong>
    </div>
    <h1>Ingresa a tu cuenta</h1>
    <label><span>Email o usuario</span><input name="email" autocomplete="username" required /></label>
    <label><span>Contraseña</span><input name="password" type="password" autocomplete="current-password" required /></label>
    <button class="primary" type="submit">Ingresar</button>
    <p class="message error" id="message"></p>
  </form>
  <script>
    const form = document.querySelector("#loginForm");
    const message = document.querySelector("#message");
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      message.textContent = "";
      const response = await fetch("/login", { method: "POST", body: new FormData(form) });
      if (response.ok) { window.location.href = "/"; return; }
      const payload = await response.json().catch(() => ({}));
      message.textContent = payload.error?.message || "No se pudo iniciar sesion";
    });
  </script>
</body>
</html>`
}

// AppHTML devuelve la aplicacion autenticada.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Documento HTML completo del producto.
func AppHTML() string {
	return `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>AceTransfer</title>
  <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
  <link rel="stylesheet" href="/assets/app.css?v=3" />
</head>
<body>
  <main class="app-shell">
    <aside class="nav">
      <div class="brand-container">
        <svg class="sd-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 6V18A2 2 0 0 0 6 20H18A2 2 0 0 0 20 18V9L15 4H6A2 2 0 0 0 4 6Z" />
          <path d="M9 10V6" />
          <path d="M12 10V6" />
          <path d="M15 10V6" />
        </svg>
        <strong class="brand">AceTransfer</strong>
      </div>
      <div class="nav-menu">
        <button data-tab="send" class="nav-button active">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14M5 12h14"/></svg>
          <span>Enviar</span>
        </button>
        <button data-tab="links" class="nav-button">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          <span>Links</span>
        </button>
        <button data-tab="files" class="nav-button">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
          <span>Archivos</span>
        </button>
        <button data-tab="account" class="nav-button">
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          <span>Cuenta</span>
        </button>
        <button data-tab="admin" class="nav-button" id="adminTab" hidden>
          <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          <span>Admin</span>
        </button>
      </div>

      <div class="quota-widget">
        <div class="quota-info">
          <span>Almacenamiento</span>
          <span id="quotaText">Cargando...</span>
        </div>
        <div class="quota-progress"><span id="quotaMeter"></span></div>
      </div>

      <button class="nav-button logout" id="logoutButton">
        <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
        <span>Salir</span>
      </button>
    </aside>
    
    <section class="workspace">
      <header class="workspace-header">
        <h1 id="pageTitle">Enviar archivos</h1>
      </header>

      <section class="view active" id="sendView">
        <div class="send-layout">
          <form class="send-panel" id="transferForm">
            <label class="dropzone" id="dropzone">
              <input id="fileInput" name="files" type="file" multiple hidden />
              <div class="dropzone-content">
                <svg class="upload-cloud-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                  <polyline points="17 8 12 3 7 8" />
                  <line x1="12" y1="3" x2="12" y2="15" />
                </svg>
                <strong id="fileLabel">Añade tus archivos</strong>
                <small>Arrastra y suelta aquí o haz clic para buscarlos</small>
              </div>
            </label>

            <div class="selected-files-container" id="selectedFilesBox" hidden>
              <div class="selected-files-header">
                <span>Archivos seleccionados</span>
                <span class="selected-files-count" id="selectedFilesCount">0 archivos</span>
              </div>
              <div class="selected-files-list" id="selectedFilesList"></div>
            </div>

            <div class="form-grid">
              <label><span>Título del envío</span><input name="title" maxlength="160" placeholder="Ej: Documentos del proyecto" /></label>
              <label><span>Vence en (días)</span><input id="expiresDays" name="expires_days" type="number" min="1" max="365" placeholder="30" /></label>
            </div>
            
            <label class="switch"><input id="neverExpires" name="never_expires" value="true" type="checkbox" /><span>Link sin vencimiento</span></label>
            <label><span>Mensaje</span><textarea name="message" maxlength="600" placeholder="Escribe un mensaje opcional para los destinatarios..."></textarea></label>
            
            <div class="upload-progress-container" id="uploadProgress" hidden>
              <div class="gauge-wrapper">
                <svg class="progress-ring" width="120" height="120">
                  <circle class="progress-ring__background" stroke="var(--line)" stroke-width="8" fill="transparent" r="50" cx="60" cy="60"/>
                  <circle class="progress-ring__circle" stroke="url(#gauge-gradient)" stroke-width="8" stroke-linecap="round" fill="transparent" r="50" cx="60" cy="60" id="uploadBar"/>
                  <defs>
                    <linearGradient id="gauge-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                      <stop offset="0%" stop-color="var(--brand)" />
                      <stop offset="100%" stop-color="var(--accent)" />
                    </linearGradient>
                  </defs>
                </svg>
                <div class="progress-percentage" id="uploadText">0%</div>
              </div>
              <div class="upload-speed-info" id="uploadSpeedInfo">Subiendo archivos...</div>
            </div>

            <div class="actions">
              <button class="primary" id="sendButton" type="submit">Enviar archivos</button>
            </div>
            
            <p class="message" id="sendMessage"></p>
            <div class="result" id="resultBox" hidden></div>
          </form>
        </div>
      </section>

      <section class="view" id="linksView">
        <div class="toolbar">
          <h2>Enlaces compartidos</h2>
          <div class="filter-bar">
            <input type="text" id="searchLinks" placeholder="Buscar enlaces..." />
            <button id="refreshLinks" class="btn-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"/></svg>
            </button>
          </div>
        </div>
        <div class="list-container">
          <div class="list" id="linksList"></div>
        </div>
      </section>

      <section class="view" id="filesView">
        <div class="toolbar">
          <h2>Todos los archivos</h2>
          <div class="filter-bar">
            <input type="text" id="searchFiles" placeholder="Buscar archivos..." />
            <button id="refreshFiles" class="btn-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"/></svg>
            </button>
          </div>
        </div>
        <div class="list-container">
          <div class="list" id="filesList"></div>
        </div>
      </section>

      <section class="view" id="accountView">
        <div class="list" id="accountBox"></div>
        <section class="panel">
          <h2>API keys</h2>
          <form id="keyForm" class="inline-form">
            <input name="name" placeholder="Nombre de la key" required />
            <button class="primary">Crear Key</button>
          </form>
          <p class="message" id="keyMessage"></p>
          <div class="list" id="keysList"></div>
        </section>
      </section>

      <section class="view" id="adminView">
        <section class="panel">
          <h2>Crear nuevo usuario</h2>
          <form id="userForm" class="grid-form">
            <input name="email" placeholder="Email" required />
            <input name="name" placeholder="Nombre completo" />
            <input name="password" placeholder="Contraseña inicial" required />
            <input name="quota_gb" type="number" min="1" placeholder="Límite en GB (Default 10)" />
            <input name="share_ttl_days" type="number" min="0" placeholder="Vencimiento default (días, 0 sin fin)" />
            <button class="primary">Crear usuario</button>
          </form>
        </section>
        <div class="list" id="usersList"></div>
      </section>
    </section>
  </main>

  <div class="modal-overlay" id="fifoModal" hidden>
    <div class="modal">
      <div class="modal-header">
        <svg class="modal-icon warning" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/>
          <line x1="12" y1="9" x2="12" y2="13"/>
          <line x1="12" y1="17" x2="12.01" y2="17"/>
        </svg>
        <h2>Espacio Insuficiente</h2>
      </div>
      <div class="modal-body">
        <p>No tienes suficiente espacio libre en tu cuota. Para poder subir estos archivos, es necesario liberar espacio eliminando automáticamente los siguientes elementos más antiguos:</p>
        <div class="fifo-file-list" id="fifoFileList"></div>
        <p class="modal-subtext">¿Deseas autorizar la eliminación de estos archivos y continuar con la subida?</p>
      </div>
      <div class="modal-footer">
        <button class="btn-secondary" id="cancelFifoBtn">Cancelar</button>
        <button class="btn-primary" id="confirmFifoBtn">Confirmar y Subir</button>
      </div>
    </div>
  </div>

  <script src="/assets/app.js?v=5"></script>
</body>
</html>`
}

// ShareHTML devuelve la pagina publica de descarga.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Documento HTML completo para links publicos.
func ShareHTML() string {
	return `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Descargar | AceTransfer</title>
  <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
  <link rel="stylesheet" href="/assets/app.css?v=3" />
</head>
<body class="share-body">
  <main class="share-card">
    <div class="brand-container">
      <svg class="sd-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 6V18A2 2 0 0 0 6 20H18A2 2 0 0 0 20 18V9L15 4H6A2 2 0 0 0 4 6Z" />
        <path d="M9 10V6" />
        <path d="M12 10V6" />
        <path d="M15 10V6" />
      </svg>
      <strong class="brand">AceTransfer</strong>
    </div>
    <h1 id="shareTitle">Preparando descarga</h1>
    <p id="shareMeta">Cargando link</p>
    <div class="list" id="shareFiles"></div>
    <a class="primary download" id="downloadButton" href="#">Descargar todo (.ZIP)</a>
  </main>
  <script src="/assets/share.js?v=3"></script>
</body>
</html>`
}
