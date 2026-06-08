package web

const shareJS = `
const code=location.pathname.split("/").filter(Boolean).pop();
const title=document.querySelector("#shareTitle"),meta=document.querySelector("#shareMeta"),files=document.querySelector("#shareFiles"),download=document.querySelector("#downloadButton");

loadShare();

async function loadShare(){
  try{
    const res=await fetch("/api/v1/public/shares/"+encodeURIComponent(code));
    const payload=await res.json();
    if(!res.ok)throw new Error(payload.error?.message||"Link no disponible");
    render(payload.data);
  }catch(err){
    title.textContent="Link no disponible";
    meta.textContent=err.message;
    files.innerHTML="";
    download.hidden=true;
  }
}

function render(data){
  const names=data.files.map((file)=>file.original_name);
  const expiry=data.share.never_expires?"sin vencimiento":"vence "+new Date(data.share.expires_at).toLocaleDateString("es-AR");
  
  title.textContent=data.transfer?.title||names[0]||"Descarga";
  meta.textContent=data.files.length+" archivo(s) — "+expiry;
  
  files.innerHTML=data.files.map((file)=>{
    const icon = getFileIconClassAndSVG(file.original_name);
    return '<article class="row share-row-item">' +
        '<div class="row-file-icon ' + icon.class + '">' + icon.svg + '</div>' +
        '<div class="row-content">' +
          '<span class="row-title">' + escapeHtml(file.original_name) + '</span>' +
          '<span class="row-meta">Tamaño: ' + bytes(file.size_bytes) + '</span>' +
        '</div>' +
      '</article>';
  }).join("");
  
  download.href="/s/"+encodeURIComponent(data.share.code)+"/download";
}

function getFileIconClassAndSVG(filename) {
  const ext = String(filename || "").split('.').pop().toLowerCase();
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp'];
  const archiveExts = ['zip', 'rar', 'tar', 'gz', '7z', 'bz2'];
  const textExts = ['txt', 'pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'csv', 'md', 'xml', 'html', 'json', 'js', 'go'];
  const videoExts = ['mp4', 'mkv', 'avi', 'mov', 'wmv', 'flv', 'webm'];
  
  if (imageExts.includes(ext)) {
    return { class: "image", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>' };
  }
  if (archiveExts.includes(ext)) {
    return { class: "archive", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"/><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/></svg>' };
  }
  if (textExts.includes(ext)) {
    return { class: "text", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>' };
  }
  if (videoExts.includes(ext)) {
    return { class: "video", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>' };
  }
  return { class: "generic", svg: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>' };
}

function bytes(value){
  if(!value)return"0 B";
  const units=["B","KB","MB","GB","TB"];
  let size=value,unit=0;
  while(size>=1024&&unit<units.length-1){size/=1024;unit++;}
  return size.toFixed(size>=10||unit===0?0:1)+" "+units[unit];
}

function escapeHtml(value){
  return String(value??"").replace(/[&<>"']/g,(c)=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#039;"}[c]));
}
`
