package config

import (
	"bufio"
	"os"
	"strings"
)

// Defaults holds values pre-loaded from a .env file (or environment) used to
// pre-fill the TUI form.
type Defaults struct {
	Type       string
	Host       string
	Port       string
	User       string
	Password   string
	Database   string
	ConnString string
	BinaryPath string
	OutputDir  string

	// Email modal
	EmailFrom        string
	EmailAppPassword string
	EmailTo          string
	EmailSubject     string
}

// Recognised environment variable names. They are also the keys allowed in the
// .env file.
const (
	envType       = "BAKDB_TYPE"
	envHost       = "BAKDB_HOST"
	envPort       = "BAKDB_PORT"
	envUser       = "BAKDB_USER"
	envPassword   = "BAKDB_PASSWORD"
	envDatabase   = "BAKDB_DATABASE"
	envConnString = "BAKDB_CONN_STRING"
	envBinaryPath = "BAKDB_BINARY_PATH"
	envOutputDir  = "BAKDB_OUTPUT_DIR"
	envFile       = "BAKDB_ENV_FILE"

	envEmailFrom    = "BAKDB_EMAIL_FROM"
	envEmailAppPass = "BAKDB_EMAIL_APP_PASSWORD"
	envEmailTo      = "BAKDB_EMAIL_TO"
	envEmailSubject = "BAKDB_EMAIL_SUBJECT"
)

// Load reads defaults in this priority order (highest wins):
//  1. Process environment variables
//  2. .env file pointed to by BAKDB_ENV_FILE (if set)
//  3. ".env" in the current working directory
//
// Missing files are ignored silently; the TUI simply starts with empty fields.
func Load() Defaults {
	values := map[string]string{}

	mergeFile(values, ".env")
	if custom := os.Getenv(envFile); custom != "" {
		mergeFile(values, custom)
	}
	for _, k := range []string{
		envType, envHost, envPort, envUser, envPassword,
		envDatabase, envConnString, envBinaryPath, envOutputDir,
		envEmailFrom, envEmailAppPass, envEmailTo, envEmailSubject,
	} {
		if v, ok := os.LookupEnv(k); ok && v != "" {
			values[k] = v
		}
	}

	return Defaults{
		Type:             NormalizeType(values[envType]),
		Host:             values[envHost],
		Port:             values[envPort],
		User:             values[envUser],
		Password:         values[envPassword],
		Database:         values[envDatabase],
		ConnString:       values[envConnString],
		BinaryPath:       values[envBinaryPath],
		OutputDir:        values[envOutputDir],
		EmailFrom:        values[envEmailFrom],
		EmailAppPassword: values[envEmailAppPass],
		EmailTo:          values[envEmailTo],
		EmailSubject:     values[envEmailSubject],
	}
}

// NormalizeType maps free-form input like "sqlserver", "MSSQL", "postgres" to
// the canonical DB type strings used by the rest of the app. Returns "" if the
// input doesn't match a known type.
func NormalizeType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mysql", "mariadb":
		return "MySQL"
	case "postgres", "postgresql", "pg":
		return "PostgreSQL"
	case "sqlserver", "sql server", "mssql", "ms-sql":
		return "SQL Server"
	}
	return ""
}

func mergeFile(dst map[string]string, path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = unquote(val)
		if key == "" {
			continue
		}
		dst[key] = val
	}
}

func unquote(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
