#!/bin/bash

# CartGO Health Check Script
# Tests each microservice by port-forwarding and checking health endpoint

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Services to test: (service_name, kubernetes_service, port)
declare -a SERVICES=(
    "api-gateway:api-gateway:8080"
    "user-service:user-service:8081"
    "product-service:product-service:8082"
    "inventory-service:inventory-service:8083"
    "order-service:order-service:8084"
    "payment-service:payment-service:8085"
    "delivery-service:delivery-service:8086"
    "support-service:support-service:8087"
    "frontend:frontend:3000"
)

echo "=========================================="
echo "CartGO Health Check"
echo "=========================================="
echo ""

test_service() {
    local service_name=$1
    local k8s_service=$2
    local port=$3
    
    echo -n "Testing $service_name ($port)... "
    
    # Start port-forward in background
    kubectl port-forward -n cartgo svc/$k8s_service $port:$port > /dev/null 2>&1 &
    PF_PID=$!
    
    # Wait for port-forward to establish
    sleep 2
    
    # Test health endpoint (or root for frontend)
    if [ "$service_name" = "frontend" ]; then
        HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port/" 2>/dev/null || echo "000")
    else
        HEALTH_RESPONSE=$(curl -s "http://localhost:$port/health" 2>/dev/null || echo "")
    fi
    
    # Kill port-forward
    kill $PF_PID 2>/dev/null || true
    wait $PF_PID 2>/dev/null || true
    
    # Check if response is healthy
    if [ "$service_name" = "frontend" ]; then
        # For frontend, check HTTP 200 status
        if [ "$HEALTH_RESPONSE" = "200" ]; then
            echo -e "${GREEN}✓ Healthy${NC}"
            return 0
        else
            echo -e "${RED}✗ HTTP $HEALTH_RESPONSE${NC}"
            return 1
        fi
    else
        # For services, check JSON response
        if echo "$HEALTH_RESPONSE" | grep -q '"status":"OK"'; then
            echo -e "${GREEN}✓ Healthy${NC}"
            return 0
        elif [ -n "$HEALTH_RESPONSE" ]; then
            echo -e "${YELLOW}⚠ Response received but unexpected format${NC}"
            echo "    Response: $HEALTH_RESPONSE"
            return 0
        else
            echo -e "${RED}✗ No response${NC}"
            return 1
        fi
    fi
}

# Test each service
PASSED=0
FAILED=0

for service in "${SERVICES[@]}"; do
    IFS=':' read -r service_name k8s_service port <<< "$service"
    
    if test_service "$service_name" "$k8s_service" "$port"; then
        ((PASSED++))
    else
        ((FAILED++))
    fi
done

echo ""
echo "=========================================="
echo "Results: $PASSED passed, $FAILED failed"
echo "=========================================="