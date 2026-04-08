# CartGO — Distributed Order Processing System

## PROJECT IDENTITY

CartGO is a distributed microservices-based order processing system built in Go.
It handles product discovery, ordering, payment simulation, inventory management, delivery coordination, and user support.

---

## CORE PRINCIPLES

- Each service is **independent**
- Each service has its **own database**
- Services communicate via **HTTP only**
- No direct code or DB sharing across services
- Focus on **system behavior**, not UI
- System is **role-based**, not user-type based

---

## USER ROLES (RBAC MODEL)

All users belong to the same system but have different roles:

- CUSTOMER
  - Browse products
  - Place orders

- WAREHOUSE_STAFF
  - Manage inventory
  - Update stock levels

- DELIVERY_PARTNER
  - View assigned deliveries
  - Update delivery status

- ADMIN
  - Monitor system
  - Manage users
  - Override operations

- SUPPORT_AGENT
  - Answer user queries
  - Guide product decisions

---

## SERVICES

### 1. User Service

- Handles signup/login
- Issues JWT tokens
- Stores user roles

### 2. Product Service

- Manages product catalog
- Read-heavy service
- Used by CUSTOMER + SUPPORT_AGENT

### 3. Inventory Service

- Manages stock
- Handles reserve/release operations
- Used by WAREHOUSE_STAFF + ORDER_SERVICE
- Must prevent overselling

### 4. Order Service (CORE)

- Orchestrates full order flow
- Calls Inventory + Payment
- Handles failures and state transitions

### 5. Payment Service (Mock)

- Simulates payment success/failure
- Randomized response

### 6. Delivery Service

- Assigns deliveries
- Tracks delivery status
- Used by DELIVERY_PARTNER

### 7. Support Service

- Handles user queries
- Enables interaction between CUSTOMER and SUPPORT_AGENT

### 8. API Gateway

- Entry point for all client requests
- Handles routing + auth middleware
- Extracts role from JWT and forwards request

---

## ROLE → SERVICE ACCESS

CUSTOMER
→ Product Service (read)
→ Order Service
→ Support Service

WAREHOUSE_STAFF
→ Inventory Service

DELIVERY_PARTNER
→ Delivery Service

SUPPORT_AGENT
→ Support Service
→ Product Service (read-only)

ADMIN
→ All services (via admin endpoints)

---

## SYSTEM FLOW

Client → API Gateway
→ User Service (auth + JWT with role)
→ Target Service (based on route)

Order Flow:

Client → API Gateway → Order Service

Order Service:
→ calls Inventory Service (reserve)
→ calls Payment Service

Decision:

- if success → confirm order
- if failure → rollback (release stock)

---

## ORDER STATES

- PENDING
- CONFIRMED
- FAILED

---

## FAILURE SCENARIOS (MANDATORY)

1. Out of stock → order rejected
2. Payment failure → inventory released + order FAILED
3. Concurrent requests → only valid stock allocated

---

## ARCHITECTURE RULES

- No shared database across services
- No cross-service imports
- Communication via HTTP only
- Keep shared `pkg/` minimal
- Business logic must live in `service/` layer
- Authorization must be enforced at service level (not just gateway)

---

## SECURITY & AUTHENTICATION (RS256 JWT)

- **API Gateway**: Acts strictly as a Layer-7 router + Rate Limiter. It does **not** decode tokens or strip headers.
- **Service-Level Defense**: Every individual microservice enforces role checks in its own router utilizing the `pkg/auth` shared library.
- **Tokens**: The system relies on Asymmetric JWTs (`RS256`).
- **Identity Storage**: `user-service` is the only service that holds the `jwt_private.pem` (Issuer). All other services pull public keys dynamically via the `KEYS_DIR` environment fallback strategy.
- **Key Rotation**: Asymmetric keys can be instantly rotated by executing `./scripts/generate_keys.sh`.

---

## LOCAL DEVELOPMENT CONFIGURATION

- Local uncontainerized executions (`go run main.go`) should use `localhost` for their DB connection strings in `.env` to prevent OS host resolution failures (like `host.docker.internal` timeouts).
- `KEYS_DIR` dynamically traverses the filesystem (i.e. `../../keys`) if run from a nested service directory to prevent missing keys on startup.

---

## DIRECTORY STRUCTURE

CartGO/

- services/
- api-gateway/
- pkg/
- deployments/
- scripts/

Each service:

- cmd/

- internal/
  - handler/
  - service/
  - repository/
  - model/

- api/

- db/

---

## DEVELOPMENT ORDER

1. user-service
2. product-service
3. inventory-service
4. payment-service
5. order-service
6. delivery-service
7. support-service
8. api-gateway
9. dockerize
10. kubernetes
11. AWS deployment

---

## DATABASE STRATEGY

- Single PostgreSQL instance
- Multiple databases:
  - user_db
  - product_db
  - inventory_db
  - order_db
  - delivery_db
  - support_db

---

## NON-GOALS

- No frontend (multiple UIs can exist but not part of system)
- No real payment integration
- No unnecessary features (reviews, recommendations, etc.)

---

## PRIMARY LEARNING GOALS

- Service boundaries
- Inter-service communication
- Failure handling
- Concurrency control (inventory)
- Role-based access control (RBAC)
- Containerization (Docker)
- Orchestration (Kubernetes)
- Deployment (AWS)

---

## SUCCESS CRITERIA

- Orders correctly processed across services
- No overselling under concurrent requests
- Failures handled gracefully
- Services independently deployable
- Role-based access enforced correctly
- System runs via Docker + Kubernetes

---
