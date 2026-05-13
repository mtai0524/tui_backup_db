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

		// Dump schema + data qua sqlcmd → file .sql paste-and-run được.
		filename = filepath.Join(absDir, fmt.Sprintf("%s_%s.sql", dbName, timestamp))
		if err := backupSQLServerToScript(bin, host, port, user, password, cfg.Database, filename); err != nil {
			_ = os.Remove(filename)
			return "", err
		}
		if err := verifyFile(filename, nil); err != nil {
			return "", err
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

// sanitizeSQLCmdArgs returns a copy of args with secret values (e.g. the value
// after -P) replaced with "***", so the args can safely appear in error
// messages or .sql file comments.
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

// writeSQLComment writes msg as a SQL comment block, prefixing every line with
// "-- " so a multi-line message never breaks the surrounding script.
func writeSQLComment(f *os.File, msg string) {
	for _, line := range strings.Split(strings.TrimRight(msg, "\n"), "\n") {
		_, _ = f.WriteString("-- " + line + "\n")
	}
}

// runSQLCmd thực thi sqlcmd và trả về stdout đã làm sạch. Stderr/output đầy đủ
// được trả về kèm error nếu thất bại để dễ debug.
func runSQLCmd(bin string, args []string) (string, error) {
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\nCommand: %s %s\nOutput: %s",
			err, bin, strings.Join(sanitizeSQLCmdArgs(args), " "), string(out))
	}
	return string(out), nil
}

// backupSQLServerToScript export schema (CREATE TABLE + PK) và data (INSERT)
// của một SQL Server database ra file .sql paste-and-run được.
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

// sqlServerColumn mô tả một cột dùng cho cả CREATE TABLE và INSERT.
type sqlServerColumn struct {
	name       string
	typeName   string // base type, lowercase ("int", "nvarchar", ...)
	maxLength  int    // -1 = MAX; bytes (nvarchar/nchar gấp đôi số ký tự)
	precision  int
	scale      int
	nullable   bool
	isIdentity bool
	seed       string
	increment  string
}

// sqlServerPK describes a table's primary key.
type sqlServerPK struct {
	name      string   // actual constraint name from sys.indexes
	clustered bool     // true if CLUSTERED, false if NONCLUSTERED
	columns   []string // ordered key columns
}

// sqlServerDefault describes one DEFAULT constraint on a column.
type sqlServerDefault struct {
	constraintName string
	columnName     string
	definition     string // includes parentheses, e.g. "(getdate())"
}

// getSQLServerColumns đọc metadata cột (bỏ cột computed) từ sys.columns.
func getSQLServerColumns(bin, host, port, user, password, database, schema, name string) ([]sqlServerColumn, error) {
	sch := strings.ReplaceAll(schema, "'", "''")
	tbl := strings.ReplaceAll(name, "'", "''")
	query := fmt.Sprintf(`SET NOCOUNT ON;
SELECT
  c.name + '|' +
  LOWER(TYPE_NAME(c.system_type_id)) + '|' +
  CONVERT(VARCHAR(20), c.max_length) + '|' +
  CONVERT(VARCHAR(20), c.precision) + '|' +
  CONVERT(VARCHAR(20), c.scale) + '|' +
  CONVERT(VARCHAR(1), c.is_nullable) + '|' +
  CONVERT(VARCHAR(1), c.is_identity) + '|' +
  ISNULL(CONVERT(VARCHAR(50), ic.seed_value), '') + '|' +
  ISNULL(CONVERT(VARCHAR(50), ic.increment_value), '')
FROM sys.columns c
LEFT JOIN sys.identity_columns ic
  ON c.object_id = ic.object_id AND c.column_id = ic.column_id
WHERE c.object_id = OBJECT_ID('[%s].[%s]')
  AND c.is_computed = 0
ORDER BY c.column_id;`, sch, tbl)

	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database, "-Q", query, "-h", "-1", "-y", "0", "-Y", "0",
	)
	out, err := runSQLCmd(bin, args)
	if err != nil {
		return nil, err
	}

	var cols []sqlServerColumn
	for _, line := range cleanSQLCmdOutput(out) {
		parts := strings.Split(line, "|")
		if len(parts) < 9 {
			continue
		}
		c := sqlServerColumn{
			name:       parts[0],
			typeName:   parts[1],
			maxLength:  atoiOr(parts[2], 0),
			precision:  atoiOr(parts[3], 0),
			scale:      atoiOr(parts[4], 0),
			nullable:   parts[5] == "1",
			isIdentity: parts[6] == "1",
			seed:       parts[7],
			increment:  parts[8],
		}
		cols = append(cols, c)
	}
	return cols, nil
}

// getSQLServerPK trả về primary key (tên constraint, clustered/nonclustered,
// danh sách cột theo thứ tự key_ordinal).
func getSQLServerPK(bin, host, port, user, password, database, schema, name string) (sqlServerPK, error) {
	sch := strings.ReplaceAll(schema, "'", "''")
	tbl := strings.ReplaceAll(name, "'", "''")
	query := fmt.Sprintf(`SET NOCOUNT ON;
SELECT i.name + '|' + i.type_desc + '|' + c.name
FROM sys.indexes i
JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id
WHERE i.is_primary_key = 1
  AND i.object_id = OBJECT_ID('[%s].[%s]')
ORDER BY ic.key_ordinal;`, sch, tbl)

	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database, "-Q", query, "-h", "-1", "-W",
	)
	out, err := runSQLCmd(bin, args)
	if err != nil {
		return sqlServerPK{}, err
	}

	var pk sqlServerPK
	for _, line := range cleanSQLCmdOutput(out) {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		if pk.name == "" {
			pk.name = parts[0]
			pk.clustered = strings.EqualFold(parts[1], "CLUSTERED")
		}
		pk.columns = append(pk.columns, parts[2])
	}
	return pk, nil
}

// getSQLServerDefaults trả về danh sách DEFAULT constraint của một table,
// theo thứ tự column_id. Mỗi entry gồm tên constraint, tên cột và definition.
func getSQLServerDefaults(bin, host, port, user, password, database, schema, name string) ([]sqlServerDefault, error) {
	sch := strings.ReplaceAll(schema, "'", "''")
	tbl := strings.ReplaceAll(name, "'", "''")
	// Dùng separator '~|~' để giảm khả năng đụng độ với ký tự trong definition.
	query := fmt.Sprintf(`SET NOCOUNT ON;
SELECT c.name + '~|~' + dc.name + '~|~' + dc.definition
FROM sys.columns c
JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
WHERE c.object_id = OBJECT_ID('[%s].[%s]')
ORDER BY c.column_id;`, sch, tbl)

	args := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database, "-Q", query, "-h", "-1", "-y", "0", "-Y", "0",
	)
	out, err := runSQLCmd(bin, args)
	if err != nil {
		return nil, err
	}

	var defaults []sqlServerDefault
	for _, line := range cleanSQLCmdOutput(out) {
		p := strings.SplitN(line, "~|~", 3)
		if len(p) != 3 {
			continue
		}
		defaults = append(defaults, sqlServerDefault{
			columnName:     p[0],
			constraintName: p[1],
			definition:     p[2],
		})
	}
	return defaults, nil
}

// renderSQLServerType ghép tên type với độ dài/precision/scale phù hợp,
// bọc type name trong [] theo style SSMS Generate Scripts.
func renderSQLServerType(c sqlServerColumn) string {
	t := c.typeName
	switch t {
	case "varchar", "char", "varbinary", "binary":
		if c.maxLength == -1 {
			return fmt.Sprintf("[%s](max)", t)
		}
		return fmt.Sprintf("[%s](%d)", t, c.maxLength)
	case "nvarchar", "nchar":
		if c.maxLength == -1 {
			return fmt.Sprintf("[%s](max)", t)
		}
		return fmt.Sprintf("[%s](%d)", t, c.maxLength/2)
	case "decimal", "numeric":
		return fmt.Sprintf("[%s](%d, %d)", t, c.precision, c.scale)
	case "datetime2", "time", "datetimeoffset":
		return fmt.Sprintf("[%s](%d)", t, c.scale)
	case "float":
		if c.precision != 0 && c.precision != 53 {
			return fmt.Sprintf("[float](%d)", c.precision)
		}
		return "[float]"
	default:
		return fmt.Sprintf("[%s]", t)
	}
}

// tableHasLOB returns true if any column is a LOB type that requires
// TEXTIMAGE_ON to be specified on the table.
func tableHasLOB(cols []sqlServerColumn) bool {
	for _, c := range cols {
		switch c.typeName {
		case "text", "ntext", "image", "xml":
			return true
		case "varchar", "nvarchar", "varbinary":
			if c.maxLength == -1 {
				return true
			}
		}
	}
	return false
}

// writeCreateTable phát sinh object header (SET ANSI_NULLS, SET QUOTED_IDENTIFIER),
// CREATE TABLE (kèm IDENTITY, NULL/NOT NULL, PRIMARY KEY CLUSTERED + WITH options)
// theo style SSMS Generate Scripts.
func writeCreateTable(f *os.File, schema, name string, cols []sqlServerColumn, pk sqlServerPK) error {
	scriptDate := time.Now().Format("1/2/2006 15:04:05")
	header := fmt.Sprintf(
		"/****** Object:  Table [%s].[%s]    Script Date: %s ******/\nSET ANSI_NULLS ON\nGO\nSET QUOTED_IDENTIFIER ON\nGO\n",
		schema, name, scriptDate,
	)
	if _, err := f.WriteString(header); err != nil {
		return err
	}

	var lines []string
	for _, c := range cols {
		line := fmt.Sprintf("    [%s] %s", c.name, renderSQLServerType(c))
		if c.isIdentity {
			seed := c.seed
			if seed == "" {
				seed = "1"
			}
			inc := c.increment
			if inc == "" {
				inc = "1"
			}
			line += fmt.Sprintf(" IDENTITY(%s,%s)", seed, inc)
		}
		if c.nullable {
			line += " NULL"
		} else {
			line += " NOT NULL"
		}
		lines = append(lines, line)
	}
	if len(pk.columns) > 0 {
		pkName := pk.name
		if pkName == "" {
			pkName = fmt.Sprintf("PK_%s", name)
		}
		clustered := "CLUSTERED"
		if !pk.clustered {
			clustered = "NONCLUSTERED"
		}
		var pkCols []string
		for _, p := range pk.columns {
			pkCols = append(pkCols, fmt.Sprintf("    [%s] ASC", p))
		}
		lines = append(lines, fmt.Sprintf(
			" CONSTRAINT [%s] PRIMARY KEY %s \n(\n%s\n)WITH (PAD_INDEX = OFF, STATISTICS_NORECOMPUTE = OFF, IGNORE_DUP_KEY = OFF, ALLOW_ROW_LOCKS = ON, ALLOW_PAGE_LOCKS = ON, OPTIMIZE_FOR_SEQUENTIAL_KEY = OFF) ON [PRIMARY]",
			pkName, clustered, strings.Join(pkCols, ",\n"),
		))
	}

	tableSuffix := " ON [PRIMARY]"
	if tableHasLOB(cols) {
		tableSuffix = " ON [PRIMARY] TEXTIMAGE_ON [PRIMARY]"
	}
	body := fmt.Sprintf("CREATE TABLE [%s].[%s](\n%s\n)%s\nGO\n",
		schema, name, strings.Join(lines, ",\n"), tableSuffix,
	)
	_, err := f.WriteString(body)
	return err
}

// writeDefaultConstraints emits each DEFAULT as an ALTER TABLE ... ADD CONSTRAINT
// statement (SSMS style — DEFAULTs are NOT inlined in CREATE TABLE).
func writeDefaultConstraints(f *os.File, schema, name string, defaults []sqlServerDefault) error {
	for _, d := range defaults {
		cname := d.constraintName
		if cname == "" {
			cname = fmt.Sprintf("DF_%s_%s", name, d.columnName)
		}
		line := fmt.Sprintf(
			"ALTER TABLE [%s].[%s] ADD  CONSTRAINT [%s]  DEFAULT %s FOR [%s]\nGO\n",
			schema, name, cname, d.definition, d.columnName,
		)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

// exportSQLServerTable phát sinh CREATE TABLE + DEFAULTs + INSERTs cho một table,
// theo style SSMS Generate Scripts.
func exportSQLServerTable(bin, host, port, user, password, database, table string, f *os.File) error {
	parts := strings.SplitN(table, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid table name %q", table)
	}
	schema, name := parts[0], parts[1]

	cols, err := getSQLServerColumns(bin, host, port, user, password, database, schema, name)
	if err != nil {
		return fmt.Errorf("get columns: %w", err)
	}
	if len(cols) == 0 {
		return nil
	}

	pk, err := getSQLServerPK(bin, host, port, user, password, database, schema, name)
	if err != nil {
		return fmt.Errorf("get primary key: %w", err)
	}
	defaults, err := getSQLServerDefaults(bin, host, port, user, password, database, schema, name)
	if err != nil {
		return fmt.Errorf("get defaults: %w", err)
	}

	if err := writeCreateTable(f, schema, name, cols, pk); err != nil {
		return fmt.Errorf("write CREATE TABLE: %w", err)
	}
	if err := writeDefaultConstraints(f, schema, name, defaults); err != nil {
		return fmt.Errorf("write DEFAULTs: %w", err)
	}

	// Build SELECT expression: convert mỗi cột thành literal SQL phù hợp với SSMS.
	var selectExprs []string
	for _, c := range cols {
		colRef := fmt.Sprintf("[%s]", c.name)
		var expr string
		switch c.typeName {
		case "int", "bigint", "smallint", "tinyint", "bit",
			"decimal", "numeric", "money", "smallmoney", "float", "real":
			expr = fmt.Sprintf("ISNULL(CONVERT(NVARCHAR(MAX), %s), 'NULL')", colRef)
		case "datetime", "smalldatetime":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s, 121)+''' AS DateTime)', 'NULL')", colRef)
		case "datetime2":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s, 121)+''' AS DateTime2)', 'NULL')", colRef)
		case "date":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s, 23)+''' AS Date)', 'NULL')", colRef)
		case "time":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s)+''' AS Time)', 'NULL')", colRef)
		case "datetimeoffset":
			expr = fmt.Sprintf("ISNULL('CAST(N'''+CONVERT(NVARCHAR(MAX), %s, 121)+''' AS DateTimeOffset)', 'NULL')", colRef)
		case "uniqueidentifier":
			expr = fmt.Sprintf("ISNULL('N'''+CONVERT(NVARCHAR(MAX), %s)+'''', 'NULL')", colRef)
		case "varbinary", "binary", "image":
			expr = fmt.Sprintf("ISNULL('0x'+CONVERT(NVARCHAR(MAX), %s, 2), 'NULL')", colRef)
		case "nvarchar", "nchar", "ntext":
			// N'value' — Unicode literal prefix, escape single quotes, drop CR/LF
			// to keep each INSERT on a single output line.
			expr = fmt.Sprintf(
				"ISNULL('N'''+REPLACE(REPLACE(REPLACE(CONVERT(NVARCHAR(MAX), %s), '''', ''''''), CHAR(13), ' '), CHAR(10), ' ')+'''', 'NULL')",
				colRef,
			)
		default:
			// varchar, char, text, xml — 'value' không có N prefix.
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

	// SSMS style: INSERT [schema].[table] (...) VALUES (...) — không có INTO, không semicolon.
	prefix := fmt.Sprintf("INSERT [%s].[%s] (%s) VALUES (",
		schema, name, strings.Join(colListForInsert, ", "))

	valueExpr := strings.Join(selectExprs, "+', '+")
	selectQuery := fmt.Sprintf(
		"SET NOCOUNT ON; SELECT '%s'+%s+')' FROM [%s].[%s];",
		strings.ReplaceAll(prefix, "'", "''"),
		valueExpr,
		schema, name,
	)

	dataArgs := append(buildSQLServerBaseArgs(host, port, user, password),
		"-d", database,
		"-Q", selectQuery,
		"-h", "-1",
		"-y", "0",
		"-Y", "0",
	)
	dataOut, err := runSQLCmd(bin, dataArgs)
	if err != nil {
		return fmt.Errorf("export data: %w", err)
	}

	rows := cleanSQLCmdOutput(dataOut)
	if len(rows) == 0 {
		return nil
	}

	hasIdentity := false
	for _, c := range cols {
		if c.isIdentity {
			hasIdentity = true
			break
		}
	}
	if hasIdentity {
		_, _ = f.WriteString(fmt.Sprintf("SET IDENTITY_INSERT [%s].[%s] ON \nGO\n", schema, name))
	}
	for _, line := range rows {
		if _, werr := f.WriteString(line + "\n"); werr != nil {
			return fmt.Errorf("write row: %w", werr)
		}
	}
	_, _ = f.WriteString("GO\n")
	if hasIdentity {
		_, _ = f.WriteString(fmt.Sprintf("SET IDENTITY_INSERT [%s].[%s] OFF\nGO\n", schema, name))
	}
	return nil
}

// atoiOr parses s as int, returning fallback on error or empty input.
func atoiOr(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	n := 0
	sign := 1
	for i, r := range s {
		if i == 0 && r == '-' {
			sign = -1
			continue
		}
		if r < '0' || r > '9' {
			return fallback
		}
		n = n*10 + int(r-'0')
	}
	return sign * n
}
