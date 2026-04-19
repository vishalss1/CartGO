#!/bin/bash

# CartGO Kubernetes Deployment Script with Nginx Ingress

set -e

# Remove any docker alias or function that might interfere
unset -f docker 2>/dev/null || true
unalias docker 2>/dev/null || true

echo "=================================================="
echo "🚀 CartGO Kubernetes Deployment with Nginx Ingress"
echo "=================================================="

# Colors

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# -------------------------------

# [1/5] Prerequisites

# -------------------------------

echo -e "\n${YELLOW}[1/5]${NC} Checking prerequisites..."

command -v kubectl >/dev/null || { echo -e "${RED}✗ kubectl not found${NC}"; exit 1; }
echo -e "${GREEN}✓ kubectl${NC}"

command -v minikube >/dev/null || { echo -e "${RED}✗ minikube not found${NC}"; exit 1; }
echo -e "${GREEN}✓ minikube${NC}"

minikube status | grep -q "host: Running" || {
echo -e "${RED}✗ Minikube is not running${NC}"
echo "Run: minikube start"
exit 1
}
echo -e "${GREEN}✓ Minikube is running${NC}"

# Enable ingress if needed

if ! minikube addons list | grep -q "ingress.*enabled"; then
echo -e "${YELLOW}! Enabling Ingress addon...${NC}"
minikube addons enable ingress
fi
echo -e "${GREEN}✓ Ingress addon is enabled${NC}"

# -------------------------------

# [2/5] Namespace

# -------------------------------

echo -e "\n${YELLOW}[2/5]${NC} Creating namespace..."
kubectl apply -f k8s/namespace.yml
echo -e "${GREEN}✓ Namespace created${NC}"

# -------------------------------

# [3/5] Build Images (FIXED)

# -------------------------------

echo -e "\n${YELLOW}[3/5]${NC} Building Docker images..."

# Point Docker to Minikube

eval $(minikube -p minikube docker-env --shell bash)

# Build with EXACT names used in Kubernetes YAML

docker build -t api-gateway:1.0 -f api-gateway/Dockerfile .

docker build -t user-service:1.0 -f services/user-service/Dockerfile .
docker build -t product-service:1.0 -f services/product-service/Dockerfile .
docker build -t inventory-service:1.0 -f services/inventory-service/Dockerfile .
docker build -t order-service:1.0 -f services/order-service/Dockerfile .
docker build -t payment-service:1.0 -f services/payment-service/Dockerfile .
docker build -t delivery-service:1.0 -f services/delivery-service/Dockerfile .
docker build -t support-service:1.0 -f services/support-service/Dockerfile .

docker build -t frontend:1.0 -f frontend/Dockerfile .

echo -e "${GREEN}✓ Docker images built inside Minikube${NC}"

# -------------------------------

# [4/5] Deploy Services

# -------------------------------

echo -e "\n${YELLOW}[4/5]${NC} Deploying services..."

kubectl apply -f k8s/postgres-db.yml

kubectl apply -f k8s/user-deployment.yml -f k8s/user-service.yml
kubectl apply -f k8s/product-deployment.yml -f k8s/product-service.yml
kubectl apply -f k8s/inventory-deployment.yml -f k8s/inventory-service.yml
kubectl apply -f k8s/order-deployment.yml -f k8s/order-service.yml
kubectl apply -f k8s/payment-deployment.yml -f k8s/payment-service.yml
kubectl apply -f k8s/delivery-deployment.yml -f k8s/delivery-service.yml
kubectl apply -f k8s/support-deployment.yml -f k8s/support-service.yml

kubectl apply -f k8s/api-gateway-deployment.yml -f k8s/api-gateway-service.yml
kubectl apply -f k8s/frontend-deployment.yml -f k8s/frontend-service.yml

echo -e "${GREEN}✓ Services deployed${NC}"

# -------------------------------

# [5/5] Ingress

# -------------------------------

echo -e "\n${YELLOW}[5/5]${NC} Setting up Ingress..."
kubectl apply -f k8s/ingress.yml
echo -e "${GREEN}✓ Ingress configured${NC}"

# -------------------------------

# Restart Pods (IMPORTANT)

# -------------------------------

echo -e "\n${YELLOW}Restarting pods to use new images...${NC}"
kubectl delete pods -n cartgo --all

# -------------------------------

# Info

# -------------------------------

MINIKUBE_IP=$(minikube ip)

echo -e "\n${GREEN}✅ Deployment complete!${NC}"
echo ""
echo "🌐 API Gateway:"
echo "   http://${MINIKUBE_IP}/api/v1"
echo ""
echo "🏥 Health:"
echo "   http://${MINIKUBE_IP}/health"
echo ""
echo "📊 Monitor:"
echo "   kubectl get pods -n cartgo -w"
echo ""
