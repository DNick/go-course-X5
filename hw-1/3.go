package main

import "fmt"

func main() {
	var num int
	_, err := fmt.Scan(&num)
	if err != nil {
		fmt.Println("Какая-то ошибка при вводе числа")
		return
	}
	switch {
	case num < 0:
		fmt.Println("Отрицательное число")
	case num < 10_000:
		fmt.Println("small")
	case num < 50_000:
		fmt.Println("regular")
	default:
		fmt.Println("large")
	}
}
