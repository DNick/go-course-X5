package memory

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = r.nextID
	r.nextID++
	r.items = append(r.items, value)
	return cloneOrder(value), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, value := range r.items {
		if value.ID == id {
			return cloneOrder(value), nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	items := make([]order.Order, len(r.items))
	for i, value := range r.items {
		items[i] = cloneOrder(value)
	}
	return items, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i, v := range r.items {
		if v.ID == value.ID {
			r.items[i] = cloneOrder(value)
			return r.items[i], nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func cloneOrder(value order.Order) order.Order {
	value.Items = append([]order.Item(nil), value.Items...)
	value.History = append([]order.StatusChange(nil), value.History...)
	return value
}
