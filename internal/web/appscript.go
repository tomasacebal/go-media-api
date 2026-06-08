package web

const appJS = `
const state={me:null,files:[],shares:[],keys:[],users:[],selectedFiles:[]};
const $=(q)=>document.querySelector(q);
const $$=(q)=>Array.from(document.querySelectorAll(q));
let title,quotaText,quotaMeter,sendMessage,resultBox,progress,progressBar,progressText,expiresDays,neverExpires;

async function init(){
  title=$("#pageTitle");quotaText=$("#quotaText");quotaMeter=$("#quotaMeter");
  sendMessage=$("#sendMessage");resultBox=$("#resultBox");
  progress=$("#uploadProgress");progressBar=$("#uploadBar");progressText=$("#uploadText");
  expiresDays=$("#expiresDays");neverExpires=$("#neverExpires");

  bindTabs();
  bindSend();
  bindLists();
  bindAccount();
  bindModal();
  bindSearch();
  await loadMe();
  await Promise.all([loadFiles(),loadShares(),loadKeys()]);
}

document.addEventListener("DOMContentLoaded", init);

function bindTabs(){
  $$(".nav-button[data-tab]").forEach((btn)=>btn.addEventListener("click",()=>showTab(btn.dataset.tab)));
  $("#logoutButton").addEventListener("click",async()=>{
    await fetch("/logout",{method:"POST"});
    location.href="/login";
  });
}

function showTab(tab){
  $$(".nav-button[data-tab]").forEach((b)=>b.classList.toggle("active",b.dataset.tab===tab));
  $$(".view").forEach((v)=>v.classList.remove("active"));
  $("#"+tab+"View").classList.add("active");
  title.textContent={send:"Enviar archivos",links:"Enlaces compartidos",files:"Todos los archivos",account:"Mi Cuenta",admin:"Administración"}[tab]||"AceTransfer";
  if(tab==="admin")loadUsers();
}

function bindSend(){
  const form=$("#transferForm"),input=$("#fileInput"),drop=$("#dropzone");
  
  input.addEventListener("change",()=>{
    const files=Array.from(input.files);
    files.forEach(f=>{
      if(!state.selectedFiles.some(x=>x.name===f.name&&x.size===f.size)){
        state.selectedFiles.push(f);
      }
    });
    input.value="";
    renderSelectedFiles();
  });
  
  neverExpires.addEventListener("change",()=>{
    expiresDays.disabled=neverExpires.checked;
    if(neverExpires.checked)expiresDays.value="";
  });
  
  drop.addEventListener("dragover",(e)=>{
    e.preventDefault();
    drop.classList.add("is-over");
  });
  drop.addEventListener("dragleave",()=>drop.classList.remove("is-over"));
  drop.addEventListener("drop",(e)=>{
    e.preventDefault();
    drop.classList.remove("is-over");
    const files=Array.from(e.dataTransfer.files);
    if(files.length){
      files.forEach(f=>{
        if(!state.selectedFiles.some(x=>x.name===f.name&&x.size===f.size)){
          state.selectedFiles.push(f);
        }
      });
      renderSelectedFiles();
    }
  });
  
  form.addEventListener("submit",async(e)=>{
    e.preventDefault();
    await submitTransfer(form);
  });
}

function renderSelectedFiles() {
  const box=$("#selectedFilesBox");
  const list=$("#selectedFilesList");
  const count=$("#selectedFilesCount");
  const label=$("#fileLabel");
  
  if(state.selectedFiles.length===0){
    box.hidden=true;
    list.innerHTML="";
    count.textContent="0 archivos";
    label.textContent="Añade tus archivos";
  }else{
    box.hidden=false;
    count.textContent=state.selectedFiles.length + (state.selectedFiles.length===1 ? " archivo" : " archivos");
    label.textContent=state.selectedFiles.length + (state.selectedFiles.length===1 ? " archivo seleccionado" : " archivos seleccionados");
    list.innerHTML=state.selectedFiles.map((f,idx)=>
      '<div class="selected-file-item">' +
        '<div class="selected-file-info">' +
          '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="selected-file-icon"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>' +
          '<span class="selected-file-name" title="' + escapeHtml(f.name) + '">' + escapeHtml(f.name) + '</span>' +
          '<span class="selected-file-size">' + bytes(f.size) + '</span>' +
        '</div>' +
        '<button type="button" class="btn-remove-file" data-index="' + idx + '" title="Quitar archivo">' +
          '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
        '</button>' +
      '</div>'
    ).join("");
    
    list.querySelectorAll(".btn-remove-file").forEach(btn=>{
      btn.addEventListener("click",()=>{
        console.log("Removing file index:", btn.dataset.index);
        const idx=parseInt(btn.dataset.index,10);
        state.selectedFiles.splice(idx,1);
        renderSelectedFiles();
      });
    });
  }
}

// ── Chunked Upload ─────────────────────────────────────────────────────────
const CHUNK_THRESHOLD = 25 * 1024 * 1024; // usar chunked si algun archivo > 25 MB
const MAX_CHUNK_RETRIES = 3;
const SHA256_LIMIT = 500 * 1024 * 1024;   // calcular hash solo si < 500 MB

// Calcula SHA-256 de un archivo usando Web Crypto API.
// Para archivos > SHA256_LIMIT devuelve cadena vacia (sin verificacion).
async function computeFileSHA256(file){
  if(file.size > SHA256_LIMIT) return "";
  const buf = await file.arrayBuffer();
  const hash = await crypto.subtle.digest("SHA-256", buf);
  return Array.from(new Uint8Array(hash)).map(b=>b.toString(16).padStart(2,"0")).join("");
}

// Sube un unico fragmento con reintentos automaticos ante fallos de red.
async function uploadChunkWithRetry(uploadId, chunkIndex, chunkBlob, retriesLeft){
  try{
    const resp = await fetch("/api/v1/upload/chunk",{
      method:"POST",
      headers:{
        "Content-Type":"application/octet-stream",
        "X-Upload-ID":uploadId,
        "X-Chunk-Index":String(chunkIndex),
      },
      body: chunkBlob,
    });
    if(!resp.ok){
      const e = await resp.json().catch(()=>({}));
      throw new Error(e.error?.message || "HTTP "+resp.status);
    }
  }catch(err){
    if(retriesLeft > 0){
      const delay = 1200 * (MAX_CHUNK_RETRIES - retriesLeft + 1);
      await new Promise(r=>setTimeout(r, delay));
      return uploadChunkWithRetry(uploadId, chunkIndex, chunkBlob, retriesLeft - 1);
    }
    throw {message: "Fragmento "+(chunkIndex+1)+" fallo tras "+MAX_CHUNK_RETRIES+" intentos: "+err.message};
  }
}

// Flujo de subida fragmentada para archivos grandes.
async function submitTransferChunked(form, confirmed){
  const button = $("#sendButton");
  button.disabled = true;
  sendMessage.textContent = "";
  sendMessage.classList.remove("error");
  resultBox.hidden = true;
  showProgress(0, false);

  try{
    const titleVal   = form.querySelector("[name='title']").value;
    const messageVal = form.querySelector("[name='message']").value;
    const expVal     = form.querySelector("[name='expires_days']").value;
    const expiresD   = expVal ? parseInt(expVal, 10) : 0;
    const neverExp   = form.querySelector("[name='never_expires']").checked || expiresD===0;

    const totalBytes = state.selectedFiles.reduce((s,f)=>s+f.size, 0);
    let uploadedBytes = 0;
    const peerFileIDs = [];

    for(let fi=0; fi<state.selectedFiles.length; fi++){
      const file = state.selectedFiles[fi];
      const isLast = fi === state.selectedFiles.length - 1;

      // 1. Calcular hash de integridad (solo archivos manejables)
      $("#uploadSpeedInfo").textContent = "Calculando integridad de "+escapeHtml(file.name)+"...";
      const sha256 = await computeFileSHA256(file);

      // 2. Iniciar sesion de carga
      $("#uploadSpeedInfo").textContent = "Iniciando sesion para "+escapeHtml(file.name)+"...";
      const initResp = await fetch("/api/v1/upload/init",{
        method:"POST",
        headers:{"Content-Type":"application/json"},
        body: JSON.stringify({filename:file.name, total_size:file.size}),
      });
      if(!initResp.ok){
        const e = await initResp.json().catch(()=>({}));
        throw {message: e.error?.message || "Error iniciando sesion de carga"};
      }
      const {data:initData} = await initResp.json();
      const uploadId   = initData.upload_id;
      const chunkSize  = initData.chunk_size_bytes;
      const totalChunks = initData.total_chunks;

      // 3. Subir fragmentos secuencialmente con reintento por fragmento
      let speedSample = {t: Date.now(), b: 0}; // para calcular velocidad
      for(let ci=0; ci<totalChunks; ci++){
        const start = ci * chunkSize;
        const end   = Math.min(start + chunkSize, file.size);
        const blob  = file.slice(start, end);
        const chunkBytes = end - start;

        const t0 = Date.now();
        await uploadChunkWithRetry(uploadId, ci, blob, MAX_CHUNK_RETRIES);
        const elapsed = (Date.now() - t0) / 1000 || 0.001;

        uploadedBytes += chunkBytes;
        speedSample.b += chunkBytes;

        // Velocidad instantanea del ultimo fragmento (MB/s)
        const mbps = (chunkBytes / elapsed / (1024 * 1024)).toFixed(1);

        const pct = Math.min(98, Math.round((uploadedBytes / totalBytes) * 98));
        showProgress(pct, false);

        const fileTag = state.selectedFiles.length > 1
          ? "Archivo "+(fi+1)+"/"+state.selectedFiles.length+" — "
          : "";
        $("#uploadSpeedInfo").textContent =
          fileTag + pct + "% — " + mbps + " MB/s";
      }

      // 4. Finalizar sesion: ensamblar + verificar SHA-256 + crear transferencia en el ultimo archivo
      showProgress(99, true);
      const finishBody = {upload_id:uploadId, expected_sha256:sha256};
      if(isLast){
        finishBody.title         = titleVal;
        finishBody.message       = messageVal;
        finishBody.expires_days  = expiresD;
        finishBody.never_expires = neverExp;
        finishBody.confirm_fifo  = confirmed;
        if(peerFileIDs.length > 0) finishBody.peer_file_ids = peerFileIDs;
      }

      const finishResp = await fetch("/api/v1/upload/finish",{
        method:"POST",
        headers:{"Content-Type":"application/json"},
        body: JSON.stringify(finishBody),
      });

      if(!finishResp.ok){
        const errData = await finishResp.json().catch(()=>({}));
        if(finishResp.status===409 && errData.error?.code==="quota_cleanup_required"){
          progress.hidden = true;
          showFifoModal(errData.data, form);
          return;
        }
        throw {message: errData.error?.message || "Error finalizando la carga"};
      }

      const {data:finishData} = await finishResp.json();

      if(isLast){
        showProgress(100, true);
        progress.hidden = true;
        state.selectedFiles = [];
        renderSelectedFiles();
        form.reset();
        expiresDays.disabled = false;
        showResult(finishData);
        await Promise.all([loadMe(), loadFiles(), loadShares()]);
      }else{
        // Guardar file_id para incluirlo en la transferencia del ultimo archivo
        peerFileIDs.push(finishData.file_id);
      }
    }
  }catch(err){
    sendMessage.textContent = err.message || "Error durante la carga fragmentada";
    sendMessage.classList.add("error");
    progress.hidden = true;
  }finally{
    button.disabled = false;
  }
}
// ────────────────────────────────────────────────────────────────────────────

async function submitTransfer(form, confirmed=false){
  sendMessage.textContent="";
  sendMessage.classList.remove("error");
  resultBox.hidden=true;

  if(state.selectedFiles.length===0){
    sendMessage.textContent="Por favor selecciona al menos un archivo para enviar";
    sendMessage.classList.add("error");
    return;
  }

  // Delegar a subida fragmentada si algun archivo supera el umbral de 25 MB
  if(state.selectedFiles.some(f=>f.size>CHUNK_THRESHOLD)){
    return submitTransferChunked(form, confirmed);
  }

  const button=$("#sendButton");
  button.disabled=true;
  showProgress(0, false);

  try{
    const body=new FormData();
    state.selectedFiles.forEach(f=>body.append("files", f));
    body.append("title", form.querySelector("[name='title']").value);

    const expVal = form.querySelector("[name='expires_days']").value;
    if(expVal) body.append("expires_days", expVal);

    if(form.querySelector("[name='never_expires']").checked){
      body.append("never_expires", "true");
    }
    body.append("message", form.querySelector("[name='message']").value);

    if(confirmed) body.set("confirm_fifo","true");

    const payload=await uploadWithProgress("/api/v1/transfers/",body);

    progress.hidden=true;
    state.selectedFiles=[];
    renderSelectedFiles();
    form.reset();
    expiresDays.disabled=false;
    showResult(payload.data);
    await Promise.all([loadMe(),loadFiles(),loadShares()]);
  }catch(err){
    if(err.status===409&&err.payload?.error?.code==="quota_cleanup_required"){
      progress.hidden=true;
      showFifoModal(err.payload.data, form);
      return;
    }
    sendMessage.textContent=err.message;
    sendMessage.classList.add("error");
    progress.hidden=true;
  }finally{
    button.disabled=false;
  }
}

function uploadWithProgress(url,body){
  return new Promise((resolve,reject)=>{
    const xhr=new XMLHttpRequest();
    xhr.open("POST",url);
    // 10-minute timeout for large files
    xhr.timeout = 10 * 60 * 1000;
    
    let lastTime = Date.now();
    let lastLoaded = 0;
    
    xhr.upload.addEventListener("progress",(event)=>{
      if(event.lengthComputable) {
        const percent = Math.round((event.loaded/event.total)*100);
        showProgress(percent, false);
        
        // speed estimation
        const now = Date.now();
        const elapsedSec = (now - lastTime) / 1000;
        if(elapsedSec > 0.5) {
          const loadedBytes = event.loaded - lastLoaded;
          const bps = loadedBytes / elapsedSec;
          $("#uploadSpeedInfo").textContent = "Subiendo: " + bytes(bps) + "/s — " + bytes(event.loaded) + " de " + bytes(event.total);
          lastTime = now;
          lastLoaded = event.loaded;
        }
      }
    });
    
    // Fires when the browser finishes sending all bytes (before server responds)
    xhr.upload.addEventListener("load",()=>{
      showProgress(99, true);
    });
    
    xhr.addEventListener("load",()=>{
      const payload=parseJSON(xhr.responseText);
      if(xhr.status>=200&&xhr.status<300){
        showProgress(100, true);
        resolve(payload);
        return;
      }
      reject({status:xhr.status,payload,message:payload.error?.message||"No se pudo crear el link"});
    });
    xhr.addEventListener("error",()=>reject({status:0,payload:null,message:"No se pudo conectar al servidor"}));
    xhr.addEventListener("timeout",()=>reject({status:0,payload:null,message:"La subida excedió el tiempo de espera. Intenta con un archivo más pequeño o revisa tu conexión."}));
    xhr.send(body);
  });
}

function showProgress(value, processing){
  progress.hidden=false;
  // Circumference of SVG circle is 2*pi*50 = 314.16
  const circumference = 314;
  const offset = circumference - (value / 100) * circumference;
  progressBar.style.strokeDashoffset = offset;
  progressText.textContent = value+"%";
  if (processing) {
    $("#uploadSpeedInfo").textContent = "Procesando en servidor...";
  }
}

function bindModal() {
  $("#cancelFifoBtn").addEventListener("click", () => {
    $("#fifoModal").hidden = true;
  });
}

function showFifoModal(preview, form) {
  const modal = $("#fifoModal");
  const list = $("#fifoFileList");
  const confirmBtn = $("#confirmFifoBtn");
  
  list.innerHTML = preview.files.map(f => 
    '<div class="fifo-file-item">' +
      '<span class="fifo-file-name" title="' + escapeHtml(f.title || f.original_name) + '">' + escapeHtml(f.title || f.original_name) + '</span>' +
      '<span class="fifo-file-size">-' + bytes(f.size_bytes) + '</span>' +
    '</div>'
  ).join("");
  
  // Clone button to strip previous retry event listener
  const newConfirmBtn = confirmBtn.cloneNode(true);
  confirmBtn.parentNode.replaceChild(newConfirmBtn, confirmBtn);
  
  newConfirmBtn.addEventListener("click", async () => {
    modal.hidden = true;
    await submitTransfer(form, true);
  });
  
  modal.hidden = false;
}

function showResult(data){
  const url=data.share.url;
  const expiry=data.share.never_expires?"Sin vencimiento":"Vence "+date(data.share.expires_at);
  resultBox.hidden=false;
  resultBox.innerHTML=
    '<strong>¡Enlace creado con éxito!</strong>' +
    '<p>' + escapeHtml(url) + '</p>' +
    '<small>' + expiry + '</small>' +
    '<div class="row-actions">' +
      '<a href="' + url + '" target="_blank" rel="noreferrer">Abrir</a>' +
      '<button type="button" id="copyLink">Copiar enlace</button>' +
    '</div>';
    
  $("#copyLink").addEventListener("click",async()=>{
    await navigator.clipboard.writeText(url);
    $("#copyLink").textContent="Copiado";
    setTimeout(() => { $("#copyLink").textContent="Copiar enlace"; }, 2000);
  });
}

function bindLists(){
  $("#refreshFiles").addEventListener("click",loadFiles);
  $("#refreshLinks").addEventListener("click",loadShares);
  $("#filesList").addEventListener("click", async (e) => {
    const copyBtn = e.target.closest("[data-copy-file-link]");
    if (copyBtn) {
      await handleCopyFileLink(copyBtn.dataset.copyFileLink, copyBtn);
      return;
    }
    const deleteBtn = e.target.closest("[data-delete-file]");
    if (deleteBtn) {
      await handleDeleteFile(deleteBtn.dataset.deleteFile);
    }
  });
  $("#linksList").addEventListener("click",revokeShare);
}

function bindSearch() {
  $("#searchFiles").addEventListener("input", () => {
    renderFilesList(state.files);
  });
  $("#searchLinks").addEventListener("input", () => {
    renderSharesList(state.shares);
  });
}

async function loadMe(){
  const payload=await api("/api/v1/me");
  state.me=payload.data;
  $("#adminTab").hidden=state.me.role!=="admin";
  const percent=Math.min(100,Math.round((state.me.storage_used_bytes/state.me.quota_bytes)*100));
  quotaText.textContent=bytes(state.me.storage_used_bytes)+" de "+bytes(state.me.quota_bytes)+" ("+percent+"%)";
  quotaMeter.style.width=percent+"%";
  renderAccount();
}

async function loadFiles(){
  const payload=await api("/api/v1/files/?limit=100");
  state.files=payload.data||[];
  renderFilesList(state.files);
}

async function loadShares(){
  const payload=await api("/api/v1/shares/?limit=100");
  state.shares=payload.data||[];
  renderSharesList(state.shares);
}

function getFileIconClassAndSVG(mime) {
  const m = String(mime || "").toLowerCase();
  if (m.startsWith("image/")) {
    return { class: "image", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>' };
  }
  if (m.includes("zip") || m.includes("tar") || m.includes("rar") || m.includes("gzip") || m.includes("7z")) {
    return { class: "archive", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"/><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/></svg>' };
  }
  if (m.startsWith("text/") || m.includes("pdf") || m.includes("document") || m.includes("msword") || m.includes("sheet") || m.includes("xml")) {
    return { class: "text", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>' };
  }
  if (m.startsWith("video/")) {
    return { class: "video", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>' };
  }
  return { class: "generic", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>' };
}

function renderFilesList(list) {
  const term = ($("#searchFiles")?.value || "").toLowerCase();
  const filtered = list.filter(f => (f.title || f.original_name || "").toLowerCase().includes(term) || (f.mime_type || "").toLowerCase().includes(term));
  
  if(filtered.length === 0) {
    $("#filesList").innerHTML = emptyStateHTML("No se encontraron archivos");
  } else {
    $("#filesList").innerHTML = filtered.map(fileRow).join("");
  }
}

function fileRow(file){
  const icon = getFileIconClassAndSVG(file.mime_type);
  return '<article class="row">' +
    '<div class="row-file-icon ' + icon.class + '">' + icon.svg + '</div>' +
    '<div class="row-content">' +
      '<span class="row-title">' + escapeHtml(file.title || file.original_name) + '</span>' +
      '<div class="row-meta-group">' +
        '<span class="pill">' + bytes(file.size_bytes) + '</span>' +
        '<span class="row-meta">' + escapeHtml(file.mime_type) + '</span>' +
      '</div>' +
    '</div>' +
    '<div class="row-actions">' +
      '<a href="/api/v1/files/' + encodeURIComponent(file.id) + '/download" title="Descargar">Descargar</a>' +
      '<button data-copy-file-link="' + escapeHtml(file.id) + '">Copiar link</button>' +
      '<button data-delete-file="' + escapeHtml(file.id) + '" class="danger" title="Borrar">Borrar</button>' +
    '</div>' +
  '</article>';
}

function renderSharesList(list) {
  const term = ($("#searchLinks")?.value || "").toLowerCase();
  const filtered = list.filter(s => (s.code || "").toLowerCase().includes(term) || (s.target_type || "").toLowerCase().includes(term));
  
  if(filtered.length === 0) {
    $("#linksList").innerHTML = emptyStateHTML("No se encontraron enlaces");
  } else {
    $("#linksList").innerHTML = filtered.map(shareRow).join("");
  }
}

function shareRow(share){
  const expiry=share.never_expires?"Sin vencimiento":"Vence "+date(share.expires_at);
  const targetIcon = share.target_type === "file" ? 
    '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>' :
    '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>';
  
  let filesListHtml = "";
  if (share.shared_files && share.shared_files.length > 0) {
    filesListHtml = '<div class="shared-files-preview">' +
      share.shared_files.map(f => '<span class="shared-file-tag" title="' + escapeHtml(f) + '">' + escapeHtml(f) + '</span>').join("") +
    '</div>';
  }

  return '<article class="row">' +
    '<div class="row-file-icon generic">' + targetIcon + '</div>' +
    '<div class="row-content">' +
      '<span class="row-title">/' + escapeHtml(share.code) + '</span>' +
      filesListHtml +
      '<div class="row-meta-group">' +
        '<span class="pill">' + escapeHtml(share.target_type) + '</span>' +
        '<span class="row-meta">' + expiry + '</span>' +
      '</div>' +
    '</div>' +
    '<div class="row-actions">' +
      '<a href="' + escapeHtml(share.url) + '" target="_blank">Abrir</a>' +
      '<button data-copy="' + escapeHtml(share.url) + '">Copiar</button>' +
      '<button data-revoke-share="' + escapeHtml(share.id) + '" class="danger">Revocar</button>' +
    '</div>' +
  '</article>';
}

async function handleDeleteFile(fileId){
  await fetch("/api/v1/files/"+encodeURIComponent(fileId),{method:"DELETE"});
  await Promise.all([loadMe(),loadFiles(),loadShares()]);
}

async function handleCopyFileLink(fileId, button) {
  try {
    const existing = state.shares.find(s => s.target_type === "file" && s.target_id === fileId);
    let url = "";
    if (existing) {
      url = existing.url;
    } else {
      const payload = await api("/api/v1/shares/", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ target_type: "file", target_id: fileId })
      });
      url = payload.data.url;
      await loadShares();
    }
    await navigator.clipboard.writeText(url);
    const originalText = button.textContent;
    button.textContent = "Copiado";
    setTimeout(() => { button.textContent = originalText; }, 2000);
  } catch (err) {
    console.error(err);
    alert("No se pudo copiar el enlace: " + err.message);
  }
}

async function revokeShare(e){
  const copy=e.target.closest("[data-copy]");
  if(copy){
    await navigator.clipboard.writeText(copy.dataset.copy);
    copy.textContent="Copiado";
    setTimeout(() => { copy.textContent="Copiar"; }, 2000);
    return;
  }
  const btn=e.target.closest("[data-revoke-share]");
  if(!btn)return;
  if(!confirm("¿Estás seguro de que deseas revocar este enlace?"))return;
  await fetch("/api/v1/shares/"+encodeURIComponent(btn.dataset.revokeShare),{method:"DELETE"});
  await loadShares();
}

function bindAccount(){
  $("#keyForm").addEventListener("submit",async(e)=>{
    e.preventDefault();
    const name=new FormData(e.target).get("name");
    const payload=await api("/api/v1/api-keys/",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({name,scopes:["read","write","delete"]})});
    $("#keyMessage").textContent="Key creada: "+payload.data.secret;
    e.target.reset();
    await loadKeys();
  });
  $("#keysList").addEventListener("click",async(e)=>{
    const btn=e.target.closest("[data-revoke-key]");
    if(!btn)return;
    if(!confirm("¿Estás seguro de que deseas revocar esta API Key?"))return;
    await fetch("/api/v1/api-keys/"+encodeURIComponent(btn.dataset.revokeKey),{method:"DELETE"});
    await loadKeys();
  });
  $("#userForm").addEventListener("submit",createUser);
  $("#usersList").addEventListener("submit",updateUser);
}

async function loadKeys(){
  const payload=await api("/api/v1/api-keys/");
  state.keys=payload.data||[];
  $("#keysList").innerHTML=state.keys.length?state.keys.map((k)=>
    '<article class="row">' +
      '<div class="row-file-icon generic">' +
        '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
      '</div>' +
      '<div class="row-content">' +
        '<span class="row-title">' + escapeHtml(k.name) + '</span>' +
        '<p class="row-meta">Prefijo: ' + escapeHtml(k.key_prefix) + '... | Permisos: ' + escapeHtml(k.scopes.join(", ")) + '</p>' +
      '</div>' +
      '<div class="row-actions">' +
        '<button data-revoke-key="' + escapeHtml(k.id) + '" class="danger">Revocar</button>' +
      '</div>' +
    '</article>'
  ).join(""):emptyStateHTML("No tienes API keys creadas");
}

function renderAccount(){
  if(!state.me)return;
  const ttl=state.me.share_ttl_days===0?"Sin vencimiento":state.me.share_ttl_days+" días";
  $("#accountBox").innerHTML=
    '<section class="panel">' +
      '<h2>' + escapeHtml(state.me.name) + '</h2>' +
      '<p class="row-meta">' + escapeHtml(state.me.email) + ' — Rol: <strong>' + escapeHtml(state.me.role) + '</strong></p>' +
      '<p class="row-meta">Uso de espacio: <strong>' + bytes(state.me.storage_used_bytes) + '</strong> de <strong>' + bytes(state.me.quota_bytes) + '</strong></p>' +
      '<p class="row-meta">Vencimiento de enlaces por defecto: <strong>' + ttl + '</strong></p>' +
    '</section>';
}

async function loadUsers(){
  if(!state.me||state.me.role!=="admin")return;
  const payload=await api("/api/v1/users/");
  state.users=payload.data||[];
  $("#usersList").innerHTML=state.users.map(userRow).join("");
}

function userRow(u){
  const quotaGB=Math.round((u.quota_bytes/1024/1024/1024)*100)/100;
  return '<form class="row user-edit" data-user-form="' + escapeHtml(u.id) + '">' +
    '<div class="user-row-header">' +
      '<span class="row-title">' + escapeHtml(u.name) + '</span>' +
      '<span class="pill">' + escapeHtml(u.role) + '</span>' +
    '</div>' +
    '<p class="row-meta user-row-meta">' + escapeHtml(u.email) + ' — Uso: ' + bytes(u.storage_used_bytes) + ' / ' + bytes(u.quota_bytes) + '</p>' +
    '<div class="form-grid user-form-grid">' +
      '<label><span>Límite (GB)</span><input name="quota_gb" type="number" min="1" step="0.5" value="' + quotaGB + '" /></label>' +
      '<label><span>Vence en (días)</span><input name="share_ttl_days" type="number" min="0" value="' + u.share_ttl_days + '" /></label>' +
    '</div>' +
    '<label class="switch user-switch"><input name="disabled" type="checkbox" ' + (u.disabled_at?"checked":"") + ' /><span>Deshabilitar usuario</span></label>' +
    '<div class="row-actions user-actions">' +
      '<button class="primary">Guardar cambios</button>' +
    '</div>' +
  '</form>';
}

async function createUser(e){
  e.preventDefault();
  const data=new FormData(e.target);
  const quotaGB=Number(data.get("quota_gb")||10);
  const ttl=Number(data.get("share_ttl_days")||30);
  await api("/api/v1/users/",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({email:data.get("email"),name:data.get("name"),password:data.get("password"),role:"user",quota_bytes:quotaGB*1024*1024*1024,share_ttl_days:ttl})});
  e.target.reset();
  await loadUsers();
}

async function updateUser(e){
  const form=e.target.closest("[data-user-form]");
  if(!form)return;
  e.preventDefault();
  const data=new FormData(form);
  const quotaGB=Number(data.get("quota_gb")||1);
  const ttl=Number(data.get("share_ttl_days")||0);
  await api("/api/v1/users/"+encodeURIComponent(form.dataset.userForm),{method:"PATCH",headers:{"Content-Type":"application/json"},body:JSON.stringify({quota_bytes:quotaGB*1024*1024*1024,share_ttl_days:ttl,disabled:form.querySelector("input[name='disabled']").checked})});
  await Promise.all([loadUsers(),loadMe()]);
}

async function api(url,options){
  const res=await fetch(url,options);
  const payload=await res.json().catch(()=>({}));
  if(!res.ok)throw new Error(payload.error?.message||"Error de API");
  return payload;
}

function parseJSON(text){try{return JSON.parse(text||"{}");}catch{return{};}}

function bytes(value){
  if(!value)return"0 B";
  const units=["B","KB","MB","GB","TB"];
  let size=value,unit=0;
  while(size>=1024&&unit<units.length-1){size/=1024;unit++;}
  return size.toFixed(size>=10||unit===0?0:1)+" "+units[unit];
}

function date(value){
  if(!value)return"Sin vencimiento";
  return new Date(value).toLocaleDateString("es-AR");
}

function escapeHtml(value){
  return String(value??"").replace(/[&<>"']/g,(c)=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#039;"}[c]));
}

function emptyStateHTML(text){
  return '<div class="empty-state">' +
    '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">' +
      '<circle cx="12" cy="12" r="10"/><path d="M8 12h8"/>' +
    '</svg>' +
    '<strong>' + escapeHtml(text) + '</strong>' +
  '</div>';
}
`
