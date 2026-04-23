CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed data (password: admin123)
INSERT INTO users (id, username, email, password_hash, role) VALUES
('usr-001', 'admin', 'admin@ecommerce.com', '$2a$10$tFo/EW/tUAZs3gbud4.xke3A0LcNPsnrUMIyhzD1ROGe6t8rnPmwm', 'admin'),
('usr-002', 'user1', 'user1@ecommerce.com', '$2a$10$tFo/EW/tUAZs3gbud4.xke3A0LcNPsnrUMIyhzD1ROGe6t8rnPmwm', 'user'),
('usr-003', 'user2', 'user2@ecommerce.com', '$2a$10$tFo/EW/tUAZs3gbud4.xke3A0LcNPsnrUMIyhzD1ROGe6t8rnPmwm', 'user');
