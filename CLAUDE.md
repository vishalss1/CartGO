# CartGO — Distributed Order Processing System

## PROJECT IDENTITY

CartGO is a distributed microservices-based order processing system built in Go.
It handles product discovery, ordering, payment simulation, and inventory management under constrained stock conditions.

---

## CORE PRINCIPLES

- Each service is **independent**
- Each service has its **own database**
- Services communicate via **HTTP only**
- No direct code or DB sharing across services
- Focus on **system behavior**, not UI

---

## SERVICES

### 1. User Service

- Handles signup/login
- Issues JWT tokens

### 2. Product Service

- Manages product catalog
- Read-heavy service

### 3. Inventory Service

- Manages stock
- Handles reserve/release operations
- Must prevent overselling

### 4. Order Service (CORE)

- Orchestrates full order flow
- Calls Inventory + Payment
- Handles failures and state transitions

### 5. Payment Service (Mock)

- Simulates payment success/failure
- Randomized response

### 6. API Gateway

- Entry point for all client requests
- Handles routing + auth middleware

---

## SYSTEM FLOW

Client → API Gateway
→ User Service (auth)
→ Product Service (browse)
→ Order Service

Order Service:

- calls Inventory Service (reserve)
- calls Payment Service

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
6. api-gateway
7. dockerize
8. kubernetes
9. AWS deployment

---

## DATABASE STRATEGY

- Single PostgreSQL instance
- Multiple databases:
  - user_db
  - product_db
  - inventory_db
  - order_db

---

## NON-GOALS

- No frontend
- No advanced auth system
- No real payment integration
- No unnecessary features (cart, reviews, etc.)

---

## PRIMARY LEARNING GOALS

- Service boundaries
- Inter-service communication
- Failure handling
- Concurrency control (inventory)
- Containerization (Docker)
- Orchestration (Kubernetes)
- Deployment (AWS)

---

## SUCCESS CRITERIA

- Orders correctly processed across services
- No overselling under concurrent requests
- Failures handled gracefully
- Services independently deployable
- System runs via Docker + Kubernetes

---
