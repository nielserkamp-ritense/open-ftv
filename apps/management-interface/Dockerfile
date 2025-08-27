# Build stage
FROM node:22.17.1-alpine AS builder

ARG VITE_PAP_BASE_URL
ARG VITE_PIP_BASE_URL

WORKDIR /app

# Copy package files
COPY ./apps/management-interface/package.json /app
COPY ./apps/management-interface/package-lock.json /app

# Install dependencies
RUN npm ci

# Copy source code
COPY ./apps/management-interface/ .

# Fix .env content
RUN sed -i -r "s|%VITE_PAP_BASE_URL%|${VITE_PAP_BASE_URL}|g" .env
RUN sed -i -r "s|%VITE_PIP_BASE_URL%|${VITE_PIP_BASE_URL}|g" .env

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration if you have custom config
# COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
