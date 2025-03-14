package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"frappuccino/helper"
	"frappuccino/internal/dal"
	"frappuccino/internal/handler"
	"frappuccino/internal/routes"
	"frappuccino/internal/service"
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

	inventoryRepo := dal.NewInventoryRepositoryJSON(*dir)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	menuRepo := dal.NewMenuRepositoryJSON(*dir)
	menuService := service.NewMenuService(menuRepo, inventoryService)
	menuHandler := handler.NewMenuHandler(menuService)

	// Order service and handler
	orderRepo := dal.NewOrderRepositoryJSON(*dir)
	orderService := service.NewOrderService(orderRepo, menuService, inventoryService)
	orderHandler := handler.NewOrderHandler(orderService)

	// Report service and handler
	reportService := service.NewReportService(menuService, orderService)
	reportHandler := handler.NewReportHandler(reportService)

	// HTTP Routes setup
	http.HandleFunc("/orders", routes.HandleRequestsOrders(orderHandler))
	http.HandleFunc("/orders/", routes.HandleRequestsOrders(orderHandler))

	http.HandleFunc("/menu", routes.HandleMenu(menuHandler))
	http.HandleFunc("/menu/", routes.HandleMenu(menuHandler))

	http.HandleFunc("/inventory", routes.HandleRequestsInventory(inventoryHandler))
	http.HandleFunc("/inventory/", routes.HandleRequestsInventory(inventoryHandler))

	http.HandleFunc("/reports/", routes.HandleRequestsReports(reportHandler))

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
