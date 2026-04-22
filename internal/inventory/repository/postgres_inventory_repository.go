package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/inventory/domain"
	apperrors "github.com/suleymankursatdemir/ecommerce-platform/pkg/errors"
)

type postgresInventoryRepository struct {
	db *sql.DB
}

func NewPostgresInventoryRepository(db *sql.DB) InventoryRepository {
	return &postgresInventoryRepository{db: db}
}

func (r *postgresInventoryRepository) GetByProductID(ctx context.Context, productID string) (*domain.Inventory, error) {
	query := `
		SELECT id, product_id, quantity, reserved, created_at, updated_at
		FROM inventory
		WHERE product_id = $1
	`

	var inv domain.Inventory
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&inv.ID,
		&inv.ProductID,
		&inv.Quantity,
		&inv.Reserved,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(apperrors.ErrProductNotFound, err)
		}
		return nil, fmt.Errorf("failed to get inventory by product_id: %w", err)
	}

	return &inv, nil
}

func (r *postgresInventoryRepository) DeductStock(ctx context.Context, productID string, quantity int) error {
	query := `
		UPDATE inventory
		SET quantity = quantity - $1, updated_at = NOW()
		WHERE product_id = $2 AND (quantity - reserved) >= $1
	`

	result, err := r.db.ExecContext(ctx, query, quantity, productID)
	if err != nil {
		return fmt.Errorf("failed to deduct stock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrInsufficientStock
	}

	return nil
}

func (r *postgresInventoryRepository) ReserveStock(ctx context.Context, productID string, quantity int) error {
	query := `
		UPDATE inventory
		SET reserved = reserved + $1, updated_at = NOW()
		WHERE product_id = $2 AND (quantity - reserved) >= $1
	`

	result, err := r.db.ExecContext(ctx, query, quantity, productID)
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return apperrors.ErrInsufficientStock
	}

	return nil
}
