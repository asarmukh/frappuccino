# Frappuccino - Coffee Shop Management System with PostgreSQL

A modern and scalable backend system for managing coffee shop operations, extending the original Hot Coffee project by migrating from JSON-based storage to a PostgreSQL database. This RESTful API service helps coffee shops streamline their operations with improved data management, advanced reporting capabilities, and enhanced performance.

## Project Overview

Frappuccino is an upgrade to the existing Hot Coffee project, refactoring the data access layer to use PostgreSQL instead of JSON files. This transition enhances scalability, improves data integrity, and enables more complex queries and reporting features.

## Features

### Core Features (Migrated from Hot Coffee)
- **Order Management**: Create, retrieve, update, and delete customer orders
- **Menu Management**: Manage coffee shop menu items and their ingredients
- **Inventory Tracking**: Track ingredient stock levels and automatically update on order fulfillment

### New Features
- **Advanced Reporting**: Enhanced reporting capabilities using SQL aggregation
- **Full-Text Search**: Search through orders, menu items, and customers
- **Period-Based Analysis**: View orders grouped by day or month
- **Inventory Management**: Track leftovers with sorting and pagination
- **Bulk Order Processing**: Process multiple orders simultaneously with transaction support
- **Database Optimization**: Proper indexing and relation design for better performance

## Technologies Used

- Go 1.21 or higher
- PostgreSQL 15
- Docker & Docker Compose for containerization
- Third-party PostgreSQL driver

## Project Structure

```
frappuccino/
├── cmd/
│   └── main.go
├── internal/
│   ├── handler/
│   │   ├── order_handler.go
│   │   ├── menu_handler.go
│   │   ├── inventory_handler.go
│   │   └── reports_handler.go
│   ├── service/
│   │   ├── order_service.go
│   │   ├── menu_service.go
│   │   └── inventory_service.go
│   └── repository/
│       ├── order_repository.go
│       └── ...
├── models/
│   ├── order.go
│   └── ...
├── Dockerfile
├── docker-compose.yml
├── init.sql
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for development)

### Running the Application

The project is containerized with Docker for easy setup and deployment:

1. Clone the repository:
```bash
git clone <repository-url>
cd frappuccino
```

2. Start the application using Docker Compose:
```bash
docker compose up
```

This will:
- Set up a PostgreSQL database container
- Initialize the database schema using `init.sql`
- Build and run the application container

The API will be available at `http://localhost:8080`

## API Endpoints

### Original Endpoints (Refactored for PostgreSQL)

#### Orders
- `POST /orders` - Create a new order
- `GET /orders` - Retrieve all orders
- `GET /orders/{id}` - Retrieve a specific order
- `PUT /orders/{id}` - Update an existing order
- `DELETE /orders/{id}` - Delete an order
- `POST /orders/{id}/close` - Close an order

#### Menu Items
- `POST /menu` - Add a new menu item
- `GET /menu` - Retrieve all menu items
- `GET /menu/{id}` - Retrieve a specific menu item
- `PUT /menu/{id}` - Update a menu item
- `DELETE /menu/{id}` - Delete a menu item

#### Inventory
- `POST /inventory` - Add a new inventory item
- `GET /inventory` - Retrieve all inventory items
- `GET /inventory/{id}` - Retrieve a specific inventory item
- `PUT /inventory/{id}` - Update an inventory item
- `DELETE /inventory/{id}` - Delete an inventory item

#### Basic Reports
- `GET /reports/total-sales` - Get total sales amount
- `GET /reports/popular-items` - Get list of popular menu items

### New Endpoints

#### Advanced Reporting and Search
- `GET /orders/numberOfOrderedItems` - Get ordered items count within a date range
- `GET /reports/search` - Full-text search across orders, menu items, and customers
- `GET /reports/orderedItemsByPeriod` - Get order counts grouped by day or month
- `GET /inventory/getLeftOvers` - Get inventory leftovers with sorting and pagination
- `POST /orders/batch-process` - Process multiple orders simultaneously

## Database Design

The project includes a comprehensive database schema that utilizes PostgreSQL's features:

### Data Types
- **JSONB**: Used for menu item customization options, order special instructions
- **Arrays**: Used for item categories/tags, allergen information
- **ENUM**: Used for order status values, payment methods, item sizes
- **Timestamp with time zone**: Used for order dates, inventory updates

### Core Tables
- **orders**: Main order information
- **order_items**: Individual items in each order
- **menu_items**: Available products for sale
- **menu_item_ingredients**: Junction table between menu items and ingredients
- **inventory**: Available ingredients and stock levels
- **order_status_history**: Order state changes
- **price_history**: Menu item price changes
- **inventory_transactions**: Inventory changes

## Development

### Database Connection Settings
```
Host: db
Port: 5432
User: latte
Password: latte
Database: frappuccino
```

### Database Schema
The database schema is defined in `init.sql`, which is automatically executed when the containers start.

## Error Handling

The API returns appropriate HTTP status codes and error messages:
- 200: Successful GET request
- 201: Successful resource creation
- 400: Bad request/Invalid input
- 404: Resource not found
- 500: Internal server error

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## INFO.txt - data for developers
