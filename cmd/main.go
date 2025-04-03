package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"frappuccino/helper"
	"frappuccino/internal/config"
	"frappuccino/internal/dal"
	"frappuccino/internal/handler"
	"frappuccino/internal/service"

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

	db := config.ConnectDB()
	defer db.Close()

	inventoryRepo := dal.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	menuRepo := dal.NewMenuRepository(db)
	menuService := service.NewMenuService(menuRepo)
	menuHandler := handler.NewMenuHandler(menuService)

	orderRepo := dal.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, menuRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	reportRepo := dal.NewReportRepository(db)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportService)

	mux := http.NewServeMux()
	config.SetupRoutes(mux, orderHandler, menuHandler, inventoryHandler, reportHandler)

	if *port < 1 || *port > 65535 {
		log.Fatal("Error port")
	}

	// Настроим сервер
	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux, // Использует уже настроенные маршруты
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
