package main

import "testing"

func TestEmptyOrder(t *testing.T) {
	order, err := NewOrder(1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if Total(order) != 0 {
		t.Fatalf("expected sum 0, got %d", Total(order))
	}
}

func TestAddItem(t *testing.T) {
	order, err := NewOrder(1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = AddItem(&order, Item{Name: "item1", PriceKopecks: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(order.Items))
	}
}

func TestTotal(t *testing.T) {
	order, err := NewOrder(1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	AddItem(&order, Item{Name: "item1", PriceKopecks: 100})
	AddItem(&order, Item{Name: "item2", PriceKopecks: 50})

	if Total(order) != 150 {
		t.Fatalf("expected sum 150, got %d", Total(order))
	}
}

func TestOrderCopyDoesNotChangeOriginal(t *testing.T) {
	original := Order{
		ID:     1,
		Items:  []Item{{Name: "item1", PriceKopecks: 100}},
		Status: "new",
	}
	copyOrder := original
	copyOrder.Items = append(copyOrder.Items, original.Items...)

	copyOrder.Items[0].Name = "updated item"

	if original.Items[0].Name != "item1" {
		t.Fatalf("changing the copy changed the original order")
	}
}

func TestWriteToMapAfterInitialization(t *testing.T) {
	// не очень понял, что требовалось в задании, но написал это
	itemPrices := make(map[string]int64)

	itemPrices["item1"] = 100

	if itemPrices["item1"] != 100 {
		t.Fatalf("expected item1 price 100, got %d", itemPrices["item1"])
	}
}
