package service_test

import (
	"context"
	"testing"

	"example.com/order-service/internal/order"
	memoryrepository "example.com/order-service/internal/repository/memory"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name      string
		input     order.CreateOrderInput
		wantError error
		wantTotal int64
	}{
		{
			name: "success",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "Keyboard", Quantity: 2, Price: 500},
				{Name: "Mouse", Quantity: 1, Price: 300},
			}},
			wantTotal: 1300,
		},
		{name: "no items", input: order.CreateOrderInput{}, wantError: order.ErrInvalidOrder},
		{
			name: "empty name",
			input: order.CreateOrderInput{Items: []order.Item{
				{Name: "   ", Quantity: 1, Price: 100},
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
		t.Run(tt.name, func(t *testing.T) {
			repository := memoryrepository.NewOrderRepository()
			orderService := service.NewOrderService(repository)
			got, err := orderService.Create(context.Background(), tt.input)
			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, int64(1), got.ID)
			assert.Equal(t, order.StatusNew, got.Status)
			assert.Equal(t, tt.wantTotal, got.TotalPrice)
		})
	}
}

func TestGetByID(t *testing.T) {
	repository := memoryrepository.NewOrderRepository()
	orderService := service.NewOrderService(repository)

	created, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500}},
	})
	require.NoError(t, err)

	got, err := orderService.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	_, err = orderService.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestList(t *testing.T) {
	repository := memoryrepository.NewOrderRepository()
	orderService := service.NewOrderService(repository)

	first, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500}},
	})
	require.NoError(t, err)
	second, err := orderService.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{{Name: "Mouse", Quantity: 2, Price: 300}},
	})
	require.NoError(t, err)

	got, err := orderService.List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, []int64{first.ID, second.ID}, []int64{got[0].ID, got[1].ID})
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name      string
		from      order.Status
		status    order.Status
		wantError error
	}{
		{name: "new to processing", from: order.StatusNew, status: order.StatusProcessing},
		{name: "new to cancelled", from: order.StatusNew, status: order.StatusCancelled},
		{name: "processing to completed", from: order.StatusProcessing, status: order.StatusCompleted},
		{name: "processing to cancelled", from: order.StatusProcessing, status: order.StatusCancelled},
		{name: "new status", from: order.StatusNew, status: order.StatusNew, wantError: order.ErrInvalidStatus},
		{name: "completed terminal", from: order.StatusCompleted, status: order.StatusProcessing, wantError: order.ErrInvalidStatus},
		{name: "cancelled terminal", from: order.StatusCancelled, status: order.StatusProcessing, wantError: order.ErrInvalidStatus},
		{name: "new cannot complete", from: order.StatusNew, status: order.StatusCompleted, wantError: order.ErrInvalidStatusTransition},
		{name: "unknown status", from: order.StatusNew, status: order.Status("unknown"), wantError: order.ErrInvalidStatusTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := memoryrepository.NewOrderRepository()
			created, err := repository.Create(context.Background(), order.Order{
				Status: tt.from,
				History: []order.StatusChange{
					{Status: tt.from},
				},
			})
			require.NoError(t, err)

			orderService := service.NewOrderService(repository)
			got, err := orderService.UpdateStatus(context.Background(), created.ID, tt.status)
			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.status, got.Status)
			assert.Len(t, got.History, 2)
		})
	}
}

func TestHistory(t *testing.T) {
	repository := memoryrepository.NewOrderRepository()
	orderService := service.NewOrderService(repository)

	created, err := repository.Create(context.Background(), order.Order{
		Status: order.StatusNew,
		History: []order.StatusChange{
			{Status: order.StatusNew},
			{Status: order.StatusProcessing},
		},
	})
	require.NoError(t, err)

	history, err := orderService.History(context.Background(), created.ID)
	require.NoError(t, err)
	require.Len(t, history, 2)
	assert.Equal(t, []order.Status{order.StatusNew, order.StatusProcessing}, []order.Status{
		history[0].Status,
		history[1].Status,
	})

	_, err = orderService.History(context.Background(), 999)
	assert.ErrorIs(t, err, order.ErrOrderNotFound)
}
