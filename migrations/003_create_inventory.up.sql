-- Inventory Service Migration

CREATE TABLE IF NOT EXISTS inventory (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL UNIQUE,
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved INTEGER NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_product_id ON inventory(product_id);

-- Seed data: Test urunleri
INSERT INTO inventory (id, product_id, quantity, reserved) VALUES
('inv-001', 'prod-001', 100, 0),
('inv-002', 'prod-002', 50, 0),
('inv-003', 'prod-003', 200, 0),
('inv-004', 'prod-004', 30, 0),
('inv-005', 'prod-005', 0, 0);
