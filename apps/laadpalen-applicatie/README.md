# Laadpalen applicatie

Demo application of Gemeente Vlierdam for requesting a laadpaal (charging point) permit. The behandelaar screens call
the Vlierdam dataspace through OpenFTV and check each step with the PDP.

## Development

```sh
npm install
npm run dev
```

The dev server listens on port 8080 and reads its settings from `.env`. `modd` starts it together with the rest of the
Vlierdam setup (see `modd.conf` in the repository root).

Built with Vite, TypeScript, React, shadcn/ui and Tailwind CSS.

## Web server and runtime configuration

The production image is a small Go web server (`cmd`, `config`, `server`) with the built app embedded from `web/dist`. It serves the app on port 8080, falls back to `index.html` for client-side routes, and serves liveness and readiness checks on port 8081 (`/livez`, `/readyz`). CI publishes it as a multi-arch image (linux/amd64 and linux/arm64) built with ko.

Build the assets before compiling the server; the embedded `web/dist` only holds a `.gitkeep` on a clean checkout, and the server refuses to start without `index.html`:
```bash
# From apps/laadpalen-applicatie
npm ci && npm run build

# From the repository root
go run ./apps/laadpalen-applicatie/cmd
```

The app loads `/runtime-env.js` before the main bundle and reads `window.__ENV__`, falling back to the build-time `import.meta.env` (see `src/lib/env.ts`). The web server renders `/runtime-env.js` at startup from these environment variables; an unset variable keeps the build-time value:
- `VITE_API_BASE_URL` and `VITE_VLIERDAM_OUTWAY_URL`
- `VITE_BRP_API_KEY`
- `VITE_DOELBINDING`, `VITE_DOELBINDING_BRP` and `VITE_DOELBINDING_BRV`
- `VITE_BEHEER_URL`, `VITE_RVIG_BEHEER_URL` and `VITE_RDW_BEHEER_URL`
- `VITE_VLIERDAM_PDP1_URL` and `VITE_VLIERDAM_PDP2_URL`

The server's own settings use the `LAADPALEN_APPLICATIE_` prefix, e.g. `LAADPALEN_APPLICATIE_PORT` and `LAADPALEN_APPLICATIE_HEALTH_PORT`.
