# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY pkg ./pkg

WORKDIR /src/services/support-service
COPY services/support-service/go.mod services/support-service/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    sh -c '\
    for i in 1 2 3; do \
      echo "Go mod download attempt $i/3"; \
      GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org \
      timeout 300 go mod download && break; \
      if [ $i -lt 3 ]; then \
        echo "Attempt $i failed, waiting before retry..."; \
        sleep $((10 * i)); \
      fi; \
    done'

COPY services/support-service .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    sh -c '\
    for i in 1 2 3; do \
      echo "Go build attempt $i/3"; \
      GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org \
      timeout 300 go build -o support-service ./cmd/main.go && break; \
      if [ $i -lt 3 ]; then \
        echo "Attempt $i failed, cleaning cache and retrying..."; \
        go clean -modcache; \
        sleep $((10 * i)); \
      fi; \
    done'


# Run Stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /src/services/support-service/support-service .

EXPOSE 8087

CMD ["./support-service"]