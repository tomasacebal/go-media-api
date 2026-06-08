package web

import _ "embed"

// SDCardIcon contiene los bytes binarios del archivo SD-Card.ico.
//go:embed SD-Card.ico
var SDCardIcon []byte

// SDCardSVG es el icono SD-Card en formato SVG para servir como favicon.svg.
// Los navegadores modernos soportan SVG como favicon de forma nativa.
const SDCardSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#6366f1" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <rect x="4" y="6" width="16" height="14" rx="2" ry="2"/>
  <path d="M4 6V18A2 2 0 0 0 6 20H18A2 2 0 0 0 20 18V9L15 4H6A2 2 0 0 0 4 6Z"/>
  <path d="M9 10V6"/>
  <path d="M12 10V6"/>
  <path d="M15 10V6"/>
</svg>`
