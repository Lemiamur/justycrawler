package parser

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// GoqueryParser — реализация Parser через goquery.
type GoqueryParser struct{}

// New создает новый GoqueryParser.
func New() *GoqueryParser {
	return &GoqueryParser{}
}

// ParseLinks реализует интерфейс crawler.Parser.
func (p *GoqueryParser) ParseLinks(baseRawURL string, htmlBody []byte) ([]string, error) {
	baseURL, err := url.Parse(baseRawURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить базовый URL %s: %w", baseRawURL, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать goquery документ: %w", err)
	}

	var links []string
	seen := make(map[string]struct{})

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}

		refURL, parseErr := url.Parse(href)
		if parseErr != nil {
			return
		}
		absURL := baseURL.ResolveReference(refURL)

		if absURL.Scheme != "http" && absURL.Scheme != "https" {
			return
		}

		absURL.Fragment = ""
		finalURL := absURL.String()

		if _, ok := seen[finalURL]; !ok {
			links = append(links, finalURL)
			seen[finalURL] = struct{}{}
		}
	})
	return links, nil
}
