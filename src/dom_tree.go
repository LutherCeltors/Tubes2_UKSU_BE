package src

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ParseToDOMTreeManual(input string) (*Node, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("Input HTML kosong")
	}
	return Parse(raw)
}

func ParseURLToDOMTree(rawURL string) (*Node, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengambil data dari URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server mengembalikan status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Gagal membaca response body: %v", err)
	}

	return ParseToDOMTreeManual(string(body))
}