package web

// AppCSS devuelve estilos compartidos.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - CSS de la interfaz.
func AppCSS() string {
	return `@import url('https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&display=swap');

:root {
  --bg-app: #f4f6fa;
  --panel-bg: rgba(255, 255, 255, 0.85);
  --panel-border: rgba(226, 232, 240, 0.8);
  --ink-primary: #0f172a;
  --ink-secondary: #475569;
  --ink-muted: #94a3b8;
  --line: #e2e8f0;
  --brand: #4f46e5;
  --brand-hover: #4338ca;
  --brand-light: #e0e7ff;
  --accent: #06b6d4;
  --accent-light: #ecfeff;
  --danger: #ef4444;
  --danger-hover: #dc2626;
  --danger-light: #fee2e2;
  --success: #10b981;
  --success-light: #d1fae5;
  --warn: #f59e0b;
  --warn-light: #fef3c7;
  
  --font-sans: 'Outfit', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 12px -1px rgba(0, 0, 0, 0.05), 0 2px 6px -2px rgba(0, 0, 0, 0.05);
  --shadow-lg: 0 10px 25px -3px rgba(0, 0, 0, 0.05), 0 4px 12px -4px rgba(0, 0, 0, 0.05);
  --shadow-xl: 0 20px 30px -5px rgba(79, 70, 229, 0.08), 0 8px 16px -6px rgba(0, 0, 0, 0.04);
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-xl: 24px;
  --radius-full: 9999px;
  --transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
  font-family: var(--font-sans);
}

[hidden] {
  display: none !important;
}

body {
  background-color: var(--bg-app);
  color: var(--ink-primary);
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  position: relative;
  overflow-x: hidden;
}

/* Background glowing blur circles */
body::before, body::after {
  content: "";
  position: fixed;
  width: 400px;
  height: 400px;
  border-radius: 50%;
  filter: blur(140px);
  z-index: -1;
  opacity: 0.25;
  pointer-events: none;
}

body::before {
  top: -100px;
  right: -50px;
  background: var(--brand);
}

body::after {
  bottom: -100px;
  left: 10%;
  background: var(--accent);
}

/* Scrollbar styles */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: var(--line);
  border-radius: var(--radius-full);
}
::-webkit-scrollbar-thumb:hover {
  background: var(--ink-muted);
}

/* Buttons & Forms */
button, input, textarea, select {
  font: inherit;
  color: inherit;
}

button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 0;
  border-radius: var(--radius-md);
  padding: 10px 18px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}

button.primary {
  background: linear-gradient(135deg, var(--brand), #6366f1);
  color: #fff;
  box-shadow: 0 4px 14px rgba(79, 70, 229, 0.35);
}

button.primary:hover:not(:disabled) {
  transform: translateY(-1.5px);
  box-shadow: 0 6px 20px rgba(79, 70, 229, 0.45);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

input[type="text"], input[type="number"], input[type="password"], textarea, select {
  width: 100%;
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  padding: 11px 16px;
  background: #fff;
  color: var(--ink-primary);
  outline: none;
  transition: var(--transition);
}

input:focus, textarea:focus {
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.15);
}

textarea {
  min-height: 100px;
  resize: vertical;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--ink-secondary);
}

label span {
  margin-left: 2px;
}

/* Brand container */
.brand-container {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
}

.sd-card-icon {
  width: 32px;
  height: 32px;
  color: var(--brand);
  filter: drop-shadow(0 2px 4px rgba(79, 70, 229, 0.15));
}

.brand {
  font-size: 1.45rem;
  font-weight: 800;
  color: var(--ink-primary);
  letter-spacing: -0.5px;
}

/* Shell Layout */
.app-shell {
  display: grid;
  grid-template-columns: 280px 1fr;
  min-height: 100vh;
  gap: 24px;
  padding: 24px;
}

/* Sidebar Navigation */
.nav {
  display: flex;
  flex-direction: column;
  gap: 28px;
  padding: 32px 24px;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-xl);
  background: var(--panel-bg);
  backdrop-filter: blur(16px);
  position: sticky;
  top: 24px;
  height: calc(100vh - 48px);
  box-shadow: var(--shadow-md);
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-grow: 1;
}

.nav-button {
  width: 100%;
  text-align: left;
  background: transparent;
  color: var(--ink-secondary);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: flex-start;
  font-weight: 500;
  transition: var(--transition);
}

.nav-button:hover {
  background: rgba(79, 70, 229, 0.05);
  color: var(--brand);
}

.nav-button.active {
  background: var(--brand);
  color: #fff;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.25);
  font-weight: 600;
}

.nav-button.active .nav-icon {
  color: #fff;
}

.nav-icon {
  width: 20px;
  height: 20px;
  color: var(--ink-secondary);
  transition: var(--transition);
}

.nav-button.logout {
  margin-top: auto;
  border: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.5);
}

.nav-button.logout:hover {
  background: var(--danger-light);
  color: var(--danger);
  border-color: rgba(239, 68, 68, 0.2);
}

.nav-button.logout:hover .nav-icon {
  color: var(--danger);
}

/* Quota widget in sidebar */
.quota-widget {
  background: rgba(241, 245, 249, 0.6);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.quota-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.quota-info span:first-child {
  font-size: 0.75rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--ink-muted);
}

.quota-info span:last-child {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--ink-secondary);
}

.quota-progress {
  width: 100%;
  height: 6px;
  background: var(--line);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.quota-progress span {
  display: block;
  height: 100%;
  width: 0;
  background: linear-gradient(90deg, var(--brand), var(--accent));
  border-radius: var(--radius-full);
  transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Workspace */
.workspace {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.workspace-header {
  border-bottom: 1px solid var(--line);
  padding-bottom: 16px;
}

.workspace-header h1 {
  font-size: 2rem;
  font-weight: 800;
  color: var(--ink-primary);
  letter-spacing: -0.5px;
}

/* Tab Views */
.view {
  display: none;
  animation: fadeIn 0.25s ease-out;
}

.view.active {
  display: block;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* WeTransfer style Send Layout */
.send-layout {
  display: grid;
  place-items: center;
  padding: 24px 0;
  width: 100%;
}

.send-panel {
  width: 100%;
  max-width: 440px;
  background: #fff;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-lg);
  padding: 32px;
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  gap: 22px;
  position: relative;
  overflow: hidden;
}

/* Dropzone style */
.dropzone {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  border: 2px dashed rgba(79, 70, 229, 0.2);
  border-radius: var(--radius-md);
  background: #fafbff;
  cursor: pointer;
  transition: var(--transition);
  padding: 24px;
}

.dropzone:hover, .dropzone.is-over {
  border-color: var(--brand);
  background: rgba(79, 70, 229, 0.02);
}

.dropzone-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.upload-cloud-icon {
  width: 48px;
  height: 48px;
  color: var(--brand);
  transition: var(--transition);
  margin-bottom: 4px;
}

.dropzone:hover .upload-cloud-icon {
  transform: translateY(-4px) scale(1.05);
}

.dropzone strong {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--ink-primary);
}

.dropzone small {
  font-size: 0.85rem;
  color: var(--ink-muted);
}

/* Selected Files Queue style */
.selected-files-container {
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  background: rgba(248, 250, 252, 0.7);
  display: flex;
  flex-direction: column;
  max-height: 200px;
  overflow: hidden;
}

.selected-files-header {
  padding: 10px 14px;
  border-bottom: 1px solid var(--line);
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--ink-secondary);
  background: rgba(241, 245, 249, 0.5);
}

.selected-files-count {
  background: var(--brand-light);
  color: var(--brand);
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
}

.selected-files-list {
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.selected-file-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  font-size: 0.88rem;
}

.selected-file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 80%;
}

.selected-file-name {
  font-weight: 600;
  color: var(--ink-primary);
  overflow: hidden;
  text-overflow: ellipsis;
}

.selected-file-size {
  color: var(--ink-muted);
  font-size: 0.8rem;
  flex-shrink: 0;
}

.btn-remove-file {
  background: transparent;
  color: var(--ink-muted);
  padding: 4px;
  border-radius: var(--radius-sm);
  border: 0;
  cursor: pointer;
  transition: var(--transition);
}

.btn-remove-file:hover {
  background: var(--danger-light);
  color: var(--danger);
}

.btn-remove-file svg {
  width: 16px;
  height: 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.switch {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  cursor: pointer;
}

.switch input {
  width: auto;
  accent-color: var(--brand);
  scale: 1.15;
}

/* WeTransfer style circular progress overlay */
.upload-progress-container {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(12px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  border-radius: var(--radius-lg);
  z-index: 50;
  padding: 32px;
  animation: fadeIn 0.2s ease-out;
}

.gauge-wrapper {
  position: relative;
  width: 120px;
  height: 120px;
}

.progress-ring {
  width: 120px;
  height: 120px;
}

.progress-ring__circle {
  stroke-dasharray: 314;
  stroke-dashoffset: 314;
  transform: rotate(-90deg);
  transform-origin: 60px 60px;
  transition: stroke-dashoffset 0.15s ease-out;
}

.progress-percentage {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  font-size: 1.6rem;
  font-weight: 800;
  color: var(--ink-primary);
}

.upload-speed-info {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--ink-secondary);
  text-align: center;
}

.result {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px;
  border: 1px solid var(--success);
  border-radius: var(--radius-md);
  background: var(--success-light);
  word-break: break-all;
  animation: fadeIn 0.25s ease;
}

.result strong {
  color: #065f46;
  font-size: 1.05rem;
}

.result p {
  background: #fff;
  border: 1px solid rgba(16, 185, 129, 0.25);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.92rem;
  color: var(--ink-primary);
}

.result small {
  color: #065f46;
  font-weight: 600;
}

.actions {
  display: flex;
  flex-direction: column;
}

.actions button {
  width: 100%;
}

.message {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--ink-secondary);
}

.message.error {
  color: var(--danger);
  background: var(--danger-light);
  padding: 10px 14px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(239, 68, 68, 0.15);
}

/* Toolbar & Filters (Google Drive style) */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--line);
  padding-bottom: 16px;
}

.toolbar h2 {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--ink-primary);
  letter-spacing: -0.3px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-bar input {
  max-width: 260px;
}

.btn-icon {
  background: #fff;
  border: 1px solid var(--line);
  padding: 10px;
  aspect-ratio: 1;
  color: var(--ink-secondary);
  border-radius: var(--radius-md);
}

.btn-icon:hover {
  background: rgba(241, 245, 249, 0.8);
  color: var(--brand);
}

.btn-icon svg {
  width: 18px;
  height: 18px;
}

/* Google Drive-inspired grids for files and links */
#filesList, #linksList {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  padding: 4px 0;
}

.list-container {
  overflow-y: auto;
  max-height: 72vh;
  padding-right: 4px;
}

.list {
  display: grid;
  gap: 16px;
}

/* Row-to-Card Grid Representation for Google Drive style */
.row:not(.user-edit):not(.share-row-item) {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 14px;
  padding: 20px;
  background: #fff;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  transition: var(--transition);
  height: 100%;
  position: relative;
}

.row:not(.user-edit):not(.share-row-item):hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: var(--brand);
}

.row-file-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: grid;
  place-items: center;
  flex-shrink: 0;
}

.row-file-icon.image { background: var(--accent-light); color: var(--accent); }
.row-file-icon.archive { background: var(--warn-light); color: var(--warn); }
.row-file-icon.text { background: var(--brand-light); color: var(--brand); }
.row-file-icon.video { background: var(--danger-light); color: var(--danger); }
.row-file-icon.generic { background: #f1f5f9; color: var(--ink-secondary); }

.row-file-icon svg {
  width: 24px;
  height: 24px;
}

.row-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  flex-grow: 1;
}

.row-title {
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--ink-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

.row-meta-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.row-meta {
  font-size: 0.8rem;
  color: var(--ink-muted);
}

.pill {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: var(--radius-full);
  background: #f1f5f9;
  color: var(--ink-secondary);
}

.row-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  margin-top: auto;
  border-top: 1px solid var(--line);
  padding-top: 14px;
  justify-content: flex-end;
}

.row-actions a, .row-actions button {
  background: #fff;
  border: 1px solid var(--line);
  color: var(--ink-secondary);
  font-size: 0.82rem;
  padding: 8px 14px;
  text-decoration: none;
  font-weight: 600;
  border-radius: var(--radius-md);
  transition: var(--transition);
}

.row-actions a:hover, .row-actions button:hover {
  background: #f8fafc;
  color: var(--brand);
  border-color: var(--brand);
}

.row-actions .danger {
  color: var(--danger);
  border-color: var(--line);
}

.row-actions .danger:hover {
  background: var(--danger-light);
  color: var(--danger);
  border-color: var(--danger);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  border: 2px dashed var(--line);
  border-radius: var(--radius-lg);
  color: var(--ink-muted);
  gap: 12px;
  grid-column: 1 / -1;
  text-align: center;
}

.empty-state svg {
  width: 56px;
  height: 56px;
  color: var(--ink-muted);
}

.panel {
  background: #fff;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-lg);
  padding: 28px;
  box-shadow: var(--shadow-sm);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.panel h2 {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--ink-primary);
}

.inline-form {
  display: flex;
  gap: 10px;
}

.inline-form input {
  flex-grow: 1;
}

.grid-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}

/* Modals styles (Glassmorphism backdrop) */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  z-index: 999;
  animation: fadeIn 0.25s ease-out;
}

.modal {
  width: 100%;
  max-width: 500px;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 24px;
  border-bottom: 1px solid var(--line);
  background: rgba(248, 250, 252, 0.85);
}

.modal-icon {
  width: 28px;
  height: 28px;
}

.modal-icon.warning {
  color: var(--warn);
}

.modal-header h2 {
  font-size: 1.3rem;
  font-weight: 800;
  color: var(--ink-primary);
}

.modal-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  font-size: 0.98rem;
  line-height: 1.55;
  color: var(--ink-secondary);
}

.fifo-file-list {
  background: var(--bg-app);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 140px;
  overflow-y: auto;
}

.fifo-file-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  background: #fff;
  font-size: 0.88rem;
  border: 1px solid var(--line);
}

.fifo-file-name {
  font-weight: 600;
  color: var(--ink-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 70%;
}

.fifo-file-size {
  color: var(--danger);
  font-weight: 600;
}

.modal-subtext {
  font-weight: 700;
  color: var(--ink-primary);
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--line);
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: rgba(248, 250, 252, 0.85);
}

.btn-primary {
  background: var(--brand);
  color: #fff;
}

.btn-primary:hover {
  background: var(--brand-hover);
}

.btn-secondary {
  background: #fff;
  border: 1px solid var(--line);
  color: var(--ink-secondary);
}

.btn-secondary:hover {
  background: var(--bg-app);
  color: var(--ink-primary);
}

/* User edit layout custom styles */
.user-edit {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px;
  background: #fff;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  width: 100%;
}

.user-edit .form-grid {
  grid-template-columns: 1fr 1fr;
}

/* Overrides for public Share Card */
.share-body {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background: radial-gradient(circle at top right, #e0e7ff 0%, #f8fafc 40%, #f1f5f9 100%);
}

.share-card {
  width: 100%;
  max-width: 480px;
  background: #fff;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-lg);
  padding: 36px 32px;
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.share-card h1 {
  font-size: 1.7rem;
  font-weight: 800;
  color: var(--ink-primary);
  letter-spacing: -0.5px;
}

.share-card p {
  font-size: 0.98rem;
  color: var(--ink-secondary);
}

.share-card .download {
  width: 100%;
  padding: 14px;
  font-size: 1rem;
}

/* Keep share rows vertical */
.share-row-item {
  display: flex !important;
  flex-direction: row !important;
  align-items: center !important;
  padding: 12px 16px !important;
  border-radius: var(--radius-md) !important;
  background: #f8fafc !important;
  border: 1px solid var(--line) !important;
  box-shadow: none !important;
  transform: none !important;
  width: 100% !important;
  gap: 16px !important;
}

.share-row-item:hover {
  transform: none !important;
  box-shadow: none !important;
  border-color: var(--line) !important;
}

/* Login layout */
.login-body {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background: radial-gradient(circle at top right, #e0e7ff 0%, #f8fafc 40%, #f1f5f9 100%);
}

.login-card {
  width: 100%;
  max-width: 440px;
  background: #fff;
  border: 1px solid var(--panel-border);
  border-radius: var(--radius-lg);
  padding: 36px 32px;
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.login-card h1 {
  font-size: 1.6rem;
  font-weight: 800;
  color: var(--ink-primary);
}

/* Mobile responsive layout */
@media(max-width: 820px) {
  .app-shell {
    grid-template-columns: 1fr;
    padding: 14px;
    gap: 14px;
  }
  
  .nav {
    height: auto;
    position: sticky;
    top: 14px;
    z-index: 100;
    border: 1px solid var(--panel-border);
    border-radius: var(--radius-lg);
    padding: 14px 18px;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    box-shadow: var(--shadow-sm);
  }

  .nav-menu {
    flex-direction: row;
    gap: 4px;
    flex-grow: 0;
  }

  .nav-button {
    width: auto;
    padding: 8px 12px;
  }

  .nav-button span {
    display: none;
  }
  
  .quota-widget {
    display: none;
  }

  .nav-button.logout {
    margin-top: 0;
    width: auto;
    border: 0;
    padding: 8px;
  }

  .workspace {
    padding: 12px 2px;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
  
  #filesList, #linksList {
    grid-template-columns: 1fr;
    gap: 14px;
  }
}

.selected-file-icon {
  width: 16px;
  height: 16px;
  color: var(--brand);
  flex-shrink: 0;
}
.user-row-header {
  display: flex;
  justify-content: space-between;
  width: 100%;
  align-items: center;
}
.user-row-meta {
  margin-bottom: 8px;
}
.user-form-grid {
  width: 100%;
  margin-bottom: 8px;
}
.user-switch {
  margin-bottom: 12px;
}
.user-actions {
  justify-content: flex-end;
  width: 100%;
}
.shared-files-preview {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin: 6px 0;
  width: 100%;
}
.shared-file-tag {
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--ink-secondary);
  background: #f1f5f9;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border: 1px solid var(--line);
  max-width: 100%;
}
`
}

// AppJS devuelve la logica de la app autenticada.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - JavaScript de la interfaz.
func AppJS() string {
	return appJS
}

// ShareJS devuelve la logica de la pagina publica.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - JavaScript de share publico.
func ShareJS() string {
	return shareJS
}
