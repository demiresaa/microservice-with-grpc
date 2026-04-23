# ============================================================================
# INVENTORY SERVICE - MULTI-STAGE DOCKERFILE
# ============================================================================
# Order service ile aynı yapı. Sadece binary adı ve path değişir.
# Detaylı açıklama için: cmd/order-service/Dockerfile
# ============================================================================

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /inventory-service ./cmd/inventory-service

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /inventory-service /inventory-service
CMD ["/inventory-service"]
