#!/bin/bash

# CartGO Kubernetes Deployment Script with Nginx Ingress
# This script deploys all CartGO services to a Minikube cluster with Nginx Ingress

set -e

echo "🚀 CartGO Kubernetes Deployment with Nginx Ingress"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo -e "\n${YELLOW}[1/5]${NC} Checking prerequisites..."

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}✗ kubectl not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ kubectl${NC}"

if ! command -v minikube &> /dev/null; then
    echo -e "${RED}✗ minikube not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ minikube${NC}"

# Verify Minikube is running
if ! minikube status | grep -q "host: Running"; then
    echo -e "${RED}✗ Minikube is not running${NC}"
    echo "  Start Minikube with: minikube start"
    exit 1
fi
echo -e "${GREEN}✓ Minikube is running${NC}"

# Check if ingress addon is enabled
if ! minikube addons list | grep -q "ingress.*enabled"; then
    echo -e "${YELLOW}! Enabling Ingress addon...${NC}"
    minikube addons enable ingress
fi
echo -e "${GREEN}✓ Ingress addon is enabled${NC}"

# Create namespace
echo -e "\n${YELLOW}[2/5]${NC} Creating namespace..."
kubectl apply -f k8s/namespace.yml
echo -e "${GREEN}✓ Namespace 'cartgo' created${NC}"

# Build and load Docker images
echo -e "\n${YELLOW}[3/5]${NC} Building Docker images..."
eval $(minikube docker-env)
docker-compose build
echo -e "${GREEN}✓ Docker images built${NC}"

# Deploy all services (except ingress first)
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

# Deploy Ingress (last)
echo -e "\n${YELLOW}[5/5]${NC} Setting up Nginx Ingress..."
kubectl apply -f k8s/ingress.yml
echo -e "${GREEN}✓ Ingress configured${NC}"

# Get Minikube IP
MINIKUBE_IP=$(minikube ip)

echo -e "\n${GREEN}✅ Deployment complete!${NC}"
echo ""
echo "📋 Access Information:"
echo "   Minikube IP: ${MINIKUBE_IP}"
echo ""
echo "🌐 Service URLs:"
echo "   API Gateway (via Ingress): http://${MINIKUBE_IP}/api/v1"
echo "   API Gateway (via NodePort): http://${MINIKUBE_IP}:30080/api/v1"
echo ""
echo "🏥 Health Checks:"
echo "   curl http://${MINIKUBE_IP}/health"
echo "   curl http://${MINIKUBE_IP}/api/v1/health"
echo ""
echo "📝 Useful Commands:"
echo "   # Monitor ingress"
echo "   kubectl get ingress -n cartgo -w"
echo ""
echo "   # View ingress logs"
echo "   kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -f"
echo ""
echo "   # Check service status"
echo "   kubectl get pods -n cartgo"
echo ""
echo "   # Describe ingress"
echo "   kubectl describe ingress cartgo-ingress -n cartgo"
echo ""
echo "💡 Tips:"
echo "   1. Add to /etc/hosts (Linux/Mac) or C:\\Windows\\System32\\drivers\\etc\\hosts (Windows):"
echo "      ${MINIKUBE_IP} cartgo.local"
echo ""
echo "   2. Then access via: http://cartgo.local"
echo ""
echo "📖 For more information, see k8s/QUICK_REFERENCE.md"
echo ""
