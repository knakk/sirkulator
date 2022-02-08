package marc

import (
	"os"
	"path/filepath"
	"testing"
)

// Verify that MARC records can be decoded.
func TestDecode(t *testing.T) {
	files, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		recs, err := NewDecoder(f).DecodeAll()
		if err != nil {
			t.Errorf("decoding %v: %v", file, err)
		}
		if len(recs) != 1 {
			t.Errorf("got %d, expected 1 record per file", len(recs))
		}
		f.Close()
	}
}
