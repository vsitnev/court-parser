package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	parser "github.com/vsitnev/court-parser/internal/parser"
)

func main() {
	fmt.Print("Введите URL: ")
	reader := bufio.NewReader(os.Stdin)
	url, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка ввода URL: ", err)
		return
	}

	parser := parser.NewParser()

	fmt.Println("Получить все записи или в заданном диапазоне? 1 - все, 2 - в диапазоне")
	var mode int
	fmt.Scan(&mode)

	switch mode {
	case 2:
		{
			err = parseInRange(*parser, url)
		}
	case 1:
		{
			err = parseDefault(*parser, url)
		}
	default:
		fmt.Println("Некорректный режим")
		return
	}

	if err != nil {
		fmt.Println("Ошибка парсинга:", err)
	} else {
		fmt.Println("Парсинг завершен.")
	}

	time.Sleep(10 * time.Second)
}

func parseInRange(parser parser.Parser, url string) error {
	fmt.Print("Введите введите диапазон страниц: ")
	var start, end int
	fmt.Scan(&start)
	fmt.Scan(&end)

	if start > end {
		return fmt.Errorf("некорректный диапазон")
	}

	fmt.Println("Считаю количество необходимых запросов...")
	fmt.Println("Запросов необходимо выполнить:", end-start+1)

	if !confirm() {
		return fmt.Errorf("операция отменена пользователем")
	}

	return parser.ParseInRange(url, start, end)
}

func parseDefault(parser parser.Parser, url string) error {
	fmt.Println("Считаю количество необходимых запросов...")
	reqCount, err := parser.GetReqCount(url)
	if err != nil {
		return fmt.Errorf("ошибка при получении количества запросов: %w", err)
	}
	fmt.Println("Запросов необходимо выполнить:", reqCount)

	if !confirm() {
		return fmt.Errorf("операция отменена пользователем")
	}

	return parser.ParseDefault(url, reqCount)
}

func confirm() bool {
	fmt.Println("Продолжлаем? 1 - да, 2 - нет")
	var answer int
	_, err := fmt.Scan(&answer)
	if err != nil {
		fmt.Println("Ошибка ввода")
		return false
	}
	return answer == 1
}
