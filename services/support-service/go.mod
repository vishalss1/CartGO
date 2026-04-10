module github.com/vishalss1/CartGO/services/support-service

go 1.25.4

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.12.3
	github.com/vishalss1/CartGO/pkg v0.0.0-20260408143815-cc524d111b12
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect

replace github.com/vishalss1/CartGO/pkg => ../../pkg
