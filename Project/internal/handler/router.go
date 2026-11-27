package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.getOrdersList)
		r.Get("/{id}", h.getOrderByID)
		r.Patch("/{id}/status", h.changeOrderStatus)
		r.Get("/{id}/history", h.getOrderHistory)
	})

	return router
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST_BODY",
		})
		return
	}

	value, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: err.Error(),
			Code:  "CREATE_ORDER_FAILED",
		})
		return
	}

	writeJSON(w, http.StatusCreated, value)
}

func (h *Handler) getOrdersList(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: err.Error(),
			Code:  "LIST_ORDERS_FAILED",
		})
		return
	}

	writeJSON(w, http.StatusOK, value)
}

func (h *Handler) getOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order ID",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	value, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: err.Error(),
			Code:  "ORDER_NOT_FOUND",
		})
		return
	}

	writeJSON(w, http.StatusOK, value)
}

func (h *Handler) changeOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order ID",
			Code:  "400",
		})
		return
	}
	var input order.UpdateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST_BODY",
		})
		return
	}

	value, err := h.service.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrOrderNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{
				Error: err.Error(),
				Code:  "ORDER_NOT_FOUND",
			})
		case errors.Is(err, order.ErrInvalidStatusTransition):
			writeJSON(w, http.StatusConflict, errorResponse{
				Error: err.Error(),
				Code:  "INVALID_STATUS_TRANSITION",
			})
		case errors.Is(err, order.ErrInvalidStatus):
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: err.Error(),
				Code:  "INVALID_STATUS",
			})
		default:
			h.logger.Error("failed to update order status",
				slog.Int64("order_id", id),
				slog.Any("error", err),
			)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
				Code:  "INTERNAL_SERVER_ERROR",
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, value)
}

func (h *Handler) getOrderHistory(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid order ID",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	value, err := h.service.History(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrOrderNotFound):
			writeJSON(w, http.StatusNotFound, errorResponse{
				Error: err.Error(),
				Code:  "ORDER_NOT_FOUND",
			})
		default:
			h.logger.Error("failed to get order history",
				slog.Int64("order_id", id),
				slog.Any("error", err),
			)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
				Code:  "INTERNAL_SERVER_ERROR",
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, value)

}

func (h *Handler) notImplemented(w http.ResponseWriter, _ *http.Request) {
	// TODO: replace this placeholder with separate handlers for every endpoint.
	writeJSON(w, http.StatusNotImplemented, errorResponse{
		Error: "endpoint is not implemented",
		Code:  "NOT_IMPLEMENTED",
	})
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
