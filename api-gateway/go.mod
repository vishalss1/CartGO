module github.com/vishalss1/CartGO/api-gateway

go 1.25.4

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/vishalss1/CartGO/pkg v0.0.0-20260409064642-f3ea5dc1de23
	golang.org/x/time v0.15.0
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect

replace github.com/vishalss1/CartGO/pkg => ../pkg
