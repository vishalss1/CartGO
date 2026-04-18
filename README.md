# CartGO - Distributed E-Commerce Order Processing System

A production-grade microservices architecture demonstrating distributed order processing, built with Go and PostgreSQL. Handles product catalog, inventory management, order orchestration, payments, delivery tracking, and customer support across independent services with separate databases and inter-service communication.

---

## Architecture Overview

CartGO is built on a **distributed microservices architecture** with 8 independent services, each with its own PostgreSQL database. Services communicate via HTTP with JWT-based authentication. The API Gateway serves as the single entry point, routing requests to appropriate services and enforcing role-based access control.

```
┌─────────────────┐
│   API Gateway   │ (HTTP routing + auth)
└────────┬────────┘
         │
    ┌────┴────┬──────────┬─────────────┬────────────┬─────────────┬───────────┐
    │          │          │             │            │             │           │
┌───▼──┐  ┌────▼──┐  ┌───▼────┐  ┌────▼────┐  ┌───▼────┐  ┌────▼───┐  ┌───▼───┐
│User  │  │Product│  │Inventory│  │ Order   │  │ Payment│  │Delivery │  │Support│
│Svc   │  │Svc    │  │Svc      │  │ Svc     │  │ Svc    │  │Svc      │  │Svc    │
└──┬───┘  └───┬───┘  └──┬──────┘  └───┬────┘  └──┬─────┘  └────┬────┘  └───┬───┘
   │          │         │             │          │              │           │
   └─────┬────┴────┬────┴─────┬──────┴──────┬───┴──────┬───────┴───────┬──┘
         │         │          │             │          │               │
      ┌──▼──┐  ┌───▼───┐  ┌───▼────┐  ┌────▼────┐ ┌──▼─────┐  ┌────▼──┐
      │User │  │Product│  │Inventory│  │ Order   │ │Payment │  │Delivery│
      │ DB  │  │ DB    │  │ DB      │  │ DB      │ │ DB     │  │ DB     │
      └─────┘  └───────┘  └─────────┘  └─────────┘ └────────┘  └────────┘
```

---

## Services

### User Service (Port 8081)
- User registration and authentication
- JWT token generation (RS256 asymmetric)
- User profile management
- Role assignment and RBAC
- Public key distribution for token validation

### Product Service (Port 8082)
- Product catalog browsing
- Product creation and management
- Product search and filtering
- Price information

### Inventory Service (Port 8083)
- Real-time stock levels
- Inventory reservations (prevents overselling)
- Stock commitment and rollback
- Low-stock alerts
- Inventory adjustments (admin)

### Order Service (Port 8084)
- Order creation and orchestration
- Order status tracking (PENDING → CONFIRMED → SHIPPED → DELIVERED)
- Order history per user
- Integrates with Inventory and Payment services
- Triggers delivery service on confirmation

### Payment Service (Port 8085)
- Payment processing and capture
- Payment status tracking
- Refund handling
- Mock payment gateway integration
- Transaction history

### Delivery Service (Port 8086)
- Delivery partner management
- Delivery assignment and tracking
- Status updates (ASSIGNED → PICKED → IN_TRANSIT → DELIVERED)
- Auto-triggered on order confirmation
- Partner delivery management

### Support Service (Port 8087)
- Support ticket creation and management
- Ticket-to-order linking
- Ticket status tracking
- Message threading
- Support agent assignment

### API Gateway (Port 8080)
- Single entry point for all clients
- JWT token validation
- Request routing to microservices
- CORS handling
- Rate limiting
- Request/response logging
- Correlation ID propagation for tracing

---

## Key Features

### Distributed Architecture
- **Service Independence**: Each service owns its database and business logic
- **Loose Coupling**: Services communicate only via HTTP APIs
- **Independent Scaling**: Services scale independently based on load
- **Failure Isolation**: Service failures don't cascade to other services

### Order Processing Flow
1. Client places order via API Gateway
2. Order Service receives request
3. Inventory Service reserves stock (fails if unavailable)
4. Payment Service processes payment (fails if insufficient funds)
5. On success: Inventory commits reservation, Order confirmed
6. On failure: Inventory releases reservation, Order cancelled
7. Delivery Service auto-triggered for confirmed orders

### Security
- **JWT Authentication** (RS256 asymmetric - private key in user-service, public keys distributed)
- **Role-Based Access Control** (RBAC) - CUSTOMER, WAREHOUSE_STAFF, DELIVERY_PARTNER, SUPPORT_AGENT, ADMIN
- **Service-to-Service Trust Headers** (X-User-ID, X-User-Role for internal calls)
- **Distributed Correlation IDs** (X-Correlation-ID for tracing across services)

### Data Consistency
- **Distributed Transactions**: Order flow implements saga pattern
- **Inventory Reservations**: Prevent overselling under concurrent load
- **Payment Idempotency**: Duplicate payment requests handled safely
- **Rollback Capability**: Failed transactions trigger compensating actions

---

## Technology Stack

- **Language**: Go 1.25
- **Database**: PostgreSQL 15 (per-service instances)
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Kubernetes (Minikube for local development)
- **Ingress**: Nginx Ingress Controller
- **Frontend**: React 18 + Vite (separate SPA)
- **Authentication**: RS256 JWT
- **HTTP Framework**: Go std library + custom routing

---

## Project Structure

```
CartGO/
├── api-gateway/              # Entry point & routing
│   ├── cmd/
│   ├── internal/
│   │   ├── config/
│   │   ├── middleware/
│   │   ├── proxy/
│   │   └── router/
│   └── go.mod
│
├── services/
│   ├── user-service/         # Authentication & users
│   ├── product-service/      # Product catalog
│   ├── inventory-service/    # Stock management
│   ├── order-service/        # Order orchestration
│   ├── payment-service/      # Payment processing
│   ├── delivery-service/     # Delivery tracking
│   └── support-service/      # Customer support
│
├── pkg/                       # Shared utilities
│   ├── auth/                 # JWT handling
│   ├── util/                 # UUID, correlation IDs, responses
│   └── go.mod
│
├── frontend/                  # React SPA
│   ├── src/
│   │   ├── pages/           # Route pages
│   │   ├── components/      # React components
│   │   ├── context/         # Auth & Toast state
│   │   └── utils/           # API, auth, constants
│   └── package.json
│
├── docker-compose.yml        # Local development setup
├── go.work                   # Go workspace file
└── README.md                 # This file
```

---

## Database Schema

Each service has its own PostgreSQL database:

- **cartgo_user_db**: Users, roles, authentication
- **cartgo_product_db**: Products, pricing, catalog
- **cartgo_inventory_db**: Stock levels, reservations
- **cartgo_order_db**: Orders, order items, status
- **cartgo_payment_db**: Transactions, payment status
- **cartgo_delivery_db**: Deliveries, partner assignments
- **cartgo_support_db**: Tickets, messages, assignments

---

## API Authentication

All endpoints (except public ones: Login, Register, List Products, Get Product) require JWT Bearer token:

```bash
curl -H "Authorization: Bearer <token>" http://api-gateway:8080/api/v1/user/me
```

Token obtained via `/api/v1/user/login` endpoint.

---

## Inter-Service Communication

Services communicate via HTTP with service discovery using Kubernetes DNS (in k8s) or docker-compose service names (locally):

```go
// Example: Order service calling Inventory service
client := &http.Client{}
req, _ := http.NewRequest("GET", "http://inventory-service:8083/api/v1/inventory/123", nil)
req.Header.Add("X-User-ID", userID)
req.Header.Add("X-User-Role", "SERVICE_ORDER")
req.Header.Add("X-Correlation-ID", correlationID)
resp, _ := client.Do(req)
```

### Trust Headers (for internal calls)
- `X-User-ID`: User making the original request
- `X-User-Role`: Role of the user (services use SERVICE_* roles internally)
- `X-Correlation-ID`: Trace ID across distributed calls

---

## Development

### Prerequisites
- Go 1.25+
- PostgreSQL 15+
- Docker & Docker Compose
- Kubernetes & Minikube (for k8s deployment)

### Local Development

Run all services with PostgreSQL via Docker Compose:

```bash
docker-compose up
```

This starts:
- 7 microservices (user, product, inventory, order, payment, delivery, support)
- 1 API Gateway
- PostgreSQL database with all 7 service databases initialized

Each service listens on its assigned port (8081-8087). API Gateway listens on 8080.

### Environment Variables

Each service reads from environment variables (set via docker-compose or locally):

```
DATABASE_URL=postgres://user:pass@localhost:5432/dbname
JWT_PRIVATE_KEY_PATH=/app/keys/jwt_private.pem (user-service only)
JWT_PUBLIC_KEY_PATH=/app/keys/jwt_public.pem (other services)
```

---

## Testing

### Health Checks
```bash
curl http://api-gateway:8080/health
```

### Create User & Get Token
```bash
curl -X POST http://api-gateway:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"pass","role":"CUSTOMER"}'

curl -X POST http://api-gateway:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"pass"}'
```

### Place Order
```bash
curl -X POST http://api-gateway:8080/api/v1/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":"123","quantity":2}]}'
```

---

## Key Design Patterns

- **API Gateway Pattern**: Single entry point for all clients
- **Database Per Service**: Data isolation and independent scaling
- **Service Registry**: Kubernetes DNS or docker-compose networking
- **Distributed Tracing**: Correlation IDs propagate across services
- **Saga Pattern**: Distributed transactions via compensating actions
- **Circuit Breaker Ready**: Service failures don't cascade
- **Role-Based Access Control**: Fine-grained authorization per endpoint

---

## Common Issues

### Service Connection Issues
- Ensure all services are running: `docker-compose ps`
- Check service logs: `docker-compose logs <service-name>`
- Verify network connectivity between services

### Database Issues
- Clear and reinitialize: `docker-compose down -v && docker-compose up`
- Check PostgreSQL logs: `docker-compose logs postgres`

### JWT Token Issues
- Ensure user-service is running (generates tokens)
- Token expires after configured duration (default 1 hour)
- Invalid tokens return 401 responses

---

## Future Improvements

- Kubernetes deployment manifests (k8s/ directory)
- Message queue integration (RabbitMQ, Kafka) for async processing
- Caching layer (Redis) for product catalog
- Metrics & monitoring (Prometheus, Grafana)
- Distributed logging (ELK stack)
- Service mesh (Istio) for advanced traffic management
- GraphQL API layer
- Pagination for list endpoints
- Advanced search and filtering

---

## Contributing

CartGO is a learning project demonstrating microservices architecture. Contributions welcome for improvements, bug fixes, or additional features.

---

## License

MIT License

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
