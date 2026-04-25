package src

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func ParseToDOMTreeManual(input string) (*Node, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("Input HTML kosong")
	}
	return Parse(raw)
}

func ParseURLToDOMTree(rawURL string) (*Node, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("URL kosong")
	}
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("URL tidak valid: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; CauksuBot/1.0; +https://github.com/) Chrome/120.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengambil data dari URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Server mengembalikan status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Gagal membaca response body: %v", err)
	}

	return ParseToDOMTreeManual(string(body))
}