package domain

import (
	"time"

	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type Inventory struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Reserved  int       `json:"reserved"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (i *Inventory) Available() int {
	return i.Quantity - i.Reserved
}

func (i *Inventory) Deduct(qty int) error {
	if i.Available() < qty {
		return apperrors.ErrInsufficientStock

	}
	i.Quantity -= qty
	i.UpdatedAt = time.Now()
	return nil
}

func (i *Inventory) Reserve(qty int) error {
	if i.Available() < qty {
		return apperrors.ErrInsufficientStock
	}
	i.Reserved += qty
	i.UpdatedAt = time.Now()
	return nil
}
