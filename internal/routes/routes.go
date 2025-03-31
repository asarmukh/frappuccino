package routes

import (
	"frappuccino/internal/handler"
	"net/http"
	"strconv"
	"strings"
)

// func HandleRequestsReports(reportHandler handler.ReportHandler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		path := strings.Trim(r.URL.Path, "/")
// 		parts := strings.SplitN(path, "/", 2)

// 		switch r.Method {
// 		case http.MethodGet:
// 			if len(parts) == 2 && parts[1] == "total-sales" {
// 				reportHandler.HandleGetTotalSales(w, r)
// 			} else if len(parts) == 2 && parts[1] == "popular-items" {
// 				reportHandler.HandleGetPopulatItem(w, r)
// 			} else {
// 				http.Error(w, "Not Found", http.StatusNotFound)
// 			}
// 		default:
// 			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 		}
// 	}
// }

func HandleRequestsInventory(inventoryHandler handler.InventoryHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.SplitN(path, "/", 3)

		var id int
		if len(parts) > 1 {
			var err error
			id, err = strconv.Atoi(parts[1])
			if err != nil {
				http.Error(w, "Invalid order ID", http.StatusBadRequest)
				return
			}
		}

		switch r.Method {
		case http.MethodPost:
			inventoryHandler.HandleCreateInventory(w, r)
		case http.MethodGet:
			if len(parts) == 1 {
				inventoryHandler.HandleGetAllInventory(w, r)
			} else if len(parts) == 2 {
				inventoryHandler.HandleGetInventoryById(w, r, id)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		case http.MethodDelete:
			if len(parts) == 2 {
				inventoryHandler.HandleDeleteInventoryItem(w, r, id)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		case http.MethodPut:
			if len(parts) == 2 {
				inventoryHandler.HandleUpdateInventoryItem(w, r, id)
			} else {
				http.Error(w, "Bad Request", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}

func HandleRequestsOrders(orderHandler handler.OrderHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		var id int
		if len(parts) > 1 {
			var err error
			id, err = strconv.Atoi(parts[1])
			if err != nil {
				http.Error(w, "Invalid order ID", http.StatusBadRequest)
				return
			}
		}

		switch r.Method {
		case http.MethodPost:
			if len(parts) == 1 {
				orderHandler.HandleCreateOrder(w, r)
			} else if len(parts) == 3 && parts[2] == "close" {
				orderHandler.HandleCloseOrder(w, r, id)
			} else {
				http.Error(w, "Bad Request", http.StatusBadRequest)
			}

		case http.MethodGet:
			if len(parts) == 1 {
				orderHandler.HandleGetAllOrders(w, r)
			} else if len(parts) == 2 {
				orderHandler.HandleGetOrderById(w, r, id)
			} else if len(parts) == 2 && parts[0] == "orders" && parts[1] == "numberOfOrderedItems" {
				startDate := r.URL.Query().Get("startDate")
				endDate := r.URL.Query().Get("endDate")
				orderHandler.HandleNumberOfOrderedItems(w, r, startDate, endDate)
				return
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		case http.MethodDelete:
			if len(parts) == 2 {
				orderHandler.HandleDeleteOrder(w, r, id)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		case http.MethodPut:
			if len(parts) == 2 {
				orderHandler.HandleUpdateOrder(w, r, id)
			} else {
				http.Error(w, "Bad Request", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}

func HandleMenu(menuHandler handler.MenuHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.SplitN(path, "/", 2)

		var id int
		if len(parts) > 1 {
			var err error
			id, err = strconv.Atoi(parts[1])
			if err != nil {
				http.Error(w, "Invalid order ID", http.StatusBadRequest)
				return
			}
		}

		switch r.Method {
		case http.MethodPost:
			if len(parts) == 1 {
				menuHandler.HandleCreateMenuItem(w, r)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}

		case http.MethodGet:
			if len(parts) == 1 {
				menuHandler.HandleGetAllMenuItems(w, r)
			} else if len(parts) == 2 {
				menuHandler.HandleGetMenuItemById(w, r, id)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}

		case http.MethodPut:
			if len(parts) == 2 {
				menuHandler.HandleUpdateMenu(w, r, id)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
		case http.MethodDelete:
			if len(parts) == 2 {
				menuHandler.HandleDeleteMenuItemById(w, r, id)
			} else {
				http.Error(w, "Not Found", http.StatusNotFound)
			}

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
