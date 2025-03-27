package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/internal/service"
	"frappuccino/models"
	"frappuccino/utils"
	"log/slog"
	"net/http"
)

type OrderHandlerInterface interface {
	HandleCreateOrder(w http.ResponseWriter, r *http.Request)
	HandleGetAllOrders(w http.ResponseWriter, r *http.Request)
	HandleGetOrderById(w http.ResponseWriter, r *http.Request, orderID string)
	HandleDeleteOrder(w http.ResponseWriter, r *http.Request, orderID string)
	HandleUpdateOrder(w http.ResponseWriter, r *http.Request, orderID string)
	HandleCloseOrder(w http.ResponseWriter, r *http.Request, orderID string)
}

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) OrderHandler {
	return OrderHandler{orderService: orderService}
}

func (h OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to create order")

	var newOrder models.Order
	if err := json.NewDecoder(r.Body).Decode(&newOrder); err != nil {
		slog.Warn("Invalid JSON format", "error", err)
		utils.ErrorInJSON(w, http.StatusBadRequest, fmt.Errorf("invalid JSON format: %v", err))
		return
	}
	// add more validations for body json
	if newOrder.Status != "" {
		utils.ErrorInJSON(w, 400, errors.New("invalid request body"))
	}

	order, err := h.orderService.CreateOrder(newOrder)
	if err != nil {
		slog.Error("Failed to create order", "error", err)
		utils.ErrorInJSON(w, 400, err)
		return
	} else {
		slog.Info("✔ Order created successfully", "orderID", order.ID)
		utils.ResponseInJSON(w, 201, order)
	}
}

func (h OrderHandler) HandleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request to get all orders")

	orders, err := h.orderService.GetAllOrders()
	if err != nil {
		slog.Error("Failed to retrieve orders", "error", err)
		utils.ErrorInJSON(w, http.StatusNotFound, err)
	}

	slog.Info("Successfully retrieved all orders", "count", len(orders))
	utils.ResponseInJSON(w, 200, orders)
}

func (h OrderHandler) HandleGetOrderById(w http.ResponseWriter, r *http.Request, orderID int) {
	slog.Info("Received request to get order", "orderID", orderID)

	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		slog.Warn("Order not found", "orderID", orderID, "error", err)
		utils.ErrorInJSON(w, http.StatusNotFound, err)
		return
	}

	slog.Info("Successfully retrieved order", "orderID", order.ID)
	utils.ResponseInJSON(w, 200, order)
}

// func (h OrderHandler) HandleDeleteOrder(w http.ResponseWriter, r *http.Request, orderID string) {
// 	slog.Info("Received request to delete order", "orderID", orderID)

// 	err := h.orderService.DeleteOrder(orderID)
// 	if err != nil {
// 		slog.Warn("Failed to delete order", "orderID", orderID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	}

// 	slog.Info("Order deleted successfully", "orderID", orderID)
// 	w.WriteHeader(http.StatusNoContent)
// }

// func (h OrderHandler) HandleUpdateOrder(w http.ResponseWriter, r *http.Request, orderID string) {
// 	slog.Info("Received request to update order", "orderID", orderID)

// 	var changeOrder models.Order
// 	if err := json.NewDecoder(r.Body).Decode(&changeOrder); err != nil {
// 		slog.Warn("Invalid JSON format", "error", err)
// 		utils.ErrorInJSON(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	order, err := h.orderService.UpdateOrder(orderID, changeOrder)
// 	if err != nil {
// 		slog.Warn("Failed to update order", "orderID", orderID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	} else {

// 		slog.Info("Order updated successfully", "orderID", order.ID)
// 		utils.ResponseInJSON(w, 200, order)
// 	}
// }

// func (h OrderHandler) HandleCloseOrder(w http.ResponseWriter, r *http.Request, orderID string) {
// 	slog.Info("Received request to close order", "orderID", orderID)

// 	order, err := h.orderService.CloseOrder(orderID)
// 	if err != nil {
// 		slog.Warn("Failed to close order", "orderID", orderID, "error", err)
// 		utils.ErrorInJSON(w, http.StatusNotFound, err)
// 		return
// 	} else {
// 		slog.Info("Order closed successfully", "orderID", order.ID)
// 		utils.ResponseInJSON(w, 200, order)
// 	}
// }
