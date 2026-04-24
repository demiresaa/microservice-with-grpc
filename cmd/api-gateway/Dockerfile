FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /api-gateway ./cmd/api-gateway

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /api-gateway /api-gateway

CMD ["/api-gateway"]
