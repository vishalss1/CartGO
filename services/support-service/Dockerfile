# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY pkg ./pkg

WORKDIR /src/services/support-service
COPY services/support-service/go.mod services/support-service/go.sum ./
RUN go mod download

COPY services/support-service .

RUN CGO_ENABLED=0 GOOS=linux go build -o support-service ./cmd/main.go


# Run Stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /src/services/support-service/support-service .

EXPOSE 8087

CMD ["./support-service"]