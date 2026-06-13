package gui

import (
	"os"
	"sort"
	"strings"
)

// writeEnv writes key=value lines (sorted for stable output) to path.
func writeEnv(path string, values map[string]string) error {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# bakdb config — written by the GUI\n")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(values[k])
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}
