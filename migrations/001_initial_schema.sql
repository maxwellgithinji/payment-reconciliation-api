-- Initial schema migration for Insurance Portal POC
-- Run with: psql -U postgres -d insurance_portal -f migrations/001_initial_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'accounts_manager')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255)
);

-- Customers table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    date_of_birth DATE NOT NULL,
    street VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255)
);

-- Policies table
CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    product_name VARCHAR(255) NOT NULL,
    premium_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'KES',
    payment_frequency VARCHAR(20) NOT NULL CHECK (payment_frequency IN ('monthly', 'quarterly', 'annually')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'lapsed', 'expired', 'canceled')),
    next_payment_due TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255)
);

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE RESTRICT,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'KES',
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('mobile_money', 'card', 'bank_transfer')),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    transaction_id VARCHAR(255) UNIQUE NOT NULL,
    payment_reference VARCHAR(100) UNIQUE NOT NULL,
    payment_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    failure_reason TEXT,
    is_reconciled BOOLEAN DEFAULT false,
    reconciliation_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255)
);

-- Reconciliations table
CREATE TABLE reconciliations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE RESTRICT,
    expected_amount BIGINT NOT NULL,
    received_amount BIGINT NOT NULL,
    variance BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'KES',
    status VARCHAR(50) NOT NULL CHECK (status IN ('matched', 'unmatched', 'partial', 'disputed')),
    match_result VARCHAR(50) NOT NULL CHECK (match_result IN ('perfect', 'amount_mismatch', 'policy_mismatch', 'no_match')),
    auto_reconciled BOOLEAN DEFAULT false,
    manually_reviewed BOOLEAN DEFAULT false,
    reviewed_by VARCHAR(255),
    reviewed_at TIMESTAMP,
    notes TEXT,
    reconciled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255)
);

-- Indexes for performance
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_status ON customers(status);

CREATE INDEX idx_policies_customer_id ON policies(customer_id);
CREATE INDEX idx_policies_status ON policies(status);
CREATE INDEX idx_policies_next_payment_due ON policies(next_payment_due);
CREATE INDEX idx_policies_policy_number ON policies(policy_number);

CREATE INDEX idx_payments_policy_id ON payments(policy_id);
CREATE INDEX idx_payments_customer_id ON payments(customer_id);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_is_reconciled ON payments(is_reconciled);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);

CREATE INDEX idx_reconciliations_payment_id ON reconciliations(payment_id);
CREATE INDEX idx_reconciliations_policy_id ON reconciliations(policy_id);
CREATE INDEX idx_reconciliations_status ON reconciliations(status);
CREATE INDEX idx_reconciliations_manually_reviewed ON reconciliations(manually_reviewed);

-- Add foreign key for reconciliation_id in payments
ALTER TABLE payments
ADD CONSTRAINT fk_payments_reconciliation
FOREIGN KEY (reconciliation_id) REFERENCES reconciliations(id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reconciliations_updated_at BEFORE UPDATE ON reconciliations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries
CREATE VIEW overdue_policies AS
SELECT 
    p.*,
    c.first_name,
    c.last_name,
    c.email,
    c.phone_number
FROM policies p
JOIN customers c ON p.customer_id = c.id
WHERE p.status = 'active'
AND p.next_payment_due < CURRENT_TIMESTAMP;

CREATE VIEW unreconciled_payments AS
SELECT 
    p.*,
    pol.policy_number,
    pol.premium_amount as expected_amount
FROM payments p
JOIN policies pol ON p.policy_id = pol.id
WHERE p.status = 'completed'
AND p.is_reconciled = false;

CREATE VIEW reconciliation_summary AS
SELECT 
    r.*,
    p.payment_reference,
    p.transaction_id,
    pol.policy_number,
    c.first_name || ' ' || c.last_name as customer_name
FROM reconciliations r
JOIN payments p ON r.payment_id = p.id
JOIN policies pol ON r.policy_id = pol.id
JOIN customers c ON pol.customer_id = c.id;