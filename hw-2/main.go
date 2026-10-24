package main

import (
	"errors"
	"fmt"
	"strconv"
)

type Item struct {
	Name         string
	PriceKopecks int64
}

type Order struct {
	ID     int
	Items  []Item
	Status string
}

func NewOrder(id int) (Order, error) {
	if id <= 0 {
		return Order{}, errors.New("id должен быть положительным")
	}
	return Order{ID: id, Status: "new"}, nil
}

func AddItem(order *Order, item Item) error {
	if order == nil {
		return errors.New("заказ не задан")
	}
	if item.PriceKopecks < 0 {
		return errors.New("цена не может быть отрицательной")
	}
	if item.Name == "" {
		return errors.New("пустое имя")
	}
	order.Items = append(order.Items, item)
	return nil
}

func Total(order Order) int64 {
	var result int64
	for _, item := range order.Items {
		result += item.PriceKopecks
	}
	return result
}

func main() {
	statusNames := map[string]string{
		"new":        "Новый",
		"processing": "В обработке",
		"sent":       "Отправлен",
		"delivered":  "Доставлен",
	}
	order, _ := NewOrder(1)

	fmt.Println("Status:", statusNames[order.Status])

	for i := 1; i <= 8; i++ {
		item := Item{
			Name:         strconv.Itoa(i),
			PriceKopecks: int64(i * 10),
		}

		err := AddItem(&order, item)

		if err != nil {
			fmt.Println("Произошла ошибка:", err)
			return
		}
		fmt.Println("Добавлен элемент:", item.Name, "Цена:", item.PriceKopecks)
		fmt.Printf("Текущий статус len=%v cap=%v\n", len(order.Items), cap(order.Items))
	}

}
