# CartGO Agent Context [BOOTSTRAP]
# Status: ACTIVE_STANDARDIZED
# Modified: 2026-04-09 16:55 UTC

## 1. INFRASTRUCTURE & DISCOVERY
Topology: Monorepo with 8 containers + Postgres
Build Context: Root directory (MUST include `pkg/`)
Networks: `cartgo-network` (Docker Bridge)
Logging: Standard Go `log` package + Distributed Trace IDs (`pkg/util`)

| Service | Port | Database | URL (Internal) |
| :--- | :--- | :--- | :--- |
| api-gateway | 8080 | N/A | http://api-gateway:8080 |
| user-service | 8081 | cartgo_user_db | http://user-service:8081 |
| product-service | 8082 | cartgo_product_db | http://product-service:8082 |
| inventory-service | 8083 | cartgo_inventory_db | http://inventory-service:8083 |
| order-service | 8084 | cartgo_order_db | http://order-service:8084 |
| payment-service | 8085 | cartgo_payment_db | http://payment-service:8085 |
| delivery-service | 8086 | cartgo_delivery_db | http://delivery-service:8086 |
| support-service | 8087 | cartgo_support_db | http://support-service:8087 |

## 2. AUTHENTICATION & SECURITY
Mechanism: RS256 JWT (Asymmetric)
Issuer: `user-service` (Holds `jwt_private.pem`)
Validators: All microservices (Load `jwt_public.pem` from `/app/keys/`)

### Identity Propagation (INTER-SERVICE TRUST)
1. api-gateway STRIPS `X-User-ID` and `X-User-Role` from incoming external traffic.
2. api-gateway PROXIES `Authorization` header to downstream services.
3. Microservices validate JWT and populate internal context.
4. Internal calls (e.g. Order -> Payment) use TRUST HEADERS:
   - `X-User-ID`: Propagated UUID
   - `X-User-Role`: `SERVICE_ORDER` (or specific service role)
5. **Distributed Tracing**: `X-Correlation-ID` injected at Gateway and propagated across all services.

## 3. CORE LOGIC & STATE
- Order Flow: PENDING -> (Reserve Stock) -> (Process Payment) -> (Commit Stock) -> CONFIRMED
- Price Fetching: `order-service` fetches real price via `ProductClient` (REPLACES mock 100.0).
- Delivery: `delivery-service` auto-triggered when `order-service` status -> CONFIRMED.
- Support tickets link directly to `order_id` in database and API filters.
- **UUID Generation**: Standardized on **Go-side generation** using `pkg/util/uuid.go` (`util.GenerateUUID()`). Database schemas use `UUID PRIMARY KEY` without defaults. Repositories MUST generate and assign the ID before executing an insert. Manual assignment in Go code via the centralized utility is required for all Primary Keys.

## 4. BUILD & RUN COMMANDS
Run Cluster: `docker-compose up --build`
Rebuild Single: `docker-compose up --build -d <service-name>`
Dependency Resolution: `go.mod` uses `replace github.com/vishalss1/CartGO/pkg => ../../pkg` (or similar relative path)

## 5. RECENT REVISIONS (CRITICAL)
- REVERTED all `slog` usage to standard `log`.
- FIXED `payment-service` AuthMiddleware to support `X-User-Role` (internal headers).
- ADDED Healthchecks to `docker-compose.yml` (Services wait for `db` to be healthy).
- IMPLEMENTED Distributed Tracing: Correlation IDs propagate across major service flows.
- HARDENED Order Handling: Dynamic pricing and automatic delivery orchestration.

## 6. PENDING / MISSING
- [x] Order Service: Real Pricing & Delivery Orchestration.
- [x] Support Service: Order ID Linkage.
- [x] Observability: Distributed Tracing (Correlation IDs).
- [] Deployment: Kubernetes manifests (EKS/ALB) and Terraform infra.
- [] Observability: Prometheus/Grafana full dashboard.
