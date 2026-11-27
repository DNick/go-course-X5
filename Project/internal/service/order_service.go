package service

import (
	"context"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}

	for _, item := range input.Items {
		if item.Quantity <= 0 || item.Price <= 0 || strings.TrimSpace(item.Name) == "" {
			return order.Order{}, order.ErrInvalidOrder
		}
	}

	now := time.Now()
	value := order.Order{
		Items:      input.Items,
		TotalPrice: calculateTotalPrice(input.Items),
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{
				Status:    order.StatusNew,
				ChangedAt: now,
			},
		},
	}
	return s.repository.Create(ctx, value)
}

func calculateTotalPrice(items []order.Item) int64 {
	var total int64
	for _, item := range items {
		total += int64(item.Quantity) * item.Price
	}
	return total
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}
	if status == order.StatusNew || value.Status == order.StatusCompleted || value.Status == order.StatusCancelled {
		return order.Order{}, order.ErrInvalidStatus
	}

	allowed := (value.Status == order.StatusNew &&
		(status == order.StatusProcessing || status == order.StatusCancelled)) ||
		(value.Status == order.StatusProcessing &&
			(status == order.StatusCompleted || status == order.StatusCancelled))
	if !allowed {
		return order.Order{}, order.ErrInvalidStatusTransition
	}

	value.Status = status
	value.History = append(value.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now(),
	})

	return s.repository.Update(ctx, value)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return value.History, nil
}
