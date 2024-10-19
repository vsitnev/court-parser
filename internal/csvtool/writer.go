package csvtool

import (
	"encoding/csv"
	"fmt"
	"os"
)

func WriteToCSV(headers []string, rows [][]string) error {
	fileName := "output.csv"
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("ошибка при записи заголовков: %v", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("ошибка при записи данных в файл: %v", err)
		}
	}
	fmt.Printf("Данные успешно сохранены в %s", fileName)
	return nil
}


func WriteFragmentToCSV(writer *csv.Writer, rows [][]string) error {
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("ошибка при записи данных в файл: %v", err)
		}
	}
	return nil
}
