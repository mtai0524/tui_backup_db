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
	Host         string
	Port         string
	User         string
	Password     string
	Database     string
	Type         string
	ConnString   string
	BinaryPath   string
	OutputDir    string
	BackupFormat string
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

		// Determine backup format: user choice > auto-detect
		format := strings.ToLower(strings.TrimSpace(cfg.BackupFormat))
		useBak := false

		switch format {
		case ".bak", "bak":
			useBak = true
		case ".sql", "sql":
			useBak = false
		default:
			// Auto-detect: local server = .bak, remote = .sql
			isLocal := host == "localhost" || host == "127.0.0.1" || host == "." || strings.HasPrefix(host, "(local)")
			useBak = isLocal
		}

		if useBak {
			filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.bak", dbName, timestamp))
			if err := backupSQLServerToBak(bin, host, port, user, password, cfg.Database, filename); err != nil {
				return "", err
			}
		} else {
			filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))
			if err := backupSQLServerToScript(bin, host, port, user, password, cfg.Database, filename); err != nil {
				_ = os.Remove(filename)
				return "", err
			}
		}

		if err := verifyFile(filename, nil); err != nil {
			return "", err
		}

	default:
		return "", fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	return filename, nil
}

// backupSQLServerToBak thực hiện full backup database ra file .bak dùng lệnh BACKUP DATABASE.
func backupSQLServerToBak(bin, host, port, user, password, database, outFile string) error {
	query := fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = N'%s' WITH FORMAT, INIT, NAME = N'Full Backup of %s', SKIP, NOREWIND, NOUNLOAD, STATS = 10",
		database, outFile, database)

	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-Q", query,
	)

	out, err := runSQLCmd(bin, args)
	if err != nil {
		if strings.Contains(out, "Operating system error 3") {
			return fmt.Errorf("SQL Server backup failed (Path Not Found).\n\nLý do: Bạn đang kết nối tới máy chủ remote (%s), nhưng lại yêu cầu lưu file backup vào đường dẫn local (%s). SQL Server không thể nhìn thấy ổ đĩa của máy bạn.\n\nGiải pháp: Sử dụng định dạng .sql cho server remote, hoặc sử dụng đường dẫn mạng (UNC).", host, outFile)
		}
		return fmt.Errorf("SQL Server backup failed: %w\nOutput: %s", err, out)
	}
	return nil
}

// ── SQL Server Script Generation (For Remote-to-Local) ────────────────────────

func backupSQLServerToScript(bin, host, port, user, password, database, outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("cannot create output file %q: %w", outFile, err)
	}
	defer f.Close()

	header := fmt.Sprintf("USE [%s]\nGO\n", database)
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
			writeSQLComment(f, fmt.Sprintf("WARNING: could not export table %s: %v", table, err))
			_, _ = f.WriteString("\n")
		}
	}
	return nil
}

// getSQLServerTables trả về danh sách "schema.table" trong database.
func getSQLServerTables(bin, host, port, user, password, database string) ([]string, error) {
	query := "SET NOCOUNT ON; SELECT TABLE_SCHEMA+'.'+TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE='BASE TABLE' ORDER BY TABLE_SCHEMA, TABLE_NAME;"
	// Loại bỏ -h -1 để tránh lỗi mutually exclusive với SQLCMDMAXVARTYPEWIDTH=0
	args := append(buildSQLServerBaseArgs(host, port, user, password), "-d", database, "-Q", query, "-s", "|")
	out, err := runSQLCmd(bin, args)
	if err != nil {
		return nil, err
	}

	// Dùng cleanSQLCmdOutput để lọc bỏ header và dashes
	return cleanSQLCmdOutput(out), nil
}

type sqlServerColumn struct {
	name, typeName              string
	maxLength, precision, scale int
	nullable, isIdentity        bool
	seed, increment             string
}

func exportSQLServerTable(bin, host, port, user, password, database, table string, f *os.File) error {
	parts := strings.SplitN(table, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid table name %q", table)
	}
	schema, name := parts[0], parts[1]

	cols, _ := getSQLServerColumns(bin, host, port, user, password, database, schema, name)
	if len(cols) == 0 {
		return nil
	}

	pk, _ := getSQLServerPK(bin, host, port, user, password, database, schema, name)
	if err := writeCreateTable(f, schema, name, cols, pk); err != nil {
		return err
	}

	var selectExprs []string
	var colNames []string
	for _, c := range cols {
		colRef := fmt.Sprintf("[%s]", c.name)
		colNames = append(colNames, colRef)
		var expr string
		switch c.typeName {
		case "int", "bigint", "smallint", "tinyint", "bit", "decimal", "numeric", "money", "smallmoney", "float", "real":
			expr = fmt.Sprintf("ISNULL(CONVERT(NVARCHAR(MAX), %s), 'NULL')", colRef)
		case "datetime", "smalldatetime", "datetime2", "datetimeoffset":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s, 121)+''' AS DateTime2)', 'NULL')", colRef)
		case "date":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s, 23)+''' AS Date)', 'NULL')", colRef)
		case "varbinary", "binary", "image":
			expr = fmt.Sprintf("ISNULL('0x'+CONVERT(NVARCHAR(MAX), %s, 2), 'NULL')", colRef)
		default:
			expr = fmt.Sprintf("ISNULL('N'''+REPLACE(CONVERT(NVARCHAR(MAX), %s), '''', '''''')+'''', 'NULL')", colRef)
		}
		selectExprs = append(selectExprs, expr)
	}

	prefix := fmt.Sprintf("INSERT [%s].[%s] (%s) VALUES (", schema, name, strings.Join(colNames, ", "))
	valueExpr := strings.Join(selectExprs, "+', '+")
	dataQuery := fmt.Sprintf("SET NOCOUNT ON; SELECT '%s'+%s+')' FROM [%s].[%s];", strings.ReplaceAll(prefix, "'", "''"), valueExpr, schema, name)

	dataArgs := append(buildSQLServerBaseArgs(host, port, user, password), "-d", database, "-Q", dataQuery)
	dataOut, err := runSQLCmd(bin, dataArgs)
	if err != nil {
		return fmt.Errorf("export data: %w", err)
	}

	hasIdentity := false
	for _, c := range cols {
		if c.isIdentity {
			hasIdentity = true
			break
		}
	}
	if hasIdentity {
		fmt.Fprintf(f, "SET IDENTITY_INSERT [%s].[%s] ON\nGO\n", schema, name)
	}

	// Lọc dữ liệu sạch (loại bỏ header) trước khi ghi vào file
	for _, line := range cleanSQLCmdOutput(dataOut) {
		_, _ = f.WriteString(line + "\n")
	}
	_, _ = f.WriteString("GO\n")

	if hasIdentity {
		fmt.Fprintf(f, "SET IDENTITY_INSERT [%s].[%s] OFF\nGO\n", schema, name)
	}
	return nil
}

func getSQLServerColumns(bin, host, port, user, password, database, schema, name string) ([]sqlServerColumn, error) {
	query := fmt.Sprintf(`SET NOCOUNT ON; SELECT c.name + '|' + LOWER(TYPE_NAME(c.system_type_id)) + '|' + CONVERT(VARCHAR(20), c.max_length) + '|' + CONVERT(VARCHAR(20), c.precision) + '|' + CONVERT(VARCHAR(20), c.scale) + '|' + CONVERT(VARCHAR(1), c.is_nullable) + '|' + CONVERT(VARCHAR(1), c.is_identity) FROM sys.columns c WHERE c.object_id = OBJECT_ID('[%s].[%s]') AND c.is_computed = 0 ORDER BY c.column_id;`, schema, name)
	args := append(buildSQLServerBaseArgs(host, port, user, password), "-d", database, "-Q", query)
	out, _ := runSQLCmd(bin, args)
	var cols []sqlServerColumn
	for _, line := range cleanSQLCmdOutput(out) {
		p := strings.Split(line, "|")
		if len(p) < 7 {
			continue
		}
		cols = append(cols, sqlServerColumn{
			name: p[0], typeName: p[1], maxLength: atoiOr(p[2], 0),
			precision: atoiOr(p[3], 0), scale: atoiOr(p[4], 0),
			nullable: p[5] == "1", isIdentity: p[6] == "1",
		})
	}
	return cols, nil
}

type sqlServerPK struct {
	name      string
	clustered bool
	columns   []string
}

func getSQLServerPK(bin, host, port, user, password, database, schema, name string) (sqlServerPK, error) {
	query := fmt.Sprintf(`SET NOCOUNT ON; SELECT i.name + '|' + i.type_desc + '|' + c.name FROM sys.indexes i JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id WHERE i.is_primary_key = 1 AND i.object_id = OBJECT_ID('[%s].[%s]') ORDER BY ic.key_ordinal;`, schema, name)
	args := append(buildSQLServerBaseArgs(host, port, user, password), "-d", database, "-Q", query)
	out, _ := runSQLCmd(bin, args)
	var pk sqlServerPK
	for _, line := range cleanSQLCmdOutput(out) {
		p := strings.Split(line, "|")
		if len(p) < 3 {
			continue
		}
		if pk.name == "" {
			pk.name = p[0]
			pk.clustered = strings.Contains(p[1], "CLUSTERED")
		}
		pk.columns = append(pk.columns, p[2])
	}
	return pk, nil
}

func writeCreateTable(f *os.File, schema, name string, cols []sqlServerColumn, pk sqlServerPK) error {
	var lines []string
	for _, c := range cols {
		line := fmt.Sprintf("    [%s] %s", c.name, c.typeName)
		if c.isIdentity {
			line += " IDENTITY(1,1)"
		}
		if c.nullable {
			line += " NULL"
		} else {
			line += " NOT NULL"
		}
		lines = append(lines, line)
	}
	if len(pk.columns) > 0 {
		lines = append(lines, fmt.Sprintf("    CONSTRAINT [%s] PRIMARY KEY (%s)", pk.name, strings.Join(pk.columns, ", ")))
	}
	fmt.Fprintf(f, "CREATE TABLE [%s].[%s](\n%s\n)\nGO\n", schema, name, strings.Join(lines, ",\n"))
	return nil
}

func sanitizeSQLCmdArgs(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i := 0; i < len(out)-1; i++ {
		if out[i] == "-P" {
			out[i+1] = "***"
		}
	}
	return out
}

func runSQLCmd(bin string, args []string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "SQLCMDMAXVARTYPEWIDTH=0", "SQLCMDMAXFIXEDTYPEWIDTH=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%w\nOutput: %s", err, string(out))
	}
	return string(out), nil
}

func writeSQLComment(f *os.File, msg string) {
	for _, line := range strings.Split(strings.TrimRight(msg, "\n"), "\n") {
		_, _ = f.WriteString("-- " + line + "\n")
	}
}

func atoiOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	n := 0
	fmt.Sscanf(s, "%d", &n)
	if n == 0 && s != "0" {
		return fallback
	}
	return n
}

var sqlcmdErrorLine = regexp.MustCompile(`^(Msg \d+,|Changed database context|HResult|Sqlcmd:)`)

func cleanSQLCmdOutput(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r\t ")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
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
