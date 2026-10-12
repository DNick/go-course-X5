package main

import (
	"fmt"
)

func main() {
	var num int
	_, err := fmt.Scan(&num)
	if err != nil {
		fmt.Println("Срез пустой")
		return
	}
	mn, mx, sm, count := num, num, num, 1

	for {
		_, err := fmt.Scan(&num)
		if err != nil {
			break
		}
		if num < mn {
			mn = num
		}
		if num > mx {
			mx = num
		}
		sm += num
		count++

	}
	fmt.Println("Минимум: ", mn)
	fmt.Println("Максимум: ", mx)
	fmt.Println("Сумма: ", sm)
	fmt.Println("Среднее: ", float64(sm)/float64(count))
}
