# ============================================================================
# MAKEFILE - BUILD & RUN KOMUTLARI
# ============================================================================
#
# KULLANIM:
#   make build-all       → 3 servisi de container olarak build et
#   make up              → Tüm stack'i (altyapı + servisler) ayağa kaldır
#   make down            → Her şeyi durdur
#   make build-SERVICE   → Tek bir servisi build et
#   make logs-SERVICE    → Tek bir servisin loglarını izle
#
# DOCKER MULTI-STAGE BUILD SÜRECİ:
#   make build-order → Dockerfile'ı okur:
#     Stage 1 (builder): golang:1.25-alpine → go mod download + go build
#     Stage 2 (runtime): distroless/static → sadece binary kopyalanır
#   Sonuç: ~15-20MB image (golang image ~300MB olurdu)
# ============================================================================

.PHONY: build-all build-order build-payment build-inventory build-gateway \
        up down stop-infra run-infra \
        up-services down-services \
        logs-order logs-payment logs-inventory logs-gateway \
        create-topics ps clean

# ---- BUILD KOMUTLARI ----

build-order:
	docker build -f cmd/order-service/Dockerfile -t ecommerce-order-service:latest .

build-payment:
	docker build -f cmd/payment-service/Dockerfile -t ecommerce-payment-service:latest .

build-inventory:
	docker build -f cmd/inventory-service/Dockerfile -t ecommerce-inventory-service:latest .

build-gateway:
	docker build -f cmd/api-gateway/Dockerfile -t ecommerce-api-gateway:latest .

build-all: build-order build-payment build-inventory build-gateway
	@echo "✅ Tüm servisler build edildi"

# ---- DOCKER COMPOSE KOMUTLARI ----

# Sadece altyapı (DB + Kafka) - servis geliştirirken kullan
run-infra:
	docker-compose up -d zookeeper kafka kafka-ui order-db payment-db inventory-db

# Tüm stack (altyapı + servisler) - build sonrası kullan
up: build-all
	docker-compose up -d
	@echo "✅ Tüm stack ayağa kalktı"

# Servisler hariç altyapıyı kaldır
up-infra:
	docker-compose up -d zookeeper kafka kafka-ui order-db payment-db inventory-db
	@echo "✅ Altyapı ayağa kalktı (servisler hariç)"

# Sadece servisleri kaldır (altyapı zaten ayakta olmalı)
up-services: build-all
	docker-compose up -d order-service payment-service inventory-service api-gateway
	@echo "✅ Servisler ayağa kalktı"

# Her şeyi durdur
down:
	docker-compose down

stop-infra:
	docker-compose down

# Servisleri durdur (altyapı çalışmaya devam eder)
down-services:
	docker-compose stop order-service payment-service inventory-service api-gateway

# ---- LOG KOMUTLARI ----

logs-order:
	docker-compose logs -f order-service

logs-payment:
	docker-compose logs -f payment-service

logs-inventory:
	docker-compose logs -f inventory-service

logs-gateway:
	docker-compose logs -f api-gateway

# ---- KAFFA TOPIC OLUŞTURMA ----

create-topics:
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic OrderCreated
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic PaymentSuccess
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic PaymentFailed
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic InventoryFailed

# ---- YARDIMCI KOMUTLAR ----

ps:
	docker-compose ps

# Container'ları + volume'ları sil (temiz başlangıç)
clean:
	docker-compose down -v
	@echo "✅ Container'lar ve volume'lar silindi"

# ---- YEREL GELİŞTİRME (Docker olmadan) ----

run-order:
	SERVICE_PORT=8081 go run cmd/order-service/main.go

run-payment:
	DB_PORT=5433 DB_USER=payment_user DB_PASSWORD=payment_pass DB_NAME=payment_db go run cmd/payment-service/main.go

run-inventory:
	DB_PORT=5434 DB_USER=inventory_user DB_PASSWORD=inventory_pass DB_NAME=inventory_db GRPC_PORT=50051 go run cmd/inventory-service/main.go

run-gateway:
	GATEWAY_PORT=8000 ORDER_SERVICE_URL=http://localhost:8081 go run cmd/api-gateway/main.go

run-all: run-infra
	@echo "Altyapi ayaga kalkti. Servisleri ayri terminallerde calistir:"
	@echo "  make run-order"
	@echo "  make run-payment"
	@echo "  make run-inventory"

# ---- TEST ----

test:
	go test ./... -count=1

test-verbose:
	go test ./... -v -count=1

test-coverage:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage raporu: coverage.html"
