# Build stage
FROM node:22.17.1-alpine AS builder

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
FROM nginx:alpine

# Install envsubst for runtime variable injection
RUN apk add --no-cache gettext

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy entrypoint script
COPY ./apps/management-interface/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Expose port 80
EXPOSE 80

# Run entrypoint that injects runtime env and starts nginx
ENTRYPOINT ["/docker-entrypoint.sh"]
