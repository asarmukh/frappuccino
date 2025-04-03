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

-- Добавление тестовых данных
-- Наполнение инвентаря
INSERT INTO inventory (ingredient_name, quantity, unit, reorder_threshold, updated_at) VALUES
('Кофейные зерна', 10.0, 'кг', 2.0, NOW()),
('Молоко', 25.0, 'л', 5.0, NOW()),
('Сахар', 8.0, 'кг', 1.5, NOW()),
('Сироп шоколадный', 5.0, 'л', 1.0, NOW()),
('Сироп ванильный', 4.5, 'л', 1.0, NOW()),
('Сироп карамельный', 4.0, 'л', 1.0, NOW()),
('Сливки', 10.0, 'л', 2.0, NOW()),
('Какао-порошок', 3.0, 'кг', 0.5, NOW()),
('Корица', 1.0, 'кг', 0.2, NOW()),
('Вода', 100.0, 'л', 20.0, NOW()),
('Чай черный', 2.0, 'кг', 0.5, NOW()),
('Чай зеленый', 1.5, 'кг', 0.5, NOW()),
('Мята', 0.5, 'кг', 0.1, NOW()),
('Лимон', 3.0, 'кг', 0.7, NOW()),
('Имбирь', 1.0, 'кг', 0.2, NOW()),
('Мед', 2.5, 'кг', 0.5, NOW()),
('Взбитые сливки', 3.0, 'л', 1.0, NOW()),
('Соевое молоко', 5.0, 'л', 1.0, NOW()),
('Миндальное молоко', 5.0, 'л', 1.0, NOW()),
('Кокосовое молоко', 4.0, 'л', 1.0, NOW());

-- Наполнение меню
INSERT INTO menu_items (name, description, price, categories, created_at, updated_at) VALUES
('Эспрессо', 'Классический итальянский кофе, приготовленный под давлением', 3.50, ARRAY['кофе', 'горячие напитки', 'классика'], NOW() - INTERVAL '6 months', NOW()),
('Капучино', 'Эспрессо с добавлением молочной пенки', 4.50, ARRAY['кофе', 'горячие напитки', 'молочные напитки'], NOW() - INTERVAL '6 months', NOW()),
('Латте', 'Эспрессо с добавлением молока и небольшого количества молочной пенки', 4.75, ARRAY['кофе', 'горячие напитки', 'молочные напитки'], NOW() - INTERVAL '6 months', NOW()),
('Американо', 'Эспрессо, разбавленный горячей водой', 3.75, ARRAY['кофе', 'горячие напитки', 'классика'], NOW() - INTERVAL '6 months', NOW()),
('Мокко', 'Эспрессо с добавлением молока и шоколадного сиропа', 5.25, ARRAY['кофе', 'горячие напитки', 'сладкие напитки', 'шоколад'], NOW() - INTERVAL '5 months', NOW()),
('Фраппучино', 'Холодный напиток на основе кофе с молоком и льдом', 5.50, ARRAY['кофе', 'холодные напитки', 'смузи'], NOW() - INTERVAL '4 months', NOW()),
('Зеленый чай', 'Классический зеленый чай', 3.00, ARRAY['чай', 'горячие напитки', 'классика'], NOW() - INTERVAL '5 months', NOW()),
('Черный чай', 'Классический черный чай', 3.00, ARRAY['чай', 'горячие напитки', 'классика'], NOW() - INTERVAL '5 months', NOW()),
('Имбирный чай с медом', 'Черный чай с добавлением имбиря и меда', 4.00, ARRAY['чай', 'горячие напитки', 'специальные'], NOW() - INTERVAL '3 months', NOW()),
('Айс латте', 'Холодный напиток на основе эспрессо с молоком и льдом', 5.00, ARRAY['кофе', 'холодные напитки', 'молочные напитки'], NOW() - INTERVAL '3 months', NOW());

-- Связь меню и ингредиентов
INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity) VALUES
(1, 1, 0.02), -- Эспрессо - кофейные зерна
(2, 1, 0.02), -- Капучино - кофейные зерна
(2, 2, 0.1),  -- Капучино - молоко
(3, 1, 0.02), -- Латте - кофейные зерна
(3, 2, 0.2),  -- Латте - молоко
(4, 1, 0.02), -- Американо - кофейные зерна
(4, 10, 0.1), -- Американо - вода
(5, 1, 0.02), -- Мокко - кофейные зерна
(5, 2, 0.15), -- Мокко - молоко
(5, 4, 0.03), -- Мокко - шоколадный сироп
(6, 1, 0.02), -- Фраппучино - кофейные зерна
(6, 2, 0.15), -- Фраппучино - молоко
(6, 17, 0.05),-- Фраппучино - взбитые сливки
(7, 12, 0.01),-- Зеленый чай
(8, 11, 0.01),-- Черный чай
(9, 11, 0.01),-- Имбирный чай - черный чай
(9, 15, 0.01),-- Имбирный чай - имбирь
(9, 16, 0.02),-- Имбирный чай - мед
(10, 1, 0.02),-- Айс латте - кофейные зерна
(10, 2, 0.2); -- Айс латте - молоко

-- История цен
INSERT INTO price_history (menu_item_id, price, effective_from, effective_to, change_reason) VALUES
(1, 3.00, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Начальная цена'),
(1, 3.50, NOW() - INTERVAL '6 months', NULL, 'Повышение из-за роста цен на кофейные зерна'),
(2, 4.00, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Начальная цена'),
(2, 4.50, NOW() - INTERVAL '6 months', NULL, 'Повышение из-за роста цен на молоко'),
(3, 4.25, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Начальная цена'),
(3, 4.75, NOW() - INTERVAL '6 months', NULL, 'Повышение из-за роста цен на молоко'),
(4, 3.25, NOW() - INTERVAL '12 months', NOW() - INTERVAL '6 months', 'Начальная цена'),
(4, 3.75, NOW() - INTERVAL '6 months', NULL, 'Повышение из-за роста цен на кофейные зерна'),
(5, 5.00, NOW() - INTERVAL '12 months', NOW() - INTERVAL '5 months', 'Начальная цена'),
(5, 5.25, NOW() - INTERVAL '5 months', NULL, 'Повышение из-за роста цен на шоколадный сироп');

-- Заказы
INSERT INTO orders (name, status, total_amount, special_instructions, created_at, updated_at) VALUES
('Анна', 'closed', 9.00, '{"сахар": "без сахара"}', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
('Иван', 'closed', 4.50, '{"молоко": "соевое"}', NOW() - INTERVAL '29 days', NOW() - INTERVAL '29 days'),
('Мария', 'closed', 8.25, '{}', NOW() - INTERVAL '28 days', NOW() - INTERVAL '28 days'),
('Петр', 'closed', 3.50, '{"сахар": "2 ложки"}', NOW() - INTERVAL '27 days', NOW() - INTERVAL '27 days'),
('Елена', 'closed', 13.75, '{"молоко": "миндальное", "сироп": "двойная порция"}', NOW() - INTERVAL '26 days', NOW() - INTERVAL '26 days'),
('Александр', 'closed', 5.25, '{}', NOW() - INTERVAL '25 days', NOW() - INTERVAL '25 days'),
('Ольга', 'closed', 9.50, '{"сахар": "без сахара"}', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
('Дмитрий', 'closed', 7.50, '{}', NOW() - INTERVAL '23 days', NOW() - INTERVAL '23 days'),
('Наталья', 'closed', 12.00, '{"лед": "без льда"}', NOW() - INTERVAL '22 days', NOW() - INTERVAL '22 days'),
('Сергей', 'closed', 5.50, '{}', NOW() - INTERVAL '21 days', NOW() - INTERVAL '21 days'),
('Юлия', 'closed', 8.75, '{"молоко": "соевое"}', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),
('Андрей', 'closed', 6.00, '{}', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
('Екатерина', 'closed', 9.25, '{"корица": "с корицей"}', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
('Максим', 'closed', 4.75, '{}', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days'),
('Татьяна', 'closed', 7.25, '{"сахар": "1 ложка"}', NOW() - INTERVAL '16 days', NOW() - INTERVAL '16 days'),
('Алексей', 'closed', 10.50, '{}', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
('Ирина', 'closed', 6.50, '{"молоко": "кокосовое"}', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),
('Владимир', 'closed', 9.00, '{}', NOW() - INTERVAL '13 days', NOW() - INTERVAL '13 days'),
('Светлана', 'closed', 5.25, '{"сахар": "2 ложки"}', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
('Денис', 'closed', 8.50, '{}', NOW() - INTERVAL '11 days', NOW() - INTERVAL '11 days'),
('Анастасия', 'closed', 7.75, '{"молоко": "миндальное"}', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
('Игорь', 'closed', 5.00, '{}', NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),
('Любовь', 'closed', 9.75, '{"сахар": "без сахара"}', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
('Виктор', 'closed', 4.50, '{}', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
('Марина', 'closed', 6.25, '{"корица": "с корицей"}', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
('Евгений', 'updated', 8.00, '{}', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('Ксения', 'closed', 5.75, '{"молоко": "соевое"}', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
('Артем', 'open', 7.00, '{}', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
('Галина', 'open', 10.25, '{"сахар": "1 ложка"}', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
('Николай', 'cancelled', 6.75, '{}', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Позиции заказов
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

-- История статусов заказов
INSERT INTO order_status_history (order_id, status, notes, created_at) VALUES
(1, 'open', 'Заказ принят', NOW() - INTERVAL '30 days' - INTERVAL '10 minutes'),
(1, 'closed', 'Заказ выполнен', NOW() - INTERVAL '30 days'),
(2, 'open', 'Заказ принят', NOW() - INTERVAL '29 days' - INTERVAL '15 minutes'),
(2, 'closed', 'Заказ выполнен', NOW() - INTERVAL '29 days'),
(3, 'open', 'Заказ принят', NOW() - INTERVAL '28 days' - INTERVAL '12 minutes'),
(3, 'closed', 'Заказ выполнен', NOW() - INTERVAL '28 days'),
(4, 'open', 'Заказ принят', NOW() - INTERVAL '27 days' - INTERVAL '8 minutes'),
(4, 'closed', 'Заказ выполнен', NOW() - INTERVAL '27 days'),
(5, 'open', 'Заказ принят', NOW() - INTERVAL '26 days' - INTERVAL '20 minutes'),
(5, 'closed', 'Заказ выполнен', NOW() - INTERVAL '26 days'),
(6, 'open', 'Заказ принят', NOW() - INTERVAL '25 days' - INTERVAL '11 minutes'),
(6, 'closed', 'Заказ выполнен', NOW() - INTERVAL '25 days'),
(7, 'open', 'Заказ принят', NOW() - INTERVAL '24 days' - INTERVAL '9 minutes'),
(7, 'closed', 'Заказ выполнен', NOW() - INTERVAL '24 days'),
(8, 'open', 'Заказ принят', NOW() - INTERVAL '23 days' - INTERVAL '14 minutes'),
(8, 'closed', 'Заказ выполнен', NOW() - INTERVAL '23 days'),
(9, 'open', 'Заказ принят', NOW() - INTERVAL '22 days' - INTERVAL '16 minutes'),
(9, 'closed', 'Заказ выполнен', NOW() - INTERVAL '22 days'),
(10, 'open', 'Заказ принят', NOW() - INTERVAL '21 days' - INTERVAL '7 minutes'),
(10, 'closed', 'Заказ выполнен', NOW() - INTERVAL '21 days'),
(11, 'open', 'Заказ принят', NOW() - INTERVAL '20 days' - INTERVAL '13 minutes'),
(11, 'closed', 'Заказ выполнен', NOW() - INTERVAL '20 days'),
(12, 'open', 'Заказ принят', NOW() - INTERVAL '19 days' - INTERVAL '10 minutes'),
(12, 'closed', 'Заказ выполнен', NOW() - INTERVAL '19 days'),
(13, 'open', 'Заказ принят', NOW() - INTERVAL '18 days' - INTERVAL '15 minutes'),
(13, 'closed', 'Заказ выполнен', NOW() - INTERVAL '18 days'),
(14, 'open', 'Заказ принят', NOW() - INTERVAL '17 days' - INTERVAL '8 minutes'),
(14, 'closed', 'Заказ выполнен', NOW() - INTERVAL '17 days'),
(15, 'open', 'Заказ принят', NOW() - INTERVAL '16 days' - INTERVAL '12 minutes'),
(15, 'closed', 'Заказ выполнен', NOW() - INTERVAL '16 days'),
(16, 'open', 'Заказ принят', NOW() - INTERVAL '15 days' - INTERVAL '18 minutes'),
(16, 'closed', 'Заказ выполнен', NOW() - INTERVAL '15 days'),
(17, 'open', 'Заказ принят', NOW() - INTERVAL '14 days' - INTERVAL '9 minutes'),
(17, 'closed', 'Заказ выполнен', NOW() - INTERVAL '14 days'),
(18, 'open', 'Заказ принят', NOW() - INTERVAL '13 days' - INTERVAL '16 minutes'),
(18, 'closed', 'Заказ выполнен', NOW() - INTERVAL '13 days'),
(19, 'open', 'Заказ принят', NOW() - INTERVAL '12 days' - INTERVAL '7 minutes'),
(19, 'closed', 'Заказ выполнен', NOW() - INTERVAL '12 days'),
(20, 'open', 'Заказ принят', NOW() - INTERVAL '11 days' - INTERVAL '10 minutes'),
(20, 'closed', 'Заказ выполнен', NOW() - INTERVAL '11 days'),
(21, 'open', 'Заказ принят', NOW() - INTERVAL '10 days' - INTERVAL '14 minutes'),
(21, 'closed', 'Заказ выполнен', NOW() - INTERVAL '10 days'),
(22, 'open', 'Заказ принят', NOW() - INTERVAL '9 days' - INTERVAL '8 minutes'),
(22, 'closed', 'Заказ выполнен', NOW() - INTERVAL '9 days'),
(23, 'open', 'Заказ принят', NOW() - INTERVAL '8 days' - INTERVAL '15 minutes'),
(23, 'closed', 'Заказ выполнен', NOW() - INTERVAL '8 days'),
(24, 'open', 'Заказ принят', NOW() - INTERVAL '7 days' - INTERVAL '9 minutes'),
(24, 'closed', 'Заказ выполнен', NOW() - INTERVAL '7 days'),
(25, 'open', 'Заказ принят', NOW() - INTERVAL '6 days' - INTERVAL '12 minutes'),
(25, 'closed', 'Заказ выполнен', NOW() - INTERVAL '6 days'),
(26, 'open', 'Заказ принят', NOW() - INTERVAL '5 days' - INTERVAL '20 minutes'),
(26, 'updated', 'Изменен состав заказа', NOW() - INTERVAL '5 days'),
(27, 'open', 'Заказ принят', NOW() - INTERVAL '4 days' - INTERVAL '7 minutes'),
(27, 'closed', 'Заказ выполнен', NOW() - INTERVAL '4 days'),
(28, 'open', 'Заказ принят', NOW() - INTERVAL '3 days'),
(29, 'open', 'Заказ принят', NOW() - INTERVAL '2 days'),
(30, 'open', 'Заказ принят', NOW() - INTERVAL '1 day' - INTERVAL '30 minutes'),
(30, 'cancelled', 'Клиент отменил заказ', NOW() - INTERVAL '1 day');

-- Транзакции инвентаря
INSERT INTO inventory_transaction (inventory_id, transaction_type, quantity, notes, created_at) VALUES
(1, 'restock', 5.0, 'Плановое пополнение кофейных зерен', NOW() - INTERVAL '31 days'),
(2, 'restock', 10.0, 'Плановое пополнение молока', NOW() - INTERVAL '31 days'),
(3, 'restock', 3.0, 'Плановое пополнение сахара', NOW() - INTERVAL '31 days'),
(4, 'restock', 2.0, 'Плановое пополнение шоколадного сиропа', NOW() - INTERVAL '31 days'),
(1, 'use', -0.6, 'Расход на приготовление кофе', NOW() - INTERVAL '30 days'),
(2, 'use', -2.5, 'Расход на приготовление кофе с молоком', NOW() - INTERVAL '30 days'),
(3, 'use', -0.3, 'Расход на подслащение напитков', NOW() - INTERVAL '30 days'),
(1, 'use', -0.5, 'Расход на приготовление кофе', NOW() - INTERVAL '29 days'),
(2, 'use', -2.0, 'Расход на приготовление кофе с молоком', NOW() - INTERVAL '29 days'),
(4, 'use', -0.2, 'Расход шоколадного сиропа', NOW() - INTERVAL '29 days'),
(1, 'use', -0.4, 'Расход на приготовление кофе', NOW() - INTERVAL '28 days'),
(2, 'use', -1.8, 'Расход на приготовление кофе с молоком', NOW() - INTERVAL '28 days'),
(1, 'adjustment', -0.1, 'Корректировка после инвентаризации', NOW() - INTERVAL '27 days'),
(5, 'restock', 2.0, 'Пополнение ванильного сиропа', NOW() - INTERVAL '26 days'),
(6, 'restock', 1.5, 'Пополнение карамельного сиропа', NOW() - INTERVAL '26 days'),
(1, 'use', -0.5, 'Расход на приготовление кофе', NOW() - INTERVAL '25 days'),
(2, 'use', -2.2, 'Расход на приготовление кофе с молоком', NOW() - INTERVAL '25 days'),
(5, 'use', -0.15, 'Расход ванильного сиропа', NOW() - INTERVAL '25 days'),
(1, 'use', -0.45, 'Расход на приготовление кофе', NOW() - INTERVAL '24 days'),
(2, 'use', -1.9, 'Расход на приготовление кофе с молоком', NOW() - INTERVAL '24 days'),
(3, 'use', -0.25, 'Расход на подслащение напитков', NOW() - INTERVAL '24 days'),
(1, 'restock', 4.0, 'Плановое пополнение кофейных зерен', NOW() - INTERVAL '23 days'),
(2, 'restock', 8.0, 'Плановое пополнение молока', NOW() - INTERVAL '23 days'),
(1, 'use', -0.4, 'Расход на приготовление кофе', NOW() - INTERVAL '22 days'),
(2, 'use', -1.7, 'Расход на приготовление кофе с молоком', NOW() - INTERVAL '22 days'),
(6, 'use', -0.1, 'Расход карамельного сиропа', NOW() - INTERVAL '22 days'),
(11, 'use', -0.05, 'Расход черного чая', NOW() - INTERVAL '21 days'),
(12, 'use', -0.04, 'Расход зеленого чая', NOW() - INTERVAL '21 days'),
(19, 'restock', 2.0, 'Пополнение миндального молока', NOW() - INTERVAL '20 days'),
(20, 'restock', 1.5, 'Пополнение кокосового молока', NOW() - INTERVAL '20 days'),
(18, 'use', -0.3, 'Расход соевого молока', NOW() - INTERVAL '19 days'),
(19, 'use', -0.2, 'Расход миндального молока', NOW() - INTERVAL '19 days');