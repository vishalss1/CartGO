# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files from frontend directory
COPY src/frontend/package.json src/frontend/package-lock.json* ./

# Install dependencies
RUN npm install

# Copy frontend source
COPY src/frontend/ .

# Build the app
RUN npm run build


# Stage 2: Serve with Nginx
FROM nginx:alpine

# Clean default nginx html
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*

# Copy built files from builder (Vite builds to dist/)
COPY --from=builder /app/dist .

# Copy nginx config
COPY src/frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
