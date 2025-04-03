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
        CREATE TYPE order_status AS ENUM ('open', 'updated', 'closed', 'cancelled');
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
    reorder_threshold DECIMAL CHECK (reorder_threshold >= 0),
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

CREATE OR REPLACE FUNCTION log_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO order_status_history(order_id, status, created_at)
        VALUES (NEW.id, NEW.status, NOW());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER order_status_change_trigger
AFTER UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION log_order_status_change();

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

CREATE OR REPLACE FUNCTION log_inventory_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.quantity IS DISTINCT FROM NEW.quantity THEN
        INSERT INTO inventory_transaction(inventory_id, transaction_type, quantity, notes, created_at)
        VALUES (NEW.id, 
                CASE 
                    WHEN NEW.quantity > OLD.quantity THEN 'restock'::transaction_type 
                    ELSE 'use'::transaction_type 
                END,
                ABS(NEW.quantity - OLD.quantity),
                'Auto update from inventory change',
                NOW());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER inventory_change_trigger
AFTER UPDATE ON inventory
FOR EACH ROW
EXECUTE FUNCTION log_inventory_change();

CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    menu_item_id INT REFERENCES menu_items(id) ON DELETE CASCADE,
    price DECIMAL(10, 2) NOT NULL,
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    change_reason TEXT
);

CREATE OR REPLACE FUNCTION log_price_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.price IS DISTINCT FROM NEW.price THEN
        -- Закрываем старую запись
        UPDATE price_history
        SET effective_to = NOW()
        WHERE menu_item_id = NEW.id AND effective_to IS NULL;

        -- Добавляем новую запись
        INSERT INTO price_history(menu_item_id, price, effective_from, change_reason)
        VALUES (NEW.id, NEW.price, NOW(), 'Auto update from menu_items');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER price_change_trigger
AFTER UPDATE OF price ON menu_items
FOR EACH ROW
EXECUTE FUNCTION log_price_change();

-- Создание индексов
-- Индекс для поиска по категориям меню
CREATE INDEX idx_menu_items_categories ON menu_items USING GIN (categories);

-- Индекс для поиска по описанию меню (полнотекстовый поиск)
CREATE INDEX idx_menu_items_description_tsvector ON menu_items USING GIN (to_tsvector('russian', description));

-- Индекс для быстрого поиска заказов по статусу
CREATE INDEX idx_orders_status ON orders (status);

-- Составной индекс для поиска заказов за определенный период
CREATE INDEX idx_orders_created_status ON orders (created_at, status);

-- Индекс для быстрого поиска ингредиентов, которые скоро закончатся
CREATE INDEX idx_inventory_reorder ON inventory (reorder_threshold, quantity);

-- Индекс для поиска транзакций по типу
CREATE INDEX idx_inventory_transaction_type ON inventory_transaction (transaction_type);

INSERT INTO inventory (ingredient_name, quantity, unit, reorder_threshold, updated_at) VALUES
('Coffee beans', 10.0, 'kg', 2.0, NOW()),
('Milk', 25.0, 'l', 5.0, NOW()),
('Sugar', 8.0, 'kg', 1.5, NOW()),
('Chocolate syrup', 5.0, 'l', 1.0, NOW()),
('Vanilla syrup', 4.5, 'l', 1.0, NOW()),
('Caramel syrup', 4.0, 'l', 1.0, NOW()),
('Cream', 10.0, 'l', 2.0, NOW()),
('Cocoa powder', 3.0, 'kg', 0.5, NOW()),
('Cinnamon', 1.0, 'kg', 0.2, NOW()),
('Water', 100.0, 'l', 20.0, NOW()),
('Black tea', 2.0, 'kg', 0.5, NOW()),
('Green tea', 1.5, 'kg', 0.5, NOW()),
('Mint', 0.5, 'kg', 0.1, NOW()),
('Lemon', 3.0, 'kg', 0.7, NOW()),
('Ginger', 1.0, 'kg', 0.2, NOW()),
('Honey', 2.5, 'kg', 0.5, NOW()),
('Whipped cream', 3.0, 'l', 1.0, NOW()),
('Soy milk', 5.0, 'l', 1.0, NOW()),
('Almond milk', 5.0, 'l', 1.0, NOW()),
('Coconut milk', 4.0, 'l', 1.0, NOW());

-- Menu population
INSERT INTO menu_items (name, description, price, categories, created_at, updated_at) VALUES
('Espresso', 'Classic Italian coffee, prepared under pressure', 3.50, ARRAY['coffee', 'hot drinks', 'classics'], NOW() - INTERVAL '6 months', NOW()),
('Cappuccino', 'Espresso with milk foam', 4.50, ARRAY['coffee', 'hot drinks', 'milk drinks'], NOW() - INTERVAL '6 months', NOW()),
('Latte', 'Espresso with milk and a small amount of milk foam', 4.75, ARRAY['coffee', 'hot drinks', 'milk drinks'], NOW() - INTERVAL '6 months', NOW()),
('Americano', 'Espresso diluted with hot water', 3.75, ARRAY['coffee', 'hot drinks', 'classics'], NOW() - INTERVAL '6 months', NOW()),
('Mocha', 'Espresso with milk and chocolate syrup', 5.25, ARRAY['coffee', 'hot drinks', 'sweet drinks', 'chocolate'], NOW() - INTERVAL '5 months', NOW()),
('Frappuccino', 'Cold coffee-based drink with milk and ice', 5.50, ARRAY['coffee', 'cold drinks', 'smoothies'], NOW() - INTERVAL '4 months', NOW()),
('Green tea', 'Classic green tea', 3.00, ARRAY['tea', 'hot drinks', 'classics'], NOW() - INTERVAL '5 months', NOW()),
('Black tea', 'Classic black tea', 3.00, ARRAY['tea', 'hot drinks', 'classics'], NOW() - INTERVAL '5 months', NOW()),
('Ginger tea with honey', 'Black tea with ginger and honey', 4.00, ARRAY['tea', 'hot drinks', 'specials'], NOW() - INTERVAL '3 months', NOW()),
('Iced latte', 'Cold espresso-based drink with milk and ice', 5.00, ARRAY['coffee', 'cold drinks', 'milk drinks'], NOW() - INTERVAL '3 months', NOW());

-- Menu and ingredients relationship
INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity) VALUES
(1, 1, 0.02), -- Espresso - coffee beans
(2, 1, 0.02), -- Cappuccino - coffee beans
(2, 2, 0.1),  -- Cappuccino - milk
(3, 1, 0.02), -- Latte - coffee beans
(3, 2, 0.2),  -- Latte - milk
(4, 1, 0.02), -- Americano - coffee beans
(4, 10, 0.1), -- Americano - water
(5, 1, 0.02), -- Mocha - coffee beans
(5, 2, 0.15), -- Mocha - milk
(5, 4, 0.03), -- Mocha - chocolate syrup
(6, 1, 0.02), -- Frappuccino - coffee beans
(6, 2, 0.15), -- Frappuccino - milk
(6, 17, 0.05),-- Frappuccino - whipped cream
(7, 12, 0.01),-- Green tea
(8, 11, 0.01),-- Black tea
(9, 11, 0.01),-- Ginger tea - black tea
(9, 15, 0.01),-- Ginger tea - ginger
(9, 16, 0.02),-- Ginger tea - honey
(10, 1, 0.02),-- Iced latte - coffee beans
(10, 2, 0.2); -- Iced latte - milk

-- Price history
INSERT INTO price_history (menu_item_id, price, effective_from, effective_to, change_reason) VALUES
(1, 3.00, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Initial price'),
(1, 3.50, NOW() - INTERVAL '6 months', NULL, 'Increase due to rising coffee bean prices'),
(2, 4.00, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Initial price'),
(2, 4.50, NOW() - INTERVAL '6 months', NULL, 'Increase due to rising milk prices'),
(3, 4.25, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Initial price'),
(3, 4.75, NOW() - INTERVAL '6 months', NULL, 'Increase due to rising milk prices'),
(4, 3.25, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Initial price'),
(4, 3.75, NOW() - INTERVAL '6 months', NULL, 'Increase due to rising coffee bean prices'),
(5, 5.00, NOW() - INTERVAL '12 months', NOW() - INTERVAL '5 months', 'Initial price'),
(5, 5.25, NOW() - INTERVAL '5 months', NULL, 'Increase due to rising chocolate syrup prices');

-- Orders
INSERT INTO orders (name, status, total_amount, special_instructions, created_at, updated_at) VALUES
('Anna', 'closed', 9.00, '{"sugar": "no sugar"}', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
('Ivan', 'closed', 4.50, '{"milk": "soy"}', NOW() - INTERVAL '29 days', NOW() - INTERVAL '29 days'),
('Maria', 'closed', 8.25, '{}', NOW() - INTERVAL '28 days', NOW() - INTERVAL '28 days'),
('Peter', 'closed', 3.50, '{"sugar": "2 teaspoons"}', NOW() - INTERVAL '27 days', NOW() - INTERVAL '27 days'),
('Elena', 'closed', 13.75, '{"milk": "almond", "syrup": "double portion"}', NOW() - INTERVAL '26 days', NOW() - INTERVAL '26 days'),
('Alexander', 'closed', 5.25, '{}', NOW() - INTERVAL '25 days', NOW() - INTERVAL '25 days'),
('Olga', 'closed', 9.50, '{"sugar": "no sugar"}', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
('Dmitry', 'closed', 7.50, '{}', NOW() - INTERVAL '23 days', NOW() - INTERVAL '23 days'),
('Natalia', 'closed', 12.00, '{"ice": "no ice"}', NOW() - INTERVAL '22 days', NOW() - INTERVAL '22 days'),
('Sergey', 'closed', 5.50, '{}', NOW() - INTERVAL '21 days', NOW() - INTERVAL '21 days'),
('Julia', 'closed', 8.75, '{"milk": "soy"}', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),
('Andrey', 'closed', 6.00, '{}', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
('Ekaterina', 'closed', 9.25, '{"cinnamon": "with cinnamon"}', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
('Maxim', 'closed', 4.75, '{}', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days'),
('Tatiana', 'closed', 7.25, '{"sugar": "1 teaspoon"}', NOW() - INTERVAL '16 days', NOW() - INTERVAL '16 days'),
('Alexey', 'closed', 10.50, '{}', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
('Irina', 'closed', 6.50, '{"milk": "coconut"}', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),
('Vladimir', 'closed', 9.00, '{}', NOW() - INTERVAL '13 days', NOW() - INTERVAL '13 days'),
('Svetlana', 'closed', 5.25, '{"sugar": "2 teaspoons"}', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
('Denis', 'closed', 8.50, '{}', NOW() - INTERVAL '11 days', NOW() - INTERVAL '11 days'),
('Anastasia', 'closed', 7.75, '{"milk": "almond"}', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
('Igor', 'closed', 5.00, '{}', NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),
('Lubov', 'closed', 9.75, '{"sugar": "no sugar"}', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
('Viktor', 'closed', 4.50, '{}', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
('Marina', 'closed', 6.25, '{"cinnamon": "with cinnamon"}', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
('Evgeny', 'updated', 8.00, '{}', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('Ksenia', 'closed', 5.75, '{"milk": "soy"}', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
('Artem', 'open', 7.00, '{}', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
('Galina', 'open', 10.25, '{"sugar": "1 teaspoon"}', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
('Nikolay', 'cancelled', 6.75, '{}', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Order items
INSERT INTO order_items (order_id, menu_item_id, quantity, price) VALUES
(1, 1, 1, 3.50), (1, 2, 1, 4.50), (1, 7, 1, 3.00),
(2, 2, 1, 4.50),
(3, 3, 1, 4.75), (3, 8, 1, 3.00),
(4, 1, 1, 3.50),
(5, 3, 1, 4.75), (5, 5, 1, 5.25), (5, 7, 1, 3.00),
(6, 5, 1, 5.25),
(7, 3, 2, 4.75),
(8, 2, 1, 4.50), (8, 7, 1, 3.00),
(9, 3, 1, 4.75), (9, 6, 1, 5.50),
(10, 6, 1, 5.50),
(11, 3, 1, 4.75), (11, 4, 1, 3.75),
(12, 2, 1, 4.50),
(13, 3, 1, 4.75), (13, 5, 1, 5.25),
(14, 3, 1, 4.75),
(15, 1, 1, 3.50), (15, 7, 1, 3.00),
(16, 3, 1, 4.75), (16, 6, 1, 5.50),
(17, 10, 1, 5.00),
(18, 1, 1, 3.50), (18, 8, 1, 3.00), (18, 9, 1, 4.00),
(19, 5, 1, 5.25),
(20, 2, 1, 4.50), (20, 4, 1, 3.75),
(21, 3, 1, 4.75), (21, 7, 1, 3.00),
(22, 10, 1, 5.00),
(23, 5, 1, 5.25), (23, 9, 1, 4.00),
(24, 2, 1, 4.50),
(25, 9, 1, 4.00), (25, 10, 1, 5.00),
(26, 4, 1, 3.75), (26, 5, 1, 5.25),
(27, 3, 1, 4.75),
(28, 1, 2, 3.50),
(29, 3, 1, 4.75), (29, 6, 1, 5.50),
(30, 10, 1, 5.00);

-- Order status history
INSERT INTO order_status_history (order_id, status, notes, created_at) VALUES
(1, 'open', 'Order received', NOW() - INTERVAL '30 days' - INTERVAL '10 minutes'),
(1, 'closed', 'Order completed', NOW() - INTERVAL '30 days'),
(2, 'open', 'Order received', NOW() - INTERVAL '29 days' - INTERVAL '15 minutes'),
(2, 'closed', 'Order completed', NOW() - INTERVAL '29 days'),
(3, 'open', 'Order received', NOW() - INTERVAL '28 days' - INTERVAL '12 minutes'),
(3, 'closed', 'Order completed', NOW() - INTERVAL '28 days'),
(4, 'open', 'Order received', NOW() - INTERVAL '27 days' - INTERVAL '8 minutes'),
(4, 'closed', 'Order completed', NOW() - INTERVAL '27 days'),
(5, 'open', 'Order received', NOW() - INTERVAL '26 days' - INTERVAL '20 minutes'),
(5, 'closed', 'Order completed', NOW() - INTERVAL '26 days'),
(6, 'open', 'Order received', NOW() - INTERVAL '25 days' - INTERVAL '11 minutes'),
(6, 'closed', 'Order completed', NOW() - INTERVAL '25 days'),
(7, 'open', 'Order received', NOW() - INTERVAL '24 days' - INTERVAL '9 minutes'),
(7, 'closed', 'Order completed', NOW() - INTERVAL '24 days'),
(8, 'open', 'Order received', NOW() - INTERVAL '23 days' - INTERVAL '14 minutes'),
(8, 'closed', 'Order completed', NOW() - INTERVAL '23 days'),
(9, 'open', 'Order received', NOW() - INTERVAL '22 days' - INTERVAL '16 minutes'),
(9, 'closed', 'Order completed', NOW() - INTERVAL '22 days'),
(10, 'open', 'Order received', NOW() - INTERVAL '21 days' - INTERVAL '7 minutes'),
(10, 'closed', 'Order completed', NOW() - INTERVAL '21 days'),
(11, 'open', 'Order received', NOW() - INTERVAL '20 days' - INTERVAL '13 minutes'),
(11, 'closed', 'Order completed', NOW() - INTERVAL '20 days'),
(12, 'open', 'Order received', NOW() - INTERVAL '19 days' - INTERVAL '10 minutes'),
(12, 'closed', 'Order completed', NOW() - INTERVAL '19 days'),
(13, 'open', 'Order received', NOW() - INTERVAL '18 days' - INTERVAL '15 minutes'),
(13, 'closed', 'Order completed', NOW() - INTERVAL '18 days'),
(14, 'open', 'Order received', NOW() - INTERVAL '17 days' - INTERVAL '8 minutes'),
(14, 'closed', 'Order completed', NOW() - INTERVAL '17 days'),
(15, 'open', 'Order received', NOW() - INTERVAL '16 days' - INTERVAL '12 minutes'),
(15, 'closed', 'Order completed', NOW() - INTERVAL '16 days'),
(16, 'open', 'Order received', NOW() - INTERVAL '15 days' - INTERVAL '18 minutes'),
(16, 'closed', 'Order completed', NOW() - INTERVAL '15 days'),
(17, 'open', 'Order received', NOW() - INTERVAL '14 days' - INTERVAL '9 minutes'),
(17, 'closed', 'Order completed', NOW() - INTERVAL '14 days'),
(18, 'open', 'Order received', NOW() - INTERVAL '13 days' - INTERVAL '16 minutes'),
(18, 'closed', 'Order completed', NOW() - INTERVAL '13 days'),
(19, 'open', 'Order received', NOW() - INTERVAL '12 days' - INTERVAL '7 minutes'),
(19, 'closed', 'Order completed', NOW() - INTERVAL '12 days'),
(20, 'open', 'Order received', NOW() - INTERVAL '11 days' - INTERVAL '10 minutes'),
(20, 'closed', 'Order completed', NOW() - INTERVAL '11 days'),
(21, 'open', 'Order received', NOW() - INTERVAL '10 days' - INTERVAL '14 minutes'),
(21, 'closed', 'Order completed', NOW() - INTERVAL '10 days'),
(22, 'open', 'Order received', NOW() - INTERVAL '9 days' - INTERVAL '8 minutes'),
(22, 'closed', 'Order completed', NOW() - INTERVAL '9 days'),
(23, 'open', 'Order received', NOW() - INTERVAL '8 days' - INTERVAL '15 minutes'),
(23, 'closed', 'Order completed', NOW() - INTERVAL '8 days'),
(24, 'open', 'Order received', NOW() - INTERVAL '7 days' - INTERVAL '9 minutes'),
(24, 'closed', 'Order completed', NOW() - INTERVAL '7 days'),
(25, 'open', 'Order received', NOW() - INTERVAL '6 days' - INTERVAL '12 minutes'),
(25, 'closed', 'Order completed', NOW() - INTERVAL '6 days'),
(26, 'open', 'Order received', NOW() - INTERVAL '5 days' - INTERVAL '20 minutes'),
(26, 'updated', 'Order composition changed', NOW() - INTERVAL '5 days'),
(27, 'open', 'Order received', NOW() - INTERVAL '4 days' - INTERVAL '7 minutes'),
(27, 'closed', 'Order completed', NOW() - INTERVAL '4 days'),
(28, 'open', 'Order received', NOW() - INTERVAL '3 days'),
(29, 'open', 'Order received', NOW() - INTERVAL '2 days'),
(30, 'open', 'Order received', NOW() - INTERVAL '1 day' - INTERVAL '30 minutes'),
(30, 'cancelled', 'Customer cancelled the order', NOW() - INTERVAL '1 day');

-- Inventory transactions
INSERT INTO inventory_transaction (inventory_id, transaction_type, quantity, notes, created_at) VALUES
(1, 'restock', 5.0, 'Scheduled coffee beans restock', NOW() - INTERVAL '31 days'),
(2, 'restock', 10.0, 'Scheduled milk restock', NOW() - INTERVAL '31 days'),
(3, 'restock', 3.0, 'Scheduled sugar restock', NOW() - INTERVAL '31 days'),
(4, 'restock', 2.0, 'Scheduled chocolate syrup restock', NOW() - INTERVAL '31 days'),
(1, 'use', -0.6, 'Used for coffee preparation', NOW() - INTERVAL '30 days'),
(2, 'use', -2.5, 'Used for coffee with milk preparation', NOW() - INTERVAL '30 days'),
(3, 'use', -0.3, 'Used for sweetening drinks', NOW() - INTERVAL '30 days'),
(1, 'use', -0.5, 'Used for coffee preparation', NOW() - INTERVAL '29 days'),
(2, 'use', -2.0, 'Used for coffee with milk preparation', NOW() - INTERVAL '29 days'),
(4, 'use', -0.2, 'Used chocolate syrup', NOW() - INTERVAL '29 days'),
(1, 'use', -0.4, 'Used for coffee preparation', NOW() - INTERVAL '28 days'),
(2, 'use', -1.8, 'Used for coffee with milk preparation', NOW() - INTERVAL '28 days'),
(1, 'adjustment', -0.1, 'Adjustment after inventory check', NOW() - INTERVAL '27 days'),
(5, 'restock', 2.0, 'Vanilla syrup restock', NOW() - INTERVAL '26 days'),
(6, 'restock', 1.5, 'Caramel syrup restock', NOW() - INTERVAL '26 days'),
(1, 'use', -0.5, 'Used for coffee preparation', NOW() - INTERVAL '25 days'),
(2, 'use', -2.2, 'Used for coffee with milk preparation', NOW() - INTERVAL '25 days'),
(5, 'use', -0.15, 'Used vanilla syrup', NOW() - INTERVAL '25 days'),
(1, 'use', -0.45, 'Used for coffee preparation', NOW() - INTERVAL '24 days'),
(2, 'use', -1.9, 'Used for coffee with milk preparation', NOW() - INTERVAL '24 days'),
(3, 'use', -0.25, 'Used for sweetening drinks', NOW() - INTERVAL '24 days'),
(1, 'restock', 4.0, 'Scheduled coffee beans restock', NOW() - INTERVAL '23 days'),
(2, 'restock', 8.0, 'Scheduled milk restock', NOW() - INTERVAL '23 days'),
(1, 'use', -0.4, 'Used for coffee preparation', NOW() - INTERVAL '22 days'),
(2, 'use', -1.7, 'Used for coffee with milk preparation', NOW() - INTERVAL '22 days'),
(6, 'use', -0.1, 'Used caramel syrup', NOW() - INTERVAL '22 days'),
(11, 'use', -0.05, 'Used black tea', NOW() - INTERVAL '21 days'),
(12, 'use', -0.04, 'Used green tea', NOW() - INTERVAL '21 days'),
(19, 'restock', 2.0, 'Almond milk restock', NOW() - INTERVAL '20 days'),
(20, 'restock', 1.5, 'Coconut milk restock', NOW() - INTERVAL '20 days'),
(18, 'use', -0.3, 'Used soy milk', NOW() - INTERVAL '19 days'),
(19, 'use', -0.2, 'Used almond milk', NOW() - INTERVAL '19 days');