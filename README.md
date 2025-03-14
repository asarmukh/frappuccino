# Hot Coffee - Coffee Shop Management System

A robust backend system for managing coffee shop operations, including order processing, menu management, and inventory tracking. This RESTful API service helps coffee shops streamline their operations by handling orders, tracking inventory, and managing menu items efficiently.

## Features

- **Order Management**: Create, retrieve, update, and delete customer orders
- **Menu Management**: Manage coffee shop menu items and their ingredients
- **Inventory Tracking**: Track ingredient stock levels and automatically update on order fulfillment
- **Sales Reporting**: Generate reports on total sales and popular items
- **JSON-based Storage**: Simple and portable data storage using JSON files
- **Layered Architecture**: Clean separation of concerns for better maintainability

## Prerequisites

- Go 1.21 or higher
- No external dependencies required (uses only standard library)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd hot-coffee
```

2. Build the application:
```bash
go build -o hot-coffee .
```

## Usage

Start the server with optional configuration flags:

```bash
./hot-coffee --port 8080 --dir ./data
```

### Command Line Options

- `--port N`: Specify the port number for the HTTP server (default: 8080)
- `--dir S`: Specify the path to the data directory (default: ./data)
- `--help`: Display usage information

## API Endpoints

### Orders

- `POST /orders` - Create a new order
- `GET /orders` - Retrieve all orders
- `GET /orders/{id}` - Retrieve a specific order
- `PUT /orders/{id}` - Update an existing order
- `DELETE /orders/{id}` - Delete an order
- `POST /orders/{id}/close` - Close an order

### Menu Items

- `POST /menu` - Add a new menu item
- `GET /menu` - Retrieve all menu items
- `GET /menu/{id}` - Retrieve a specific menu item
- `PUT /menu/{id}` - Update a menu item
- `DELETE /menu/{id}` - Delete a menu item

### Inventory

- `POST /inventory` - Add a new inventory item
- `GET /inventory` - Retrieve all inventory items
- `GET /inventory/{id}` - Retrieve a specific inventory item
- `PUT /inventory/{id}` - Update an inventory item
- `DELETE /inventory/{id}` - Delete an inventory item

### Reports

- `GET /reports/total-sales` - Get total sales amount
- `GET /reports/popular-items` - Get list of popular menu items

## Data Storage

The application uses JSON files for data storage, located in the specified data directory:

- `orders.json` - Stores order information
- `menu_items.json` - Stores menu items and their ingredients
- `inventory.json` - Stores inventory levels

## Project Structure

```
hot-coffee/
├── cmd/
│   └── main.go
├── internal/
│   ├── handler/
│   │   ├── order_handler.go
│   │   ├── menu_handler.go
│   │   └── inventory_handler.go
│   ├── service/
│   │   ├── order_service.go
│   │   ├── menu_service.go
│   │   └── inventory_service.go
│   └── dal/
│       ├── order_repository.go
│       └── ...
├── models/
│   ├── order.go
│   └── ...
├── go.mod
└── go.sum
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- 200: Successful GET request
- 201: Successful resource creation
- 400: Bad request/Invalid input
- 404: Resource not found
- 500: Internal server error

## Logging

The application uses Go's `log/slog` package for logging operations and errors. Logs include timestamps and contextual information for better debugging and monitoring.
