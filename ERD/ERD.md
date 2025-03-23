```mermaid
erDiagram
    orders ||--|{ order_items : contains
    orders ||--|{ order_status_history : tracks
    menu_items ||--o{ order_items : "ordered as"
    menu_items ||--|{ menu_item_ingredients : requires
    inventory ||--|{ inventory_transactions : logs
    menu_items ||--o{ price_history : records

    orders {
        int id PK
        varchar name
        enum order_status
        decimal total_amount
        jsonb special_instructions
        boolean is_completed
        timestamp_tz created_at
        timestamp_tz updated_at
    }

    order_items {
        int id PK
        int order_id FK
        int menu_item_id FK
        int quantity
        decimal unit_price
        jsonb customizations
        text notes
    }

    order_status_history {
        int id PK
        int order_id FK
        enum status
        int staff_id FK
        text notes
        timestamp_tz created_at
    }

    menu_items {
        int id PK
        varchar name
        text description
        decimal base_price
        varchar[] categories
        varchar[] allergens
        enum size
        boolean available
        jsonb customization_options
        timestamp_tz created_at
        timestamp_tz updated_at
    }

    price_history {
        int id PK
        int menu_item_id FK
        decimal price
        timestamp_tz effective_from
        timestamp_tz effective_to
        text change_reason
    }

    menu_item_ingredients {
        int menu_item_id PK, FK
        decimal quantity
        varchar unit
        boolean is_optional
        varchar[] substitutes
    }

    inventory {
        int id PK
        varchar name
        decimal current_quantity
        decimal cost_per_unit
        timestamp_tz last_restocked
        timestamp_tz expiry_date
    }

    inventory_transactions {
        int id PK
        int inventory_id FK
        enum transaction_type
        decimal quantity
        text notes
        timestamp_tz created_at
    }
