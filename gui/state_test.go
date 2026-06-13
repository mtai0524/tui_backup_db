package gui

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSplitAddresses(t *testing.T) {
	got := splitAddresses(" a@x.com, b@y.com ;c@z.com ")
	want := []string{"a@x.com", "b@y.com", "c@z.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if len(splitAddresses("   ")) != 0 {
		t.Fatalf("blank should yield no addresses")
	}
}

func TestResolveDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := resolveDir("~/backups"); got != filepath.Join(home, "backups") {
		t.Fatalf("tilde not expanded: %q", got)
	}
	if got := resolveDir(""); got == "" {
		t.Fatalf("empty should fall back to a non-empty default dir")
	}
}

func TestListBackups(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.sql"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.bak"), []byte("xx"), 0o644)
	files, err := listBackups(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 backup files, got %d (%v)", len(files), files)
	}
}
