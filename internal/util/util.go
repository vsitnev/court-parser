package util

import (
	"fmt"
	"time"
)

func GorutineRecover() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Программа завершилась с ошибкой:", r)
			time.Sleep(30 * time.Second)
		}
	}()
}