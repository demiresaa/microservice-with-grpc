.PHONY: run-infra stop-infra run-order run-payment run-inventory run-all create-topics

run-infra:
	docker-compose up -d

stop-infra:
	docker-compose down

create-topics:
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic OrderCreated
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic PaymentSuccess
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic PaymentFailed
	docker exec ecommerce-kafka kafka-topics --create --if-not-exists --bootstrap-server localhost:29092 --partitions 3 --replication-factor 1 --topic InventoryFailed

run-order:
	SERVICE_PORT=8081 go run cmd/order-service/main.go

run-payment:
	DB_PORT=5433 DB_USER=payment_user DB_PASSWORD=payment_pass DB_NAME=payment_db go run cmd/payment-service/main.go

run-inventory:
	DB_PORT=5434 DB_USER=inventory_user DB_PASSWORD=inventory_pass DB_NAME=inventory_db go run cmd/inventory-service/main.go

run-all: run-infra
	@echo "Altyapi ayaga kalkti. Servisleri ayri terminallerde calistir:"
	@echo "  make run-order"
	@echo "  make run-payment"
	@echo "  make run-inventory"
