package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseToDOMTreeManual(input string) (*Node, error) {
	raw:=strings.TrimSpace(input)
	if raw=="" {
		return nil, fmt.Errorf("Input HTML kosong")
	}
	return Parse(raw)
}

func ParseHTMLToDOMTreeInput(filePath string) (*Node, error) {
	path:=strings.TrimSpace(filePath)
	if path=="" {
		return nil, fmt.Errorf("File kosong")
	}

	extension:=strings.ToLower(filepath.Ext(path))
	if extension!=".html" {
		return nil, fmt.Errorf("Format input invalid")
	}

	content, err:=os.ReadFile(path)
	if err!=nil {
		return nil, fmt.Errorf("Gagal membaca file")
	}

	return parseToDOMTreeManual(string(content))
}