package service_test

import (
	"context"
	"errors"
	"testing"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
)

type createRepositoryStub struct {
	created order.Order
	err     error
}

func (r *createRepositoryStub) Create(_ context.Context, value order.Order) (order.Order, error) {
	r.created = value
	if r.err != nil {
		return order.Order{}, r.err
	}
	value.ID = 1
	return value, nil
}

func (r *createRepositoryStub) GetByID(context.Context, int64) (order.Order, error) {
	return order.Order{}, order.ErrNotImplemented
}

func (r *createRepositoryStub) List(context.Context) ([]order.Order, error) {
	return nil, order.ErrNotImplemented
}

func (r *createRepositoryStub) Update(context.Context, order.Order) (order.Order, error) {
	return order.Order{}, order.ErrNotImplemented
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	validItems := []order.Item{
		{Name: "Keyboard", Quantity: 2, Price: 500},
		{Name: "Mouse", Quantity: 1, Price: 300},
	}

	tests := []struct {
		name      string
		input     order.CreateOrderInput
		wantError error
		wantTotal int64
	}{
		{
			name:      "successful creation",
			input:     order.CreateOrderInput{Items: validItems},
			wantTotal: 1300,
		},
		{
			name:      "empty items",
			input:     order.CreateOrderInput{},
			wantError: order.ErrInvalidOrder,
		},
		{
			name: "empty name",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: " ", Quantity: 1, Price: 100},
			}},
			wantError: order.ErrInvalidOrder,
		},
		{
			name: "zero quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 0, Price: 100},
			}},
			wantError: order.ErrInvalidOrder,
		},
		{
			name: "negative quantity",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: -1, Price: 100},
			}},
			wantError: order.ErrInvalidOrder,
		},
		{
			name: "zero price",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 1, Price: 0},
			}},
			wantError: order.ErrInvalidOrder,
		},
		{
			name: "negative price",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 1, Price: -100},
			}},
			wantError: order.ErrInvalidOrder,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repository := &createRepositoryStub{}
			orderService := service.NewOrderService(repository)

			got, err := orderService.Create(context.Background(), tt.input)

			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("Create() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if got.ID != 1 {
				t.Errorf("Create() ID = %d, want 1", got.ID)
			}
			if got.Status != order.StatusNew {
				t.Errorf("Create() Status = %q, want %q", got.Status, order.StatusNew)
			}
			if got.TotalPrice != tt.wantTotal {
				t.Errorf("Create() TotalPrice = %d, want %d", got.TotalPrice, tt.wantTotal)
			}
			if got.CreatedAt.IsZero() {
				t.Error("Create() CreatedAt is zero")
			}
			if len(got.History) != 1 {
				t.Fatalf("Create() history length = %d, want 1", len(got.History))
			}
			if got.History[0].Status != order.StatusNew {
				t.Errorf("Create() history status = %q, want %q", got.History[0].Status, order.StatusNew)
			}
		})
	}
}

func TestCreateOrderRepositoryError(t *testing.T) {
	t.Parallel()

	repositoryErr := errors.New("repository error")
	repository := &createRepositoryStub{err: repositoryErr}
	orderService := service.NewOrderService(repository)

	_, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500}},
	})

	if !errors.Is(err, repositoryErr) {
		t.Fatalf("Create() error = %v, want %v", err, repositoryErr)
	}
}
