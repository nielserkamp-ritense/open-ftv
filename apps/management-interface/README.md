# OpenFTV Management Interface

This application serves as the frontend for the OpenFTV project. It provides a user-friendly interface for managing and interacting with the various components of the OpenFTV (Federatieve Toegangsverlening) reference implementation.

## Build

Install dependencies

```bash
npm install
```

Run development server

```bash
npm run dev
```

## Docker images

CI publishes multi-arch images (linux/amd64 and linux/arm64) built with ko. The Dockerfile builds the same web server for Docker Compose and local use; its build context is the repository root.

Build the image

```bash
# From the repository root
docker build -f apps/management-interface/Dockerfile -t management-interface-app:latest-dev .
```

Run the container, serving the app on port 8080 and the health checks on port 8081 (see [Web server and runtime configuration](#web-server-and-runtime-configuration) for the environment variables)

```bash
docker run -p 8080:8080 -p 8081:8081 management-interface-app:latest-dev
```

## Technical Architecture

The OpenFTV Management Interface is built using modern web technologies and follows a component-based architecture. Here's an overview of the technical stack and structure:

### Technology Stack

- **Framework**: React 19 with TypeScript for type safety
- **Build Tool**: Vite for fast development and optimized production builds
- **Routing**: TanStack Router (formerly React Router) for declarative routing
- **Styling**: Tailwind CSS for utility-first styling

### Project Structure

The application follows a modular structure organized by feature and responsibility:

- **/src/components/**: Reusable UI components (buttons, inputs, tables, etc.)
- **/src/routes/**: Page components and route definitions
- **/src/oas/**: OpenAPI specification types and API clients
- **/src/models/**: TypeScript interfaces and type definitions
- **/src/assets/**: Static assets like images and icons

### Component Architecture

The UI is built using a component-based approach with:

1. **Base Components**: Fundamental UI elements like buttons, inputs, and form controls
2. **Layout Components**: Structural components like sidebar layouts and stacked layouts
3. **Page Components**: Full page implementations that compose multiple components

### Routing System

The application uses TanStack Router with a file-based routing system:

- **/__root.tsx**: The root layout with navigation sidebar and header
- **/index.tsx**: The home page
- **/policies/index.tsx**: The policies management page
- **/attributes.tsx**: The attributes management page

### API Integration

The application integrates with backend services using:

1. **OpenAPI Generated Types**: Automatically generated TypeScript types from OpenAPI specifications
2. **Service Functions**: Dedicated functions for data fetching and API interactions
3. **Type-Safe API Clients**: Ensuring type safety between frontend and backend

## Web server and runtime configuration

The production image is a small Go web server (`cmd`, `config`, `server`) with the built app embedded from `web/dist`. It serves the app on port 8080, falls back to `index.html` for client-side routes, and serves liveness and readiness checks on port 8081 (`/livez`, `/readyz`).

Build the assets before compiling the server; the embedded `web/dist` only holds a `.gitkeep` on a clean checkout, and the server refuses to start without `index.html`:
```bash
# From apps/management-interface
npm ci && npm run build
```

Vite normally inlines `import.meta.env.*` at build time. To keep the image the same for every environment, the app loads `/runtime-env.js` before the main bundle and reads `window.__ENV__`, falling back to `import.meta.env` during `npm run dev`. The web server renders `/runtime-env.js` at startup from these environment variables:
- `VITE_PAP_BASE_URL` and `VITE_PIP_BASE_URL` (default `http://localhost:8080`)
- `VITE_OIDC_AUTHORITY` and `VITE_OIDC_CLIENT_ID`
- `MANAGER_MAX_LOGO_SIZE`
- `VITE_APP_VERSION`

The server's own settings use the `MANAGEMENT_INTERFACE_` prefix, e.g. `MANAGEMENT_INTERFACE_PORT` and `MANAGEMENT_INTERFACE_HEALTH_PORT`.

Build and run the image (the build context is the repository root):
```bash
# From the repository root
docker build -f apps/management-interface/Dockerfile -t management-interface-app:latest .

docker run -p 8080:8080 \
  -e VITE_PAP_BASE_URL="https://pap.example.com" \
  -e VITE_PIP_BASE_URL="https://pip.example.com" \
  management-interface-app:latest
```
