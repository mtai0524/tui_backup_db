package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Host       string
	Port       string
	User       string
	Password   string
	Database   string
	Type       string
	ConnString string
	BinaryPath string
	OutputDir  string
}

// ── Connection string parsers ─────────────────────────────────────────────────

func parseSQLServerConnString(connStr, defaultHost, defaultPort, defaultUser, defaultPass string) (host, port, user, password string) {
	host, port, user, password = defaultHost, defaultPort, defaultUser, defaultPass
	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return
	}
	if !strings.Contains(connStr, "=") {
		raw := strings.TrimPrefix(connStr, "tcp:")
		parts := strings.SplitN(raw, ",", 2)
		if parts[0] != "" {
			host = strings.TrimSpace(parts[0])
		}
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			port = strings.TrimSpace(parts[1])
		}
		return
	}
	for _, pair := range strings.Split(connStr, ";") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])
		switch key {
		case "server", "data source":
			raw := strings.TrimPrefix(val, "tcp:")
			parts := strings.SplitN(raw, ",", 2)
			if parts[0] != "" {
				host = strings.TrimSpace(parts[0])
			}
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				port = strings.TrimSpace(parts[1])
			}
		case "port":
			port = val
		case "user id", "uid", "user":
			user = val
		case "password", "pwd":
			password = val
		}
	}
	return
}

func parseMySQLConnString(connStr, defaultHost, defaultPort, defaultUser, defaultPass string) (host, port, user, password string) {
	host, port, user, password = defaultHost, defaultPort, defaultUser, defaultPass
	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return
	}
	raw := strings.TrimPrefix(connStr, "mysql://")
	if raw != connStr {
		atIdx := strings.LastIndex(raw, "@")
		if atIdx > 0 {
			userInfo := raw[:atIdx]
			hostInfo := raw[atIdx+1:]
			if colonIdx := strings.Index(userInfo, ":"); colonIdx >= 0 {
				user = userInfo[:colonIdx]
				password = userInfo[colonIdx+1:]
			} else {
				user = userInfo
			}
			if slashIdx := strings.Index(hostInfo, "/"); slashIdx >= 0 {
				hostInfo = hostInfo[:slashIdx]
			}
			if colonIdx := strings.LastIndex(hostInfo, ":"); colonIdx >= 0 {
				host = hostInfo[:colonIdx]
				port = hostInfo[colonIdx+1:]
			} else {
				host = hostInfo
			}
		}
		return
	}
	atIdx := strings.LastIndex(connStr, "@")
	if atIdx > 0 {
		userInfo := connStr[:atIdx]
		hostInfo := connStr[atIdx+1:]
		if colonIdx := strings.Index(userInfo, ":"); colonIdx >= 0 {
			user = userInfo[:colonIdx]
			password = userInfo[colonIdx+1:]
		} else {
			user = userInfo
		}
		hostInfo = strings.TrimPrefix(hostInfo, "tcp(")
		hostInfo = strings.TrimSuffix(strings.SplitN(hostInfo, ")", 2)[0], ")")
		if slashIdx := strings.Index(hostInfo, "/"); slashIdx >= 0 {
			hostInfo = hostInfo[:slashIdx]
		}
		if colonIdx := strings.LastIndex(hostInfo, ":"); colonIdx >= 0 {
			host = hostInfo[:colonIdx]
			port = hostInfo[colonIdx+1:]
		} else {
			host = hostInfo
		}
	}
	return
}

func parsePostgresConnString(connStr, defaultHost, defaultPort, defaultUser, defaultPass string) (host, port, user, password string) {
	host, port, user, password = defaultHost, defaultPort, defaultUser, defaultPass
	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return
	}
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if strings.HasPrefix(connStr, prefix) {
			raw := strings.TrimPrefix(connStr, prefix)
			atIdx := strings.LastIndex(raw, "@")
			if atIdx > 0 {
				userInfo := raw[:atIdx]
				hostInfo := raw[atIdx+1:]
				if colonIdx := strings.Index(userInfo, ":"); colonIdx >= 0 {
					user = userInfo[:colonIdx]
					password = userInfo[colonIdx+1:]
				} else {
					user = userInfo
				}
				if slashIdx := strings.Index(hostInfo, "/"); slashIdx >= 0 {
					hostInfo = hostInfo[:slashIdx]
				}
				if qIdx := strings.Index(hostInfo, "?"); qIdx >= 0 {
					hostInfo = hostInfo[:qIdx]
				}
				if colonIdx := strings.LastIndex(hostInfo, ":"); colonIdx >= 0 {
					host = hostInfo[:colonIdx]
					port = hostInfo[colonIdx+1:]
				} else {
					host = hostInfo
				}
			}
			return
		}
	}
	if strings.Contains(connStr, "=") && !strings.Contains(connStr, "://") {
		for _, pair := range strings.Fields(connStr) {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}
			switch strings.ToLower(kv[0]) {
			case "host":
				host = kv[1]
			case "port":
				port = kv[1]
			case "user":
				user = kv[1]
			case "password":
				password = kv[1]
			}
		}
	}
	return
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func resolveOutputDir(outputDir string) (string, error) {
	dir := outputDir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			dir = "."
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create output directory %q: %w", dir, err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve output directory: %w", err)
	}
	return abs, nil
}

func buildSQLServerBaseArgs(host, port, user, password string) []string {
	var server string
	if port != "" {
		server = fmt.Sprintf("tcp:%s,%s", host, port)
	} else {
		server = fmt.Sprintf("tcp:%s", host)
	}
	args := []string{"-S", server, "-l", "30", "-C"}
	if user == "" {
		args = append(args, "-E")
	} else {
		args = append(args, "-U", user, "-P", password)
	}
	return args
}

// verifyFile kiểm tra file tồn tại và không rỗng.
func verifyFile(filename string, toolOutput []byte) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf(
			"backup command succeeded but output file not found at:\n  %s\n\nFor remote SQL Server, the .bak file is saved on the server machine.\nUse localhost or set a shared network path (\\\\server\\share).\nTool output: %s",
			filename, string(toolOutput),
		)
	}
	if info.Size() == 0 {
		_ = os.Remove(filename)
		return fmt.Errorf("backup file is empty (0 bytes).\nTool output: %s", string(toolOutput))
	}
	return nil
}

// ── ExecuteBackup ─────────────────────────────────────────────────────────────

func ExecuteBackup(cfg Config) (string, error) {
	timestamp := time.Now().Format("20060102_150405")

	absDir, err := resolveOutputDir(cfg.OutputDir)
	if err != nil {
		return "", err
	}

	dbName := cfg.Database
	if dbName == "" {
		dbName = "backup"
	}

	bin := cfg.BinaryPath
	var filename string

	switch cfg.Type {

	// ── MySQL ─────────────────────────────────────────────────────────────────
	case "MySQL":
		if bin == "" {
			bin = "mysqldump"
		}
		filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))

		host, port, user, password := cfg.Host, cfg.Port, cfg.User, cfg.Password
		if cfg.ConnString != "" {
			host, port, user, password = parseMySQLConnString(cfg.ConnString, host, port, user, password)
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}

		cmd := exec.Command(bin,
			"-h", host, "-P", port,
			"-u", user, fmt.Sprintf("-p%s", password),
			cfg.Database, "-r", filename,
		)
		out, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			return "", fmt.Errorf("backup failed: %v\nOutput: %s", cmdErr, string(out))
		}
		if err := verifyFile(filename, out); err != nil {
			return "", err
		}

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	case "PostgreSQL":
		if bin == "" {
			bin = "pg_dump"
		}
		filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))

		host, port, user, password := cfg.Host, cfg.Port, cfg.User, cfg.Password
		if cfg.ConnString != "" {
			host, port, user, password = parsePostgresConnString(cfg.ConnString, host, port, user, password)
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}

		cmd := exec.Command(bin,
			"-h", host, "-p", port,
			"-U", user, "-f", filename,
			cfg.Database,
		)
		if password != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", password))
		}
		out, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			return "", fmt.Errorf("backup failed: %v\nOutput: %s", cmdErr, string(out))
		}
		if err := verifyFile(filename, out); err != nil {
			return "", err
		}

	// ── SQL Server ────────────────────────────────────────────────────────────
	case "SQL Server":
		if bin == "" {
			bin = "sqlcmd"
		}

		host, port, user, password := cfg.Host, cfg.Port, cfg.User, cfg.Password
		if cfg.ConnString != "" {
			host, port, user, password = parseSQLServerConnString(cfg.ConnString, host, port, user, password)
		}
		if host == "" {
			host = "localhost"
		}

		isLocal := host == "localhost" || host == "127.0.0.1" || host == "."

		if isLocal {
			// Local: BACKUP DATABASE → .bak (SQL Server native format)
			filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.bak", dbName, timestamp))
			backupQuery := fmt.Sprintf(
				"BACKUP DATABASE [%s] TO DISK = N'%s' WITH FORMAT, INIT, STATS = 10",
				cfg.Database, filename,
			)
			args := append(buildSQLServerBaseArgs(host, port, user, password), "-Q", backupQuery)
			cmd := exec.Command(bin, args...)
			out, cmdErr := cmd.CombinedOutput()
			if cmdErr != nil {
				return "", fmt.Errorf("backup failed: %v\nOutput: %s", cmdErr, string(out))
			}
			if err := verifyFile(filename, out); err != nil {
				return "", err
			}

		} else {
			// Remote: chạy sqlcmd và capture stdout trực tiếp vào file
			// Không dùng flag -o vì -o chỉ ghi messages, không ghi SELECT results
			filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))

			if err := backupSQLServerRemote(bin, host, port, user, password, cfg.Database, filename); err != nil {
				return "", err
			}
			if err := verifyFile(filename, nil); err != nil {
				return "", err
			}
		}

	default:
		return "", fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	return filename, nil
}

// backupSQLServerRemote export schema + data từ SQL Server remote ra file SQL local.
// Dùng sqlcmd với Stdout redirect thay vì flag -o.
func backupSQLServerRemote(bin, host, port, user, password, database, outFile string) error {
	// Tạo file output trước để sqlcmd có thể ghi vào
	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("cannot create output file %q: %w", outFile, err)
	}
	defer f.Close()

	// Ghi header
	header := fmt.Sprintf(
		"-- SQL Server backup\n-- Database: %s\n-- Generated: %s\n-- Host: %s\nUSE [%s];\nGO\n\n",
		database, time.Now().Format("2006-01-02 15:04:05"), host, database,
	)
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("cannot write file header: %w", err)
	}

	// Lấy danh sách tables
	tables, err := getSQLServerTables(bin, host, port, user, password, database)
	if err != nil {
		return fmt.Errorf("cannot list tables: %w", err)
	}
	if len(tables) == 0 {
		_, _ = f.WriteString("-- No tables found\n")
		return nil
	}

	// Export từng table
	for _, table := range tables {
		if err := exportSQLServerTable(bin, host, port, user, password, database, table, f); err != nil {
			// Ghi warning vào file thay vì abort toàn bộ
			_, _ = f.WriteString(fmt.Sprintf("-- WARNING: could not export table %s: %v\n\n", table, err))
		}
	}

	return nil
}

// getSQLServerTables trả về danh sách "schema.table" trong database.
func getSQLServerTables(bin, host, port, user, password, database string) ([]string, error) {
	query := fmt.Sprintf(
		"SET NOCOUNT ON; SELECT TABLE_SCHEMA+'.'+TABLE_NAME FROM [%s].INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE' ORDER BY TABLE_SCHEMA, TABLE_NAME;",
		database,
	)
	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", query,
		"-h", "-1", // không in header
		"-W", // trim trailing spaces
	)
	out, err := exec.Command(bin, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%w\nOutput: %s", err, string(out))
	}

	var tables []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || line == "---" {
			continue
		}
		tables = append(tables, line)
	}
	return tables, nil
}

// exportSQLServerTable export một table dưới dạng INSERT statements, ghi vào w.
func exportSQLServerTable(bin, host, port, user, password, database, table string, f *os.File) error {
	// Bước 1: lấy tên columns
	colQuery := fmt.Sprintf(
		"SET NOCOUNT ON; SELECT COLUMN_NAME FROM [%s].INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA+'.'+ TABLE_NAME = '%s' ORDER BY ORDINAL_POSITION;",
		database, table,
	)
	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", colQuery,
		"-h", "-1",
		"-W",
	)
	colOut, err := exec.Command(bin, args...).Output()
	if err != nil {
		return fmt.Errorf("get columns: %w", err)
	}

	var cols []string
	for _, line := range strings.Split(string(colOut), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") || line == "---" {
			continue
		}
		cols = append(cols, line)
	}
	if len(cols) == 0 {
		return nil
	}

	// Bước 2: ghi comment và SET IDENTITY_INSERT
	_, _ = f.WriteString(fmt.Sprintf("-- Table: %s\nSET IDENTITY_INSERT [%s] ON;\n", table, table))

	// Bước 3: SELECT data và ghi dưới dạng INSERT
	// Dùng FOR XML PATH để tránh vấn đề với giá trị chứa dấu phẩy/newline
	colList := "[" + strings.Join(cols, "],[") + "]"
	selectQuery := fmt.Sprintf("SET NOCOUNT ON; SELECT %s FROM [%s].[%s] FOR XML PATH('row'), ROOT('rows');",
		colList,
		strings.Split(table, ".")[0], // schema
		strings.Split(table, ".")[1], // table name
	)
	args2 := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", selectQuery,
		"-h", "-1",
		"-W",
		"-y", "0", // unlimited column width
	)

	cmd := exec.Command(bin, args2...)
	cmd.Stdout = f // ghi thẳng stdout vào file
	errOut, err2 := cmd.Output()
	if err2 != nil {
		return fmt.Errorf("export data: %w\n%s", err2, string(errOut))
	}

	_, _ = f.WriteString(fmt.Sprintf("\nSET IDENTITY_INSERT [%s] OFF;\nGO\n\n", table))
	return nil
}
