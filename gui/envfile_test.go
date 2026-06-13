package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	err := writeEnv(path, map[string]string{
		"BAKDB_OUTPUT_DIR":  "~/backups",
		"BAKDB_BINARY_PATH": "/usr/bin/sqlcmd",
	})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if !strings.Contains(s, "BAKDB_OUTPUT_DIR=~/backups") {
		t.Fatalf("missing output dir line:\n%s", s)
	}
	if !strings.Contains(s, "BAKDB_BINARY_PATH=/usr/bin/sqlcmd") {
		t.Fatalf("missing binary path line:\n%s", s)
	}
}
