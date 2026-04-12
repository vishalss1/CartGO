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
- **COMPLETED Frontend Integration**: React SPA fully connected to all 7 backend microservices.
- **RESOLVED CORS & API Mismatch**: Implemented Vite Proxying and systemic response unwrapping (`{ success, data, error }`).
- **ENFORCED Clean UX**: Stripped all debug technical text and internal system paths from user-facing pages.

## 6. PENDING / MISSING
- [x] Order Service: Real Pricing & Delivery Orchestration.
- [x] Support Service: Order ID Linkage.
- [x] Observability: Distributed Tracing (Correlation IDs).
- [x] Frontend: Full backend integration with role-based UI.
- [] Deployment: Kubernetes manifests (EKS/ALB) and Terraform infra.
- [] Observability: Prometheus/Grafana full dashboard.

## 7. FRONTEND ARCHITECTURE [IMPLEMENTED]
### Tech Stack
- **Framework**: React 18 + Vite 5 (SPA).
- **Routing**: react-router-dom v6 with role-based protected routes.
- **Styling**: TailwindCSS 3 with custom design tokens (surface, paper, accent, line, muted).
- **Fonts**: Inter (Google Fonts).

### File Structure
```
frontend/src/
├── main.jsx              # Entry: BrowserRouter > AuthProvider > ToastProvider > App
├── App.jsx               # Route definitions (/, /login, /register, /shop, /inventory, /delivery, /admin, /support)
├── index.css             # Tailwind directives + global resets
├── context/
│   ├── AuthContext.jsx   # JWT auth state, login(), register(), logout(), backendStatus
│   └── ToastContext.jsx  # Global toast notifications (showSuccess, showError, showInfo)
├── components/
│   ├── Navbar.jsx        # Role-aware navigation bar with sign-out
│   ├── ProtectedRoute.jsx # Auth guard + role guard (redirects to /login or correct role route)
│   ├── StatusBanner.jsx  # Inline status banner (success/error/default tones)
│   └── Toast.jsx         # Toast notification UI (auto-dismiss 4s, slide animation)
├── layout/
│   └── MainLayout.jsx    # Navbar + main content wrapper with sanitized gateway status
├── pages/
│   ├── Login.jsx         # Email/password login, links to /register
│   ├── Register.jsx      # Username/email/password/role registration
│   ├── RoleRedirect.jsx  # Redirects authenticated users to their role route
│   ├── Shop.jsx          # CUSTOMER: product catalog, search, cart, checkout, order history, payment retry
│   ├── Inventory.jsx     # WAREHOUSE_STAFF: stock levels, +1/-1 adjust, low-stock alerts
│   ├── Delivery.jsx      # DELIVERY_PARTNER: assigned deliveries, status updates
│   ├── Admin.jsx         # ADMIN: user management, role assignment, product CRUD, order inspection
│   └── Support.jsx       # SUPPORT_AGENT: ticket list, filters, status updates, message thread + reply
└── utils/
    ├── api.js            # Centralized fetch wrapper: apiRequest(), authenticatedRequest(), 401 auto-logout, error sanitization
    ├── auth.js           # localStorage session management (token, refreshToken, user, cart)
    └── constants.js      # API_BASE_URL, GATEWAY_HEALTH_URL, ROLE_TO_ROUTE, VALID_ROLES, status enums
```

### Auth Flow
1. Login: `POST /api/v1/user/login` → receives `{ access_token, refresh_token, user }`.
2. Register: `POST /api/v1/user/register` → creates account, redirects to login.
3. Token stored in `localStorage`; attached as `Authorization: Bearer <token>` on all authenticated requests.
4. On 401 response: auto-clear session + redirect to `/login`.
5. Role-based redirect after login: CUSTOMER→/shop, WAREHOUSE_STAFF→/inventory, DELIVERY_PARTNER→/delivery, ADMIN→/admin, SUPPORT_AGENT→/support.

### API Layer (utils/api.js)
- **Base URL**: Relative path `/api/v1` (Proxied via Vite in development).
- **Vite Proxy**: Configured in `vite.config.js` to forward `/api/v1` and `/health` to `http://localhost:8080`.
- **Response Unwrapping**: Every backend response is wrapped in `{ success: boolean, data: any, error: string }`. The API layer automatically unwraps `data` and throws `error`.
- **Error Sanitization**: All backend error messages are sanitized before display. Internal paths, service names, and technical jargon are stripped. Users see clean messages only.
- **401 Handling**: Automatic session clear + redirect to `/login` for authenticated requests. 401s on login/register endpoints are treated as "Invalid credentials" and displayed to the user.
- **403 Handling**: Generic "permission denied" message.
- **Network Errors**: Generic "unable to connect" message.

### UX Rules (ENFORCED)
- **NO backend endpoints** displayed anywhere in UI.
- **NO service names** (product-service, order-service, etc.) in user-facing text.
- **NO technical descriptions** of architecture, middleware, or RBAC in any page copy.
- **Toast notifications** for all success/error feedback (not inline raw errors).
- **User-friendly messages only** throughout all role interfaces.

### API Endpoint Mapping (Frontend → Gateway)
| Feature | Method | Path |
|---|---|---|
| Login | POST | /api/v1/user/login |
| Register | POST | /api/v1/user/register |
| Current user | GET | /api/v1/user/me |
| List users (admin) | GET | /api/v1/user/admin/users |
| Update role (admin) | PATCH | /api/v1/user/admin/users/:id/role |
| List products | GET | /api/v1/products |
| Get product | GET | /api/v1/products/:id |
| Create product (admin) | POST | /api/v1/products |
| Delete product (admin) | DELETE | /api/v1/products/:id |
| Get inventory | GET | /api/v1/inventory/:product_id |
| Adjust stock | POST | /api/v1/inventory/:product_id/adjust |
| Create order | POST | /api/v1/orders |
| User orders | GET | /api/v1/orders/user/:user_id |
| Process payment | POST | /api/v1/payments |
| Partner deliveries | GET | /api/v1/deliveries/partner/:partner_id |
| Update delivery status | PATCH | /api/v1/deliveries/:id/status |
| List tickets | GET | /api/v1/support/tickets |
| Update ticket status | PATCH | /api/v1/support/tickets/:id/status |
| Ticket messages | GET | /api/v1/support/tickets/:id/messages |
| Add ticket message | POST | /api/v1/support/tickets/:id/messages |
