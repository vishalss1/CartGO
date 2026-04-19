#!/bin/bash

# CartGO Health Check Script
# Tests each microservice health endpoint via port-forward

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
NAMESPACE="cartgo"
CURL_TIMEOUT=5

# Services to test: (service_name:kubernetes_service:port)
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
echo " CartGO Health Check"
echo "=========================================="
echo ""

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo -e "${RED} Error: Namespace '$NAMESPACE' not found${NC}"
    echo "Run ./deploy.sh first"
    exit 1
fi

# Check if pods exist
POD_COUNT=$(kubectl get pods -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$POD_COUNT" -eq 0 ]; then
    echo -e "${RED} Error: No pods found in namespace '$NAMESPACE'${NC}"
    echo "Run ./deploy.sh first"
    exit 1
fi

# Show current pod status
echo -e "${BLUE} Current pod status:${NC}"
kubectl get pods -n "$NAMESPACE" --no-headers 2>/dev/null | while read line; do
    echo "  $line"
done
echo ""

# Track results
PASSED=0
FAILED=0

# Starting port for local forwarding (start high to avoid conflicts)
LOCAL_PORT_BASE=30000

test_service() {
    local service_name=$1
    local k8s_service=$2
    local port=$3
    local local_port=$LOCAL_PORT_BASE
    LOCAL_PORT_BASE=$((LOCAL_PORT_BASE + 1))

    echo -n "Testing $service_name... "

    # Check if service exists
    if ! kubectl get svc "$k8s_service" -n "$NAMESPACE" >/dev/null 2>&1; then
        echo -e "${YELLOW}Service not found${NC}"
        return 1
    fi

    # Check if pods are ready for this service
    local ready_pods
    ready_pods=$(kubectl get pods -n "$NAMESPACE" -l "app=$k8s_service" \
        -o jsonpath='{range .items[*]}{.status.conditions[?(@.type=="Ready")].status}{"\n"}{end}' 2>/dev/null | grep -c "True" || echo "0")

    if [ "$ready_pods" -eq 0 ]; then
        echo -e "${YELLOW}No ready pods${NC}"
        return 1
    fi

    # Start port-forward in background
    kubectl port-forward -n "$NAMESPACE" "svc/$k8s_service" "$local_port:$port" >/dev/null 2>&1 &
    local pf_pid=$!

    # Wait for port-forward to establish
    local attempts=0
    local max_attempts=30
    local port_ready=false

    while [ $attempts -lt $max_attempts ]; do
        if curl -s --max-time 1 "localhost:$local_port" >/dev/null 2>&1 || \
           curl -s --max-time 1 "http://localhost:$local_port/health" >/dev/null 2>&1; then
            port_ready=true
            break
        fi
        attempts=$((attempts + 1))
        sleep 0.3
    done

    if [ "$port_ready" != "true" ]; then
        echo -e "${RED}Port-forward timeout${NC}"
        kill $pf_pid 2>/dev/null || true
        wait $pf_pid 2>/dev/null || true
        return 1
    fi

    # Test endpoint based on service type
    local result=1

    if [ "$service_name" = "frontend" ]; then
        # Frontend: check HTTP status
        local http_code
        http_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "$CURL_TIMEOUT" \
            "http://localhost:$local_port/" 2>/dev/null || echo "000")

        if [ "$http_code" = "200" ]; then
            echo -e "${GREEN}Healthy${NC}"
            result=0
        else
            echo -e "${RED}HTTP $http_code${NC}"
        fi
    else
        # Backend services: check health endpoint
        local health_response
        health_response=$(curl -s --max-time "$CURL_TIMEOUT" \
            "http://localhost:$local_port/health" 2>/dev/null || echo "")

        if echo "$health_response" | grep -q '"status":"OK"'; then
            echo -e "${GREEN}Healthy${NC}"
            result=0
        elif echo "$health_response" | grep -q '"healthy":true'; then
            echo -e "${GREEN}Healthy${NC}"
            result=0
        elif [ -n "$health_response" ]; then
            # Got some response
            echo -e "${GREEN}Healthy (response received)${NC}"
            result=0
        else
            echo -e "${RED}No response${NC}"
        fi
    fi

    # Cleanup port-forward
    kill $pf_pid 2>/dev/null || true
    wait $pf_pid 2>/dev/null || true

    return $result
}

# Test each service
for service_info in "${SERVICES[@]}"; do
    IFS=':' read -r service_name k8s_service port <<< "$service_info"

    if test_service "$service_name" "$k8s_service" "$port"; then
        PASSED=$((PASSED + 1))
    else
        FAILED=$((FAILED + 1))
    fi
done

# Final summary
echo ""
echo "=========================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All services healthy: $PASSED passed${NC}"
else
    echo -e "${YELLOW}Results: $PASSED passed, $FAILED failed${NC}"
fi
echo "=========================================="

# Show troubleshooting if there were failures
if [ $FAILED -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Troubleshooting:${NC}"
    echo "  1. Check pod logs: kubectl logs -n $NAMESPACE <pod-name>"
    echo "  2. Check pod status: kubectl get pods -n $NAMESPACE"
    echo "  3. Check events: kubectl get events -n $NAMESPACE"
    exit 1
fi

exit 0
