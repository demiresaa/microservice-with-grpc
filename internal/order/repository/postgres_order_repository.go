package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/order/domain"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type postgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, product_id, quantity, total_price, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.ProductID,
		order.Quantity,
		order.TotalPrice,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *postgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, product_id, quantity, total_price, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ProductID,
		&order.Quantity,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(apperrors.ErrOrderNotFound, err)
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	return &order, nil
}

func (r *postgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `
		UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrOrderNotFound
	}

	return nil
}
