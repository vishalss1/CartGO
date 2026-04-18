#!/bin/bash

set -e

echo "=========================================="
echo "CartGO Kubernetes Deployment"
echo "=========================================="

# Ensure namespace exists
echo "Creating namespace..."
kubectl create namespace cartgo --dry-run=client -o yaml | kubectl apply -f -

# Generate ConfigMap from SQL file at deploy time
echo "Creating PostgreSQL init ConfigMap from scripts/postgresql/init-db.sql..."
kubectl create configmap postgres-init-scripts \
  --from-file=init-db.sql=scripts/postgresql/init-db.sql \
  -n cartgo \
  --save-config \
  --dry-run=client \
  -o yaml | kubectl apply -f -

# Apply Kubernetes manifests in order
echo "Applying Kubernetes manifests..."

# Database
echo "Applying PostgreSQL..."
kubectl apply -f k8s/postgres-db.yml

# Services (in dependency order)
echo "Applying User Service..."
kubectl apply -f k8s/user-deployment.yml
kubectl apply -f k8s/user-service.yml

echo "Applying Product Service..."
kubectl apply -f k8s/product-deployment.yml
kubectl apply -f k8s/product-service.yml

echo "Applying Inventory Service..."
kubectl apply -f k8s/inventory-deployment.yml
kubectl apply -f k8s/inventory-service.yml

echo "Applying Order Service..."
kubectl apply -f k8s/order-deployment.yml
kubectl apply -f k8s/order-service.yml

echo "Applying Payment Service..."
kubectl apply -f k8s/payment-deployment.yml
kubectl apply -f k8s/payment-service.yml

echo "Applying Delivery Service..."
kubectl apply -f k8s/delivery-deployment.yml
kubectl apply -f k8s/delivery-service.yml

echo "Applying Support Service..."
kubectl apply -f k8s/support-deployment.yml
kubectl apply -f k8s/support-service.yml

# API Gateway (last, depends on all services)
echo "Applying API Gateway..."
kubectl apply -f k8s/api-gateway-deployment.yml
kubectl apply -f k8s/api-gateway-service.yml

# Frontend
echo "Applying Frontend..."
kubectl apply -f k8s/frontend-deployment.yml
kubectl apply -f k8s/frontend-service.yml

echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
echo ""
echo "Checking pod status..."
kubectl get pods -n cartgo
echo ""
echo "To port-forward API Gateway (8080):"
echo "  kubectl port-forward -n cartgo svc/api-gateway 8080:8080"
echo ""
echo "To port-forward Frontend (3000):"
echo "  kubectl port-forward -n cartgo svc/frontend 3000:80"
echo ""
echo "To view logs:"
echo "  kubectl logs -n cartgo -f deployment/<service-name>"
