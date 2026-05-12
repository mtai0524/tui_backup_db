package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// expandHome thay thế "~" hoặc "~/..." bằng thư mục home của user.
func expandHome(p string) string {
	if p == "" || p[0] != '~' {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == "~" {
		return home
	}
	if len(p) > 1 && (p[1] == '/' || p[1] == filepath.Separator) {
		return filepath.Join(home, p[2:])
	}
	return p
}

func resolveOutputDir(outputDir string) (string, error) {
	dir := strings.TrimSpace(outputDir)
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			dir = "."
		}
	}
	dir = expandHome(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create output directory %q: %w", dir, err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve output directory: %w", err)
	}
	info, statErr := os.Stat(abs)
	if statErr != nil {
		return "", fmt.Errorf("cannot stat output directory %q: %w", abs, statErr)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("output path %q exists but is not a directory", abs)
	}
	// Kiểm tra quyền ghi bằng cách tạo file tạm.
	probe, err := os.CreateTemp(abs, ".bakdb_probe_*")
	if err != nil {
		return "", fmt.Errorf("output directory %q is not writable: %w", abs, err)
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probeName)
	return abs, nil
}

func buildSQLServerBaseArgs(host, port, user, password string) []string {
	var server string
	if port != "" {
		server = fmt.Sprintf("tcp:%s,%s", host, port)
	} else {
		server = fmt.Sprintf("tcp:%s", host)
	}
	// -b: return non-zero exit code on SQL errors (so we don't silently treat
	// error messages as data).
	args := []string{"-S", server, "-l", "30", "-C", "-b"}
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

	if strings.TrimSpace(cfg.Database) == "" {
		return "", fmt.Errorf("database name is required (please fill in the 'Database Name' field)")
	}
	dbName := cfg.Database

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
			_ = os.Remove(filename)
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
		// Preserve PATH/HOME/etc. – assigning a fresh slice would wipe the
		// inherited environment and break tools that rely on it.
		cmd.Env = os.Environ()
		if password != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", password))
		}
		out, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			_ = os.Remove(filename)
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
			// Remote: dump schema/data qua sqlcmd → file SQL local.
			filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))

			if err := backupSQLServerRemote(bin, host, port, user, password, cfg.Database, filename); err != nil {
				_ = os.Remove(filename)
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

// sqlcmdErrorLine matches sqlcmd diagnostic rows like "Msg 1038, Level 15, ..."
var sqlcmdErrorLine = regexp.MustCompile(`^(Msg \d+,|Changed database context|HResult|Sqlcmd:)`)

// cleanSQLCmdOutput loại bỏ separator lines, header dashes và dòng chẩn đoán
// của sqlcmd để chỉ giữ lại các giá trị thật.
func cleanSQLCmdOutput(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r\t ")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Bỏ header dashes do sqlcmd in ra ("----------").
		if strings.Trim(trimmed, "-") == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, "affected)") {
			continue
		}
		if sqlcmdErrorLine.MatchString(trimmed) {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

// runSQLCmd thực thi sqlcmd và trả về stdout đã làm sạch. Stderr/output đầy đủ
// được trả về kèm error nếu thất bại để dễ debug.
func runSQLCmd(bin string, args []string) (string, error) {
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\nCommand: %s %s\nOutput: %s",
			err, bin, strings.Join(args, " "), string(out))
	}
	return string(out), nil
}

// backupSQLServerRemote export schema + data từ SQL Server remote ra file SQL local.
func backupSQLServerRemote(bin, host, port, user, password, database, outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("cannot create output file %q: %w", outFile, err)
	}
	defer f.Close()

	header := fmt.Sprintf(
		"-- SQL Server backup\n-- Database: %s\n-- Generated: %s\n-- Host: %s\nUSE [%s];\nGO\n\n",
		database, time.Now().Format("2006-01-02 15:04:05"), host, database,
	)
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("cannot write file header: %w", err)
	}

	tables, err := getSQLServerTables(bin, host, port, user, password, database)
	if err != nil {
		return fmt.Errorf("cannot list tables: %w", err)
	}
	if len(tables) == 0 {
		_, _ = f.WriteString("-- No tables found\n")
		return nil
	}

	for _, table := range tables {
		if err := exportSQLServerTable(bin, host, port, user, password, database, table, f); err != nil {
			_, _ = f.WriteString(fmt.Sprintf("-- WARNING: could not export table %s: %v\n\n", table, err))
		}
	}

	return nil
}

// getSQLServerTables trả về danh sách "schema.table" trong database.
func getSQLServerTables(bin, host, port, user, password, database string) ([]string, error) {
	query := "SET NOCOUNT ON; SELECT TABLE_SCHEMA+'.'+TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE' ORDER BY TABLE_SCHEMA, TABLE_NAME;"
	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", query,
		"-h", "-1", // không in header
		"-W",       // trim trailing spaces
		"-s", "|", // dùng | làm separator (an toàn hơn comma)
	)
	out, err := runSQLCmd(bin, args)
	if err != nil {
		return nil, err
	}

	tables := cleanSQLCmdOutput(out)
	// Loại thêm các tên không hợp lệ (không chứa dấu chấm schema.table).
	var valid []string
	for _, t := range tables {
		if strings.Contains(t, ".") {
			valid = append(valid, t)
		}
	}
	return valid, nil
}

// exportSQLServerTable export một table dưới dạng INSERT statements, ghi vào f.
func exportSQLServerTable(bin, host, port, user, password, database, table string, f *os.File) error {
	parts := strings.SplitN(table, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid table name %q", table)
	}
	schema, name := parts[0], parts[1]

	// Lấy cột + kiểu để biết khi nào cần quote/escape.
	colQuery := fmt.Sprintf(
		"SET NOCOUNT ON; SELECT COLUMN_NAME+'|'+DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS "+
			"WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s' ORDER BY ORDINAL_POSITION;",
		strings.ReplaceAll(schema, "'", "''"),
		strings.ReplaceAll(name, "'", "''"),
	)
	colArgs := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", colQuery,
		"-h", "-1",
		"-W",
	)
	colOut, err := runSQLCmd(bin, colArgs)
	if err != nil {
		return fmt.Errorf("get columns: %w", err)
	}

	type column struct{ name, dataType string }
	var cols []column
	for _, line := range cleanSQLCmdOutput(colOut) {
		p := strings.SplitN(line, "|", 2)
		if len(p) != 2 {
			continue
		}
		cols = append(cols, column{name: p[0], dataType: strings.ToLower(p[1])})
	}
	if len(cols) == 0 {
		return nil
	}

	_, _ = f.WriteString(fmt.Sprintf("-- Table: %s\n", table))

	// Build SELECT expression: convert mỗi cột thành literal SQL phù hợp.
	var selectExprs []string
	for _, c := range cols {
		colRef := fmt.Sprintf("[%s]", c.name)
		var expr string
		switch c.dataType {
		case "int", "bigint", "smallint", "tinyint", "bit",
			"decimal", "numeric", "money", "smallmoney", "float", "real":
			expr = fmt.Sprintf("ISNULL(CONVERT(NVARCHAR(MAX), %s), 'NULL')", colRef)
		case "datetime", "smalldatetime", "datetime2", "date", "time", "datetimeoffset":
			expr = fmt.Sprintf("ISNULL(''''+CONVERT(NVARCHAR(MAX), %s, 121)+'''', 'NULL')", colRef)
		case "uniqueidentifier":
			expr = fmt.Sprintf("ISNULL(''''+CONVERT(NVARCHAR(MAX), %s)+'''', 'NULL')", colRef)
		case "varbinary", "binary", "image":
			expr = fmt.Sprintf("ISNULL('0x'+CONVERT(NVARCHAR(MAX), %s, 2), 'NULL')", colRef)
		default:
			// String types — escape single quotes, drop CR/LF to keep each
			// INSERT statement on a single output line.
			expr = fmt.Sprintf(
				"ISNULL(''''+REPLACE(REPLACE(REPLACE(CONVERT(NVARCHAR(MAX), %s), '''', ''''''), CHAR(13), ' '), CHAR(10), ' ')+'''', 'NULL')",
				colRef,
			)
		}
		selectExprs = append(selectExprs, expr)
	}

	colListForInsert := make([]string, 0, len(cols))
	for _, c := range cols {
		colListForInsert = append(colListForInsert, fmt.Sprintf("[%s]", c.name))
	}

	prefix := fmt.Sprintf("INSERT INTO [%s].[%s] (%s) VALUES (",
		schema, name, strings.Join(colListForInsert, ","))

	// Nối các expression bằng + ',' + để output có dạng "a,b,c".
	valueExpr := strings.Join(selectExprs, "+','+")
	selectQuery := fmt.Sprintf(
		"SET NOCOUNT ON; SELECT '%s'+%s+');' FROM [%s].[%s];",
		strings.ReplaceAll(prefix, "'", "''"),
		valueExpr,
		schema, name,
	)

	dataArgs := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", selectQuery,
		"-h", "-1",
		"-W",
		"-y", "0", // unlimited variable-length column width
		"-Y", "0", // unlimited fixed-length column width
	)
	dataOut, err := runSQLCmd(bin, dataArgs)
	if err != nil {
		return fmt.Errorf("export data: %w", err)
	}

	for _, line := range cleanSQLCmdOutput(dataOut) {
		if _, werr := f.WriteString(line + "\n"); werr != nil {
			return fmt.Errorf("write row: %w", werr)
		}
	}
	_, _ = f.WriteString("GO\n\n")
	return nil
}
