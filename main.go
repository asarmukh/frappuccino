package main

import (
	"context"
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
	"os/signal"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	port := flag.Int("port", 8080, "Listening port number")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		helper.PrintUsage()
		return
	}

	db := connectDB()
	defer db.Close()

	inventoryRepo := dal.NewInventoryPostgresRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	menuRepo := dal.NewMenuRepository(db)
	menuService := service.NewMenuService(menuRepo, inventoryService)
	menuHandler := handler.NewMenuHandler(menuService)

	orderRepo := dal.NewOrderPostgresRepository(db)
	orderService := service.NewOrderService(orderRepo, menuRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	setupRoutes(orderHandler, menuHandler, inventoryHandler)

	if *port < 1 || *port > 65535 {
		log.Fatal("Error port")
	}

	// Настроим сервер
	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      nil, // Использует уже настроенные маршруты
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Канал для обработки сигналов завершения работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Запуск сервера в горутине
	go func() {
		log.Printf("🚀 The server is running on the port: %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	// Ожидаем сигнала для остановки
	<-stop
	log.Println("Получен сигнал остановки, завершаем работу...")

	// Корректное завершение работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Ошибка при завершении работы сервера:", err)
	}

	log.Println("Сервер успешно завершил работу")
}

// Подключение к базе данных с тайм-аутом
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
		log.Fatal("❌ Database connection error:", err)
	}

	// Ожидание доступности базы данных с тайм-аутом
	waitForDB(db)
	fmt.Println("✅ Connected to PostgreSQL")
	return db
}

// Ожидание доступности базы данных
func waitForDB(db *sql.DB) {
	timeout := time.After(30 * time.Second) // Тайм-аут через 30 секунд
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			log.Fatal("❌ Тайм-аут подключения к БД")
		case <-tick:
			if err := db.Ping(); err == nil {
				return
			}
			fmt.Println("⏳ Ожидание подключения к БД...")
		}
	}
}

func setupRoutes(orderHandler handler.OrderHandler, menuHandler handler.MenuHandler, inventoryHandler handler.InventoryHandler) {
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔥 Request processed in /orders")
		routes.HandleRequestsOrders(orderHandler)(w, r)
	})

	http.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔥 Request processed in /menu")
		routes.HandleMenu(menuHandler)(w, r)
	})

	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔥 Request processed in /inventory")
		routes.HandleRequestsInventory(inventoryHandler)(w, r)
	})

	http.HandleFunc("/inventory/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔥 Request processed in /inventory")
		routes.HandleRequestsInventory(inventoryHandler)(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("🔥 Request for an unknown route:", r.URL.Path)
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
