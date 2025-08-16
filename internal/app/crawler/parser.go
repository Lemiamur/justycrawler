package crawler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Parser interface {
	ParseLinks(baseRawURL, html string) ([]string, error)
}

type GoqueryParser struct{}

func (p *GoqueryParser) ParseLinks(baseRawURL string, html string) ([]string, error) {
	baseURL, err := url.Parse(baseRawURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить базовый URL %s: %w", baseRawURL, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать goquery документ: %w", err)
	}

	var links []string
	seen := make(map[string]struct{}) // Для удаления дубликатов на одной странице

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}

		// Приводим относительную ссылку к абсолютной
		refURL, err := url.Parse(href)
		if err != nil {
			return // Пропускаем невалидные href
		}
		absURL := baseURL.ResolveReference(refURL)

		// Удаляем якорь
		absURL.Fragment = ""

		finalURL := absURL.String()

		// Добавляем только http/https ссылки и избегаем дубликатов
		if strings.HasPrefix(finalURL, "http") {
			if _, ok := seen[finalURL]; !ok {
				links = append(links, finalURL)
				seen[finalURL] = struct{}{}
			}
		}
	})
	return links, nil
}
