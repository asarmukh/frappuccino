package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"frappuccino/helper"
	"frappuccino/internal/dal"
	"frappuccino/internal/handler"
	"frappuccino/internal/routes"
	"frappuccino/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	port := flag.Int("port", 8080, "Port number to listen on")
	help := flag.Bool("help", false, "Show help")
	dir := flag.String("dir", "data", "Directory path for storing data")
	flag.Parse()

	if *help {
		helper.PrintUsage()
		return
	}

	helper.CreateNewDir(*dir)

	db := connectDB()
	defer db.Close()

	inventoryRepo := dal.NewInventoryRepositoryJSON(*dir)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	menuRepo := dal.NewMenuRepositoryJSON(*dir)
	menuService := service.NewMenuService(menuRepo, inventoryService)
	menuHandler := handler.NewMenuHandler(menuService)

	orderRepo := dal.NewOrderPostgresRepository(db)
	orderService := service.NewOrderService(orderRepo, menuService, inventoryService)
	orderHandler := handler.NewOrderHandler(orderService)

	setupRoutes(orderHandler, menuHandler, inventoryHandler)

	if *port < 1 || *port > 65535 {
		log.Fatal("Error port")
	}

	// Start Server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("🚀 Server start on port: %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Error start server:", err)
	}
}

// Подключение к базе данных с ожиданием её готовности
func connectDB() *sql.DB {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "latte"),
		getEnv("DB_PASSWORD", "latte"),
		getEnv("DB_NAME", "frappuccino"),
		getEnv("DB_PORT", "5432"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к БД:", err)
	}

	waitForDB(db)
	fmt.Println("✅ Подключено к PostgreSQL")
	return db
}

// Ожидание доступности базы данных
func waitForDB(db *sql.DB) {
	for {
		if err := db.Ping(); err == nil {
			return
		}
		fmt.Println("⏳ Ожидание подключения к БД...")
		time.Sleep(2 * time.Second)
	}
}

// Функция для установки маршрутов
func setupRoutes(orderHandler handler.OrderHandler, menuHandler handler.MenuHandler, inventoryHandler handler.InventoryHandler) {
	http.HandleFunc("/orders", routes.HandleRequestsOrders(orderHandler))
	http.HandleFunc("/orders/", routes.HandleRequestsOrders(orderHandler))

	http.HandleFunc("/menu", routes.HandleMenu(menuHandler))
	http.HandleFunc("/menu/", routes.HandleMenu(menuHandler))

	http.HandleFunc("/inventory", routes.HandleRequestsInventory(inventoryHandler))
	http.HandleFunc("/inventory/", routes.HandleRequestsInventory(inventoryHandler))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
}

// Функция для получения переменных окружения с дефолтным значением
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
