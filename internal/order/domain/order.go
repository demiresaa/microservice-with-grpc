package domain

import (
	"time"

	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	ProductID  string      `json:"product_id"`
	Quantity   int         `json:"quantity"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return apperrors.ErrEmptyCustomerID
	}
	if o.ProductID == "" {
		return apperrors.ErrEmptyProductID
	}
	if o.Quantity <= 0 {
		return apperrors.ErrInvalidQuantity
	}
	if o.TotalPrice <= 0 {
		return apperrors.ErrInvalidTotalPrice
	}
	return nil
}

func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:   {OrderStatusPaid, OrderStatusCancelled, OrderStatusFailed},
		OrderStatusPaid:      {OrderStatusCompleted, OrderStatusCancelled},
		OrderStatusCancelled: {},
		OrderStatusCompleted: {},
		OrderStatusFailed:    {},
	}

	allowed, exists := transitions[o.Status]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

func (o *Order) MarkAsPaid() error {
	if !o.CanTransitionTo(OrderStatusPaid) {
		return apperrors.ErrInvalidOrderStatus
	}
	o.Status = OrderStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsCompleted() error {
	if !o.CanTransitionTo(OrderStatusCompleted) {
		return apperrors.ErrInvalidOrderStatus
	}
	o.Status = OrderStatusCompleted
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsCancelled() error {
	if !o.CanTransitionTo(OrderStatusCancelled) {
		return apperrors.ErrInvalidOrderStatus
	}
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsFailed() error {
	if !o.CanTransitionTo(OrderStatusFailed) {
		return apperrors.ErrInvalidOrderStatus
	}
	o.Status = OrderStatusFailed
	o.UpdatedAt = time.Now()
	return nil
}
