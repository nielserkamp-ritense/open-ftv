// This file is processed at container startup by envsubst to produce runtime-env.js
// Pass values via `docker run -e VITE_PAP_BASE_URL=... -e VITE_PIP_BASE_URL=...`
window.__ENV__ = {
  VITE_PAP_BASE_URL: "${VITE_PAP_BASE_URL}",
  VITE_PIP_BASE_URL: "${VITE_PIP_BASE_URL}",
};
