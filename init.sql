DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS order_status_history CASCADE;
DROP TABLE IF EXISTS menu_items CASCADE;
DROP TABLE IF EXISTS menu_item_ingredients CASCADE;
DROP TABLE IF EXISTS price_history CASCADE;
DROP TABLE IF EXISTS inventory CASCADE;
DROP TABLE IF EXISTS inventory_transaction CASCADE;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM ('open', 'updated', 'close', 'cancelled');
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'temperature_type') THEN
        CREATE TYPE temperature_type AS ENUM ('hot', 'warm', 'cold');
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_type') THEN
        CREATE TYPE transaction_type AS ENUM ('restock', 'use', 'adjustment');
    END IF;
END $$;

CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    categories VARCHAR[],
    -- available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    status order_status NOT NULL DEFAULT 'open',
    total_amount DECIMAL(10, 2) NOT NULL CHECK (total_amount >= 0),
    special_instructions JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    ingredient_name VARCHAR(50) NOT NULL UNIQUE,
    quantity DECIMAL NOT NULL CHECK (quantity >= 0),
    unit VARCHAR(50) NOT NULL,
    reorder_threshold DECIMAL CHECK (reorder_threshold >= 0), -- Порог перезаказа
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE order_items (
    order_id INT REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id INT REFERENCES menu_items(id) ON DELETE CASCADE,
    quantity INT CHECK(quantity > 0),
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    PRIMARY KEY (order_id, menu_item_id)
);

CREATE TABLE order_status_history (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id) ON DELETE CASCADE,
    status order_status NOT NULL DEFAULT 'open',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE menu_item_ingredients (
    menu_item_id INT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    ingredient_id INT NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    quantity DECIMAL NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (menu_item_id, ingredient_id)
);

CREATE TABLE inventory_transaction (
    id SERIAL PRIMARY KEY,
    inventory_id INT REFERENCES inventory(id) ON DELETE CASCADE,
    transaction_type transaction_type NOT NULL DEFAULT 'use',
    quantity DECIMAL NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    menu_item_id INT REFERENCES menu_items(id) ON DELETE CASCADE,
    price DECIMAL(10, 2) NOT NULL,
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    change_reason TEXT
);
