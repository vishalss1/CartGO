import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Get the Minikube IP from environment or use localhost as fallback
const MINIKUBE_IP = process.env.MINIKUBE_IP || "localhost";
const API_TARGET = process.env.API_TARGET || `http://${MINIKUBE_IP}`;

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    hmr: {
      protocol: "ws",
      host: "localhost",
      port: 5173,
    },
    proxy: {
      "/api/v1": {
        target: API_TARGET,
        changeOrigin: true,
        rewrite: (path) => path,
      },
      "/health": {
        target: API_TARGET,
        changeOrigin: true,
        rewrite: (path) => path,
      },
    },
  },
});

/**
 * Usage:
 * 
 * For local development (default):
 *   npm run dev
 *   - Uses localhost:8080 as API target
 * 
 * For Minikube development:
 *   MINIKUBE_IP=192.168.49.2 npm run dev
 *   - Uses http://192.168.49.2 (Ingress) as API target
 * 
 * For custom API target:
 *   API_TARGET=http://api.cartgo.local npm run dev
 *   - Uses http://api.cartgo.local as API target
 * 
 * Environment Setup:
 *   1. Get Minikube IP: minikube ip
 *   2. Set environment variable: export MINIKUBE_IP=$(minikube ip)
 *   3. Start dev server: npm run dev
 * 
 * Windows PowerShell:
 *   $env:MINIKUBE_IP=$(minikube ip)
 *   npm run dev
 * 
 * Windows CMD:
 *   for /f %i in ('minikube ip') do set MINIKUBE_IP=%i
 *   npm run dev
 */
