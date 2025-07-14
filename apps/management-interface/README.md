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
Build the image
```bash
docker build -t management-interface-app:latest-dev .
```

Run the container on port 8080
```bash
docker run -p 8080:80 management-interface-app:latest-dev
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