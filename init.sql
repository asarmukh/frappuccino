-- database: :memory:
DROP TABLE IF EXISTS orders CASCADE;

DROP TABLE IF EXISTS customers CASCADE;

-- Проверяем, существует ли ENUM order_status, чтобы не пытаться создать его дважды
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM ('open', 'close');
    END IF;
END $$;

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id) ON DELETE SET NULL,
    status order_status NOT NULL DEFAULT 'open',
    total_amount DECIMAL(10, 2) NOT NULL CHECK (total_amount >= 0), -- Для денег
    special_instructions TEXT,
    payment_method VARCHAR(50) DEFAULT 'credit card',
    is_completed BOOLEAN NOT NULL DEFAULT FALSE, -- Быстрый фильтр завершённых заказов
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);