# Build stage
FROM docker.io/node:22.17.1-alpine AS builder

WORKDIR /app

# Copy package files
COPY ./apps/management-interface/package.json /app
COPY ./apps/management-interface/package-lock.json /app

# Install dependencies
RUN npm ci

# Copy source code
COPY ./apps/management-interface/ .

# Build the application
RUN npm run build

# Production stage
FROM docker.io/nginxinc/nginx-unprivileged:1.29-alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html
COPY --from=builder /app/util/default.conf.template /etc/nginx/templates/default.conf.template
