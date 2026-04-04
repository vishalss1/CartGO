# CartGO

Distributed order processing system built with Go (microservices architecture).

---

## What is this?

CartGO is a backend system that simulates how real e-commerce platforms handle:

- product browsing
- order placement
- inventory under load
- payment failures
- delivery tracking

Built to understand **microservices, concurrency, and system design** — not UI.

---

## Architecture

- independent services
- separate databases per service
- HTTP communication only
- API Gateway as entry point
- role-based access (RBAC)

---

## Services

- user-service → auth + JWT
- product-service → catalog
- inventory-service → stock + reservation
- order-service → core orchestration
- payment-service → mock payment
- delivery-service → delivery tracking
- support-service → user queries
- api-gateway → routing + auth

---

## Roles

- customer → browse + order
- warehouse_staff → manage inventory
- delivery_partner → handle deliveries
- support_agent → assist users
- admin → full system access

---

## Core Flow

order flow:

```text
client → api-gateway → order-service
        → inventory (reserve)
        → payment

success → confirm order
failure → release stock + fail order
```

---

## Key Problems Solved

- preventing overselling
- handling concurrent orders
- rollback on failure
- service-to-service communication

---

## Tech Stack

- Go
- PostgreSQL
- Docker
- Kubernetes
- AWS (deployment)

---

## Run (high level)

```bash
# start dependencies (db, etc.)
docker-compose up

# run services
make run
```

---

## Project Structure

```text
services/
api-gateway/
pkg/
deployments/
scripts/
```

---

## Why this project?

This is not a CRUD app.

It’s built to understand:

- how real systems break
- how services coordinate
- where failures happen
- how to design around them

---

## Status

in progress 🚧

---

## Notes

- no frontend
- payment is mocked
- focused on backend behavior only

---
