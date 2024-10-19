package htmltool

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ExtractTableRowsFromHTML(pageUrl string, html []byte) ([][]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML: %v", err)
	}

	tableIndex := 1
	targetTableIndex := 7
	var csvRow [][]string

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		tableIndex++
		if tableIndex == targetTableIndex {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				var elements []string
				row.Find("td").Each(func(k int, col *goquery.Selection) {
					text := strings.TrimSpace(col.Text())
					if text == "" {
						// Ищем ссылки на документы внутри ячейки
						var docLinks []string
						col.Find("a").Each(func(l int, link *goquery.Selection) {
							if href, exists := link.Attr("href"); exists {
								docLinks = append(docLinks, pageUrl+strings.TrimSpace(href))
							}
						})
						if len(docLinks) != 0 {
							elements = append(elements, strings.Join(docLinks, " "))
						}
					} else {
						elements = append(elements, fixText(text))
					}
				})
				csvRow = append(csvRow, elements)
			})
		}
	})

	return csvRow, nil
}

func fixText(text string) string {
	literFixed := strings.ReplaceAll(text, "ё", "е")
	spaceFixed := strings.Replace(literFixed, "ОТВЕТЧИК", " ОТВЕТЧИК", 1)
	return spaceFixed
}

type PaginationInfo struct {
	TotalElements int
	PerPage       int
}

func ExtractPaginationInfoFromHTML(html []byte) (PaginationInfo, error) {
	totalRegex := regexp.MustCompile(`(\d+)`)
	tableIndex := 1
	targetTableIndex := 6
	targetTdIndex := 2

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return PaginationInfo{}, fmt.Errorf("ошибка при парсинге HTML: %v", err)
	}

	info := PaginationInfo{}

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		tableIndex++
		if tableIndex == targetTableIndex {
			s.Find("tr").Each(func(j int, row *goquery.Selection) {
				tdIndex := 1
				row.Find("td").Each(func(k int, col *goquery.Selection) {
					if tdIndex == targetTdIndex {
						text := strings.TrimSpace(col.Text())
						var totalElementsCount int
						elementsMatched := totalRegex.FindStringSubmatch(text)
						if len(elementsMatched) > 0 {
							totalElementsCount, err = strconv.Atoi(elementsMatched[0])
							if err != nil {
								fmt.Printf("ошибка при получении общего количества элементов: %v", err)
								return
							}
							info.TotalElements = totalElementsCount
						} else {
							fmt.Println("Не удалось найти общее количество элементов.")
							return
						}

						info.PerPage, err = extractPageElementsCount(text)
						if err != nil {
							fmt.Printf("%v", err)
							return
						}

						requests := totalElementsCount / info.PerPage
						if totalElementsCount%info.PerPage != 0 {
							requests++
						}
					}

					tdIndex++
				})
			})
		}
	})
	return info, nil
}

func extractPageElementsCount(text string) (int, error) {
	perPageStr := strings.Replace(strings.Split(text, "по ")[2], ".", "", 1)
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении количества элементов: %v", err)
	}
	return perPage, nil
}
