package gui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bakdb/config"
)

// AppState holds values loaded once at startup and the most recent results,
// shared across views.
type AppState struct {
	Defaults       config.Defaults
	LastBackupFile string // path of the last successful backup, "" if none
}

// loadState reads .env defaults into a fresh AppState.
func loadState() *AppState {
	return &AppState{Defaults: config.Load()}
}

// splitAddresses splits a comma/semicolon separated address list, trimming
// blanks and dropping empties.
func splitAddresses(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if s := strings.TrimSpace(f); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// resolveDir expands a leading "~" and falls back to the user's home dir when
// the input is empty.
func resolveDir(dir string) string {
	dir = strings.TrimSpace(dir)
	home, _ := os.UserHomeDir()
	if dir == "" {
		return home
	}
	if dir == "~" {
		return home
	}
	if strings.HasPrefix(dir, "~/") {
		return filepath.Join(home, dir[2:])
	}
	return dir
}

var backupExts = map[string]bool{".sql": true, ".bak": true, ".gz": true, ".zip": true}

// listBackups returns backup files in dir, newest first.
func listBackups(dir string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(resolveDir(dir))
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !backupExts[strings.ToLower(filepath.Ext(e.Name()))] {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].ModTime().After(infos[j].ModTime()) })
	return infos, nil
}
