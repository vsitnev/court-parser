package htmltool

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func GetHTML(url string) ([]byte, error) {
	client := &http.Client{}
	cleanURL := strings.ReplaceAll(url, "\r\n", "")

	req, err := http.NewRequest("GET", cleanURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка: статус ответа: %s", resp.Status)
	}

	reader, err  := prepareResponseReader(resp.Body, resp.Header.Get("Content-Type"), resp.Header.Get("Content-Encoding"))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении тела ответа: %v", err)
	}
	fixedBody := removeStyles(body)

	return fixedBody, nil
}

func prepareResponseReader(body io.Reader, contentType, encoding string) (io.Reader, error) {
	var reader io.Reader = body
	if strings.Contains(encoding, "gzip") {
		fmt.Println(body)
		gzipReader, err := gzip.NewReader(body)
		if err != nil {
			return nil, fmt.Errorf("ошибка при создании Gzip Reader: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	if strings.Contains(contentType, "charset=") {
		parts := strings.Split(contentType, "charset=")
		if len(parts) > 1 {
			encoding := parts[1]
			if encoding == "windows-1251" {
				dec := charmap.Windows1251.NewDecoder()
				reader = transform.NewReader(reader, dec)
			} else if encoding != "utf-8" {
				return nil, fmt.Errorf("неизвестная кодировка: %s", encoding)
			}
		}
	}
	return reader, nil
}

func removeStyles(html []byte) []byte {
	styleRegex := regexp.MustCompile(`style="[^"]*"`)
	cleanHtml := styleRegex.ReplaceAll(html, []byte{})
	return cleanHtml
}
