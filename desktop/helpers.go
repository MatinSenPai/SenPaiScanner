package main

import (
	"os"

	"github.com/atotto/clipboard"
)

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
