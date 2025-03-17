package main

import (
	"database/sql"
	"flag"
	"fmt"
	"frappuccino/helper"
	"frappuccino/internal/dal"
	"frappuccino/internal/handler"
	"frappuccino/internal/routes"
	"frappuccino/internal/service"
	"log"
	"net/http"
	"os"

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

	// Order service and handler
	orderRepo := dal.NewOrderPostgresRepository(db)
	orderService := service.NewOrderService(orderRepo, menuService, inventoryService) //всё переделывать
	orderHandler := handler.NewOrderHandler(orderService)

	// Report service and handler
	// reportService := service.NewReportService(menuService, orderService)
	// reportHandler := handler.NewReportHandler(reportService)

	// HTTP Routes setup
	http.HandleFunc("/orders", routes.HandleRequestsOrders(orderHandler))
	http.HandleFunc("/orders/", routes.HandleRequestsOrders(orderHandler))

	http.HandleFunc("/menu", routes.HandleMenu(menuHandler))
	http.HandleFunc("/menu/", routes.HandleMenu(menuHandler))

	http.HandleFunc("/inventory", routes.HandleRequestsInventory(inventoryHandler))
	http.HandleFunc("/inventory/", routes.HandleRequestsInventory(inventoryHandler))

	// http.HandleFunc("/reports/", routes.HandleRequestsReports(reportHandler))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found.", http.StatusNotFound)
	})

	if *port < 0 || *port > 65535 {
		log.Fatal("Invalid port number")
	}

	addr := fmt.Sprintf(":%d", *port)
	// // Запуск браузера
	// go helper.OpenBrowser(addr)

	log.Printf("Server running on port %s with BaseDir %s\n", addr, *dir)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func connectDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "latte"),
		getEnv("DB_PASSWORD", "latte"),
		getEnv("DB_NAME", "frappuccino"),
		getEnv("DB_PORT", "5432"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("БД недоступна:", err)
	}
	fmt.Println("✅ Подключено к PostgreSQL")
	return db
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
