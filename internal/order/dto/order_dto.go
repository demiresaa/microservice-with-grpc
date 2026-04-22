package dto

import "github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	CustomerID string  `json:"customer_id" example:"cust-001"`
	ProductID  string  `json:"product_id" example:"prod-001"`
	Quantity   int     `json:"quantity" example:"2"`
	TotalPrice float64 `json:"total_price" example:"99.90"`
}

// OrderResponse represents the response body for an order
type OrderResponse struct {
	ID         string             `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CustomerID string             `json:"customer_id" example:"cust-001"`
	ProductID  string             `json:"product_id" example:"prod-001"`
	Quantity   int                `json:"quantity" example:"2"`
	TotalPrice float64            `json:"total_price" example:"99.90"`
	Status     domain.OrderStatus `json:"status" example:"PENDING"`
	CreatedAt  string             `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt  string             `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type OrderEvent struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
}
