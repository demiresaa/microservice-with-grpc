package dto

import "github.com/suleymankursatdemir/ecommerce-platform/internal/payment/domain"

type PaymentEvent struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	Amount     float64 `json:"total_price"`
}

type PaymentResult struct {
	OrderID   string               `json:"order_id"`
	ProductID string               `json:"product_id"`
	Quantity  int                  `json:"quantity"`
	Status    domain.PaymentStatus `json:"status"`
}
