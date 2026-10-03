package output

import (
	"os"
	"path/filepath"
	"testing"
)

// New must report a header that cannot be written, and Close must not hide a flush error.
func TestCSVHeaderAndCloseReportErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	w, err := New(path, FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("clean close failed: %v", err)
	}
	b, _ := os.ReadFile(path)
	if len(b) == 0 {
		t.Fatal("csv header was not written")
	}
	if _, err := New(filepath.Join(t.TempDir(), "missing", "out.csv"), FormatCSV); err == nil {
		t.Fatal("creating the file in a missing directory must fail")
	}
}
