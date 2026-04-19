#!/bin/bash

# CartGO Kubernetes Deployment Script with Nginx Ingress
# Fixed: Proper ordering, waits for postgres, no premature pod deletion

set -e

# Remove any docker alias or function that might interfere
unset -f docker 2>/dev/null || true
unalias docker 2>/dev/null || true

echo "=================================================="
echo " CartGO Kubernetes Deployment"
echo "=================================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Track failed builds
FAILED_BUILDS=()

# -------------------------------
# Helper Functions
# -------------------------------

check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED} $1 not found${NC}"
        return 1
    fi
    return 0
}

wait_for_pods() {
    local namespace=$1
    local label=$2
    local expected_ready=$3
    local timeout=$4
    local start_time=$(date +%s)

    echo -e "${YELLOW} Waiting for $label pods to be ready...${NC}"

    while true; do
        local ready_count
        ready_count=$(kubectl get pods -n "$namespace" -l "$label" \
            -o jsonpath='{range .items[*]}{.status.phase}{"\n"}{end}' 2>/dev/null | grep -c "Running" || echo "0")

        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ "$ready_count" -ge "$expected_ready" ]; then
            echo -e "${GREEN} $label pods are ready ($ready_count/$expected_ready)${NC}"
            return 0
        fi

        if [ $elapsed -gt $timeout ]; then
            echo -e "${RED} Timeout waiting for $label pods after ${timeout}s${NC}"
            return 1
        fi

        echo "  Waiting... ($ready_count/$expected_ready ready, ${elapsed}s/${timeout}s)"
        kubectl get pods -n "$namespace" -l "$label" --no-headers 2>/dev/null | while read line; do
            echo "    $line"
        done
        sleep 5
    done
}

wait_for_postgres_ready() {
    local namespace=$1
    local timeout=${2:-120}
    local start_time=$(date +%s)

    echo -e "${YELLOW} Waiting for PostgreSQL to accept connections...${NC}"

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ $elapsed -gt $timeout ]; then
            echo -e "${RED} Timeout waiting for PostgreSQL after ${timeout}s${NC}"
            return 1
        fi

        # Check if postgres pod is running and ready
        local pod_status
        pod_status=$(kubectl get pods -n "$namespace" -l "app=postgres" \
            -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NotFound")

        if [ "$pod_status" = "Running" ]; then
            # Try to connect
            if kubectl exec -n "$namespace" deploy/postgres -- pg_isready -U cartgo_user &>/dev/null; then
                echo -e "${GREEN} PostgreSQL is ready and accepting connections${NC}"
                return 0
            fi
        fi

        echo "  PostgreSQL status: $pod_status (${elapsed}s/${timeout}s)"
        sleep 5
    done
}

build_image() {
    local image_name=$1
    local dockerfile=$2
    local tag="${image_name}:1.0"

    echo -e "${BLUE} Building ${image_name}...${NC}"

    # Retry logic for Docker builds
    local attempts=0
    local max_attempts=3

    while [ $attempts -lt $max_attempts ]; do
        attempts=$((attempts + 1))
        echo "  Build attempt $attempts/$max_attempts..."

        if docker build -t "$tag" -f "$dockerfile" . 2>&1; then
            if docker images "$tag" --format "{{.Repository}}" | grep -q "$image_name"; then
                echo -e "${GREEN}  $image_name built successfully${NC}"
                return 0
            fi
        fi

        echo -e "${YELLOW}  Build failed, retrying in 5s...${NC}"
        sleep 5
    done

    echo -e "${RED}  Failed to build $image_name${NC}"
    FAILED_BUILDS+=("$image_name")
    return 1
}

# -------------------------------
# [1/7] Prerequisites
# -------------------------------

echo -e "\n${YELLOW}[1/7]${NC} Checking prerequisites..."

check_command kubectl || exit 1
echo -e "${GREEN} kubectl${NC}"

check_command minikube || exit 1
echo -e "${GREEN} minikube${NC}"

# Check if minikube is running
if ! minikube status &>/dev/null; then
    echo -e "${RED} Minikube is not running${NC}"
    echo "Run: minikube start --driver=docker"
    exit 1
fi
echo -e "${GREEN} Minikube is running${NC}"

# Enable ingress if needed
if ! minikube addons list 2>/dev/null | grep -E "ingress.*enabled" &>/dev/null; then
    echo -e "${YELLOW}! Enabling Ingress addon...${NC}"
    minikube addons enable ingress
fi
echo -e "${GREEN} Ingress addon enabled${NC}"

# -------------------------------
# [2/7] Configure Docker
# -------------------------------

echo -e "\n${YELLOW}[2/7]${NC} Configuring Docker environment..."

if ! docker info 2>/dev/null | grep -q "minikube"; then
    eval $(minikube -p minikube docker-env --shell bash) || {
        echo -e "${RED} Failed to configure Docker environment${NC}"
        exit 1
    }
    echo -e "${GREEN} Docker configured for Minikube${NC}"
else
    echo -e "${GREEN} Docker already configured${NC}"
fi

# -------------------------------
# [3/7] Build Images
# -------------------------------

echo -e "\n${YELLOW}[3/7]${NC} Building Docker images..."

build_image "api-gateway" "src/api-gateway/Dockerfile"
build_image "user-service" "src/services/user-service/Dockerfile"
build_image "product-service" "src/services/product-service/Dockerfile"
build_image "inventory-service" "src/services/inventory-service/Dockerfile"
build_image "order-service" "src/services/order-service/Dockerfile"
build_image "payment-service" "src/services/payment-service/Dockerfile"
build_image "delivery-service" "src/services/delivery-service/Dockerfile"
build_image "support-service" "src/services/support-service/Dockerfile"
build_image "frontend" "src/frontend/Dockerfile"

if [ ${#FAILED_BUILDS[@]} -gt 0 ]; then
    echo -e "\n${RED} Failed builds: ${FAILED_BUILDS[*]}${NC}"
    exit 1
fi

echo -e "${GREEN} All images built successfully${NC}"

# -------------------------------
# [4/7] Create Namespace & Config
# -------------------------------

echo -e "\n${YELLOW}[4/7]${NC} Creating namespace and config..."
kubectl apply -f src/k8s/namespace.yml
echo -e "${GREEN} Namespace ready${NC}"

# Create postgres init configmap
kubectl apply -f src/k8s/postgres-init-configmap.yml
echo -e "${GREEN} Postgres init scripts created${NC}"

# -------------------------------
# [5/7] Deploy PostgreSQL First
# -------------------------------

echo -e "\n${YELLOW}[5/7]${NC} Deploying PostgreSQL..."

# Delete any old postgres resources to ensure clean start
kubectl delete deployment postgres -n cartgo --ignore-not-found=true
kubectl delete pvc postgres-pvc -n cartgo --ignore-not-found=true
sleep 2

# Apply postgres with PVC
kubectl apply -f src/k8s/postgres-db.yml

# Wait for postgres to be ready
echo -e "${YELLOW} Waiting for PostgreSQL (this may take 1-2 minutes)...${NC}"
wait_for_pods "cartgo" "app=postgres" 1 180 || {
    echo -e "${RED} PostgreSQL failed to start${NC}"
    echo "Check logs: kubectl logs -n cartgo deploy/postgres"
    exit 1
}

# Wait for postgres to accept connections
wait_for_postgres_ready "cartgo" 120 || {
    echo -e "${RED} PostgreSQL is not accepting connections${NC}"
    echo "Check logs: kubectl logs -n cartgo deploy/postgres"
    exit 1
}

# Give extra time for init scripts to run
sleep 5

echo -e "${GREEN} PostgreSQL is fully ready${NC}"

# -------------------------------
# [6/7] Deploy Microservices
# -------------------------------

echo -e "\n${YELLOW}[6/7]${NC} Deploying microservices..."

# Apply all deployments and services
kubectl apply -f src/k8s/user-deployment.yml -f src/k8s/user-service.yml
kubectl apply -f src/k8s/product-deployment.yml -f src/k8s/product-service.yml
kubectl apply -f src/k8s/inventory-deployment.yml -f src/k8s/inventory-service.yml
kubectl apply -f src/k8s/order-deployment.yml -f src/k8s/order-service.yml
kubectl apply -f src/k8s/payment-deployment.yml -f src/k8s/payment-service.yml
kubectl apply -f src/k8s/delivery-deployment.yml -f src/k8s/delivery-service.yml
kubectl apply -f src/k8s/support-deployment.yml -f src/k8s/support-service.yml
kubectl apply -f src/k8s/api-gateway-deployment.yml -f src/k8s/api-gateway-service.yml
kubectl apply -f src/k8s/frontend-deployment.yml -f src/k8s/frontend-service.yml

echo -e "${GREEN} Microservices deployed${NC}"

# Wait for all pods to be ready
echo -e "\n${YELLOW} Waiting for all services to be ready (this may take 2-3 minutes)...${NC}"

SERVICES=("api-gateway" "user-service" "product-service" "inventory-service" \
          "order-service" "payment-service" "delivery-service" "support-service" "frontend")

for svc in "${SERVICES[@]}"; do
    wait_for_pods "cartgo" "app=$svc" 1 180 || {
        echo -e "${YELLOW} $svc may need more time${NC}"
    }
done

# -------------------------------
# [7/7] Configure Ingress
# -------------------------------

echo -e "\n${YELLOW}[7/7]${NC} Configuring Ingress..."
kubectl apply -f src/k8s/ingress.yml
echo -e "${GREEN} Ingress configured${NC}"

# -------------------------------
# Final Status
# -------------------------------

echo ""
echo "=========================================="
echo " Deployment Status"
echo "=========================================="
sleep 2
kubectl get pods -n cartgo

echo ""
MINIKUBE_IP=$(minikube ip 2>/dev/null || echo "minikube-ip")

echo -e "${GREEN} Deployment complete!${NC}"
echo ""
echo " API Gateway: http://${MINIKUBE_IP}/api/v1"
echo " Health:      http://${MINIKUBE_IP}/health"
echo " Frontend:    http://${MINIKUBE_IP}/"
echo ""
echo " Test health: ./test_health.sh"
echo ""

# Check final status
NOT_RUNNING=$(kubectl get pods -n cartgo --field-selector=status.phase!=Running,status.phase!=Succeeded 2>/dev/null | grep -v "NAME" | wc -l || echo "0")
if [ "$NOT_RUNNING" -gt 0 ]; then
    echo -e "${YELLOW} $NOT_RUNNING pod(s) not yet running.${NC}"
    echo "Run: kubectl get pods -n cartgo -w"
fi
