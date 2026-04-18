# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files from frontend folder
COPY frontend/package.json frontend/package-lock.json* frontend/yarn.lock* ./

# Install dependencies
RUN npm install

# Copy full frontend source
COPY frontend/ .

# Build the app
RUN npm run build


# Stage 2: Serve with Nginx
FROM nginx:alpine

# Clean default nginx html
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*

# Copy built files from builder
COPY --from=builder /app/dist .

# Copy nginx config from frontend folder
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]