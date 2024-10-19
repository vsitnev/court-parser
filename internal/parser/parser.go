package parser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vsitnev/court-parser/internal/csvtool"
	"github.com/vsitnev/court-parser/internal/htmltool"
	"github.com/vsitnev/court-parser/internal/util"
	"golang.org/x/sync/errgroup"
)

var headers = []string{"№ дела", "Дата поступления", "Категория / Стороны / Суд", "Судья", "Дата решения", "Решение", "Дата вступления в законную силу", "Судебные акты"}

const (
	maxPagesCountInRange = 1000
	maxGorutines         = 3
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) GetReqCount(url string) (int, error) {
	html, err := htmltool.GetHTML(url)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении HTML: %v", err)
	}

	info, err := htmltool.ExtractPaginationInfoFromHTML(html)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении информации о странице: %v", err)
	}

	return сountTotalRequests(info.TotalElements, info.PerPage), nil
}

func сountTotalRequests(totalElements, perPage int) int {
	req := totalElements / perPage
	if totalElements%perPage != 0 {
		req++
	}
	return req
}

func ParseSync(url string, pageCount int) {
	rowBatch := make([][]string, 0)
	for i := 1; i <= pageCount; i++ {
		time.Sleep(1 * time.Second)

		baseHost := strings.Split(url, "/modules.php?")[0]
		fmt.Println("Загружаю страницу: ", i)
		pageUrl := url + "&page=" + fmt.Sprint(i)

		html, err := htmltool.GetHTML(pageUrl)
		if err != nil {
			fmt.Println("Ошибка при получении HTML:", err)
			return
		}
		rows, err := htmltool.ExtractTableRowsFromHTML(baseHost, html)
		if err != nil {
			fmt.Println("Ошибка при парсинге HTML:", err)
			return
		}
		rowBatch = append(rowBatch, rows...)
	}

	err := csvtool.WriteToCSV(headers, rowBatch)
	if err != nil {
		fmt.Println("Ошибка при записи данных в файл:", err)
		return
	}
}

func (p *Parser) ParseInRange(url string, start, end int) error {
	semaphore := make(chan struct{}, maxGorutines)

	results := make(map[int][][]string)
	var mu sync.Mutex
	var g errgroup.Group
	for i := start; i <= end; i++ {
		inc := i
		g.Go(func() error {
			util.GorutineRecover()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			rows, err := p.parse(url, inc)
			if err != nil {
				return fmt.Errorf("ошибка при парсинге страницы %d: %v", inc, err)
			}
		
			mu.Lock()
			results[inc] = rows
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if err := p.writeToFiles(results, start, end); err != nil {
		return fmt.Errorf("ошибка при записи файлов: %v", err)
	}

	return nil
}

func (p *Parser) writeToFiles(data map[int][][]string, start, end int) error {
	rowCount := len(data)
	filesCount := (rowCount + maxPagesCountInRange - 1) / maxPagesCountInRange
	startPage, endPage := start, min(start+maxPagesCountInRange-1, end)

	for i := 0; i < filesCount; i++ {
		fileName := fmt.Sprintf("output_%d-%d.csv", startPage, endPage)

		if err := p.createCSVFile(fileName, data, startPage, endPage); err != nil {
			return err
		}

		startPage = endPage + 1
		endPage = min(startPage+maxPagesCountInRange-1, end)
	}

	return nil
}

func (p *Parser) ParseDefault(url string, pageCount int) error {
	semaphore := make(chan struct{}, maxGorutines)

	results := make(map[int][][]string)
	var mu sync.Mutex
	var g errgroup.Group
	for i := 1; i <= pageCount; i++ {
		inc := i
		g.Go(func() error {
			util.GorutineRecover()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			rows, err := p.parse(url, inc)
			if err != nil {
				return fmt.Errorf("ошибка при парсинге страницы %d: %v", inc, err)
			}

			mu.Lock()
			results[inc] = rows
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	file, err := os.Create("output.csv")
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

	for i := 1; i < pageCount; i++ {
		if err = csvtool.WriteFragmentToCSV(writer, results[i]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parse(url string, page int) ([][]string, error) {
	fmt.Println("Загружаю страницу: ", page)

	var pageUrl string
	if strings.Contains(url, "page=") {
		pageUrl = strings.Replace(url, "page="+fmt.Sprint(page-1), "page="+fmt.Sprint(page), 1)
	} else {
		urlArr := strings.Split(url, "?")
		pageUrl = urlArr[0] + "?page=" + fmt.Sprint(page) + "&" + urlArr[1]
	}
	baseHost := strings.Split(url, "/modules.php?")[0]

	html, err := htmltool.GetHTML(pageUrl)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении HTML: %v", err)
	}

	rows, err := htmltool.ExtractTableRowsFromHTML(baseHost, html)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML: %v", err)
	}

	return rows, nil
}

func (p *Parser) createCSVFile(fileName string, results map[int][][]string, startPage, endPage int) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("ошибка при записи заголовков: %v", err)
	}

	for i := startPage; i <= endPage; i++ {
		if rows, ok := results[i]; ok {
			if err := csvtool.WriteFragmentToCSV(writer, rows); err != nil {
				return fmt.Errorf("ошибка при записи данных в файл: %v", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("ошибка при записи файла: %v", err)
	}

	fmt.Printf("Файл %s успешно создан\n", fileName)
	return nil
}
