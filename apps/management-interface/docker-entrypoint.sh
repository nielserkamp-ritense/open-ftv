#!/usr/bin/env sh
set -eu

HTML_DIR=/usr/share/nginx/html
TEMPLATE="$HTML_DIR/runtime-env.template.js"
TARGET="$HTML_DIR/runtime-env.js"

# Generate runtime-env.js from template using environment variables
if [ -f "$TEMPLATE" ]; then
  echo "[entrypoint] Generating runtime-env.js from template"
  # Substitute only the variables we expect
  envsubst '\$VITE_PAP_BASE_URL \$VITE_PIP_BASE_URL' < "$TEMPLATE" > "$TARGET"
else
  echo "[entrypoint] Template not found at $TEMPLATE; ensuring $TARGET exists"
  if [ ! -f "$TARGET" ]; then
    echo "window.__ENV__ = window.__ENV__ || {};" > "$TARGET"
  fi
fi

# Log what got injected (for debugging)
if [ -f "$TARGET" ]; then
  echo "[entrypoint] Effective runtime env values:"
  grep -E 'VITE_(PAP|PIP)_BASE_URL' "$TARGET" || true
fi

# Start nginx in foreground
exec nginx -g 'daemon off;'
