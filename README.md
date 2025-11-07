# Frappuccino ☕

A PostgreSQL-based coffee shop management system built with Go, featuring order management, inventory tracking, and advanced reporting capabilities.

## 📋 Overview

Frappuccino is a refactored version of the hot-coffee project, migrated from JSON-based storage to a robust PostgreSQL database. This project demonstrates advanced SQL operations, database design principles, and RESTful API development.

## 🎯 Features

### Core Functionality
- **Order Management**: Create, read, update, delete, and close orders
- **Menu Management**: Full CRUD operations for menu items
- **Inventory Control**: Track ingredients and stock levels
- **Price History**: Monitor menu item price changes over time
- **Order Status Tracking**: Follow order lifecycle through status transitions

### Advanced Features
- **Full-Text Search**: Search across orders, menu items, and customers with relevance ranking
- **Sales Reports**: Total sales and popular items analytics
- **Time-Based Analytics**: Track orders by day or month
- **Batch Order Processing**: Handle multiple concurrent orders with transaction support
- **Inventory Pagination**: Sorted and paginated inventory views

## 🏗️ Database Architecture

### Core Tables
- `orders` - Main order information with customer details
- `order_items` - Individual items within orders
- `menu_items` - Available products for sale
- `menu_item_ingredients` - Recipe definitions
- `inventory` - Ingredient stock management
- `order_status_history` - Order state change tracking
- `price_history` - Menu item price changes
- `inventory_transactions` - Stock movement records

### Advanced PostgreSQL Features
- **JSONB**: Menu customizations, order instructions, customer preferences
- **Arrays**: Categories, allergens, tags
- **ENUMs**: Order status, payment methods, item sizes
- **Timestamps with timezone**: Accurate time tracking across operations
- **Full-text search indexes**: Fast search capabilities
- **Composite indexes**: Optimized query performance

## 🚀 Getting Started

### Prerequisites
- Docker
- Docker Compose

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd frappuccino
```

2. Ensure the following files are in the root directory:
   - `Dockerfile`
   - `docker-compose.yml`
   - `init.sql`

3. Start the application:
```bash
docker compose up
```

The API will be available at `http://localhost:8080`

### Database Connection Settings
- **Host**: db
- **Port**: 5432
- **User**: latte
- **Password**: latte
- **Database**: frappuccino

## 📡 API Endpoints

### Orders
- `POST /orders` - Create a new order
- `GET /orders` - Retrieve all orders
- `GET /orders/{id}` - Get specific order
- `PUT /orders/{id}` - Update order
- `DELETE /orders/{id}` - Delete order
- `POST /orders/{id}/close` - Close order
- `GET /orders/numberOfOrderedItems` - Get ordered items count by date range
- `POST /orders/batch-process` - Process multiple orders simultaneously

### Menu Items
- `POST /menu` - Add new menu item
- `GET /menu` - Retrieve all menu items
- `GET /menu/{id}` - Get specific menu item
- `PUT /menu/{id}` - Update menu item
- `DELETE /menu/{id}` - Delete menu item

### Inventory
- `POST /inventory` - Add inventory item
- `GET /inventory` - Retrieve all inventory
- `GET /inventory/{id}` - Get specific inventory item
- `PUT /inventory/{id}` - Update inventory
- `DELETE /inventory/{id}` - Delete inventory item
- `GET /inventory/getLeftOvers` - Get paginated inventory with sorting

### Reports & Analytics
- `GET /reports/total-sales` - Total sales amount
- `GET /reports/popular-items` - Most popular menu items
- `GET /reports/search` - Full-text search across entities
- `GET /reports/orderedItemsByPeriod` - Orders grouped by time period

## 📊 Example API Calls

### Search Menu and Orders
```bash
GET /reports/search?q=chocolate%20cake&filter=menu,orders&minPrice=10
```

### Get Ordered Items Count
```bash
GET /orders/numberOfOrderedItems?startDate=2024-11-01&endDate=2024-11-30
```

### Get Monthly Order Statistics
```bash
GET /reports/orderedItemsByPeriod?period=month&year=2024
```

### Get Inventory with Pagination
```bash
GET /inventory/getLeftOvers?sortBy=quantity&page=1&pageSize=10
```

### Batch Order Processing
```bash
POST /orders/batch-process
Content-Type: application/json

{
  "orders": [
    {
      "customer_name": "Alice",
      "items": [
        {"menu_item_id": 1, "quantity": 2}
      ]
    }
  ]
}
```

## 🛠️ Technical Requirements

- **Language**: Go (with gofumpt formatting)
- **Database**: PostgreSQL
- **Driver**: PostgreSQL driver (only external package allowed)
- **Architecture**: Layered architecture with Data Access Layer (DAL)
- **Containerization**: Docker and Docker Compose
---

**Note**: For detailed endpoint specifications and response formats, refer to the project documentation above.
