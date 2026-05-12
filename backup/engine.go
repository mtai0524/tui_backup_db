package backup

import (
	"fmt"
	"os/exec"
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
	BinaryPath string // Path to mysqldump/pg_dump/sqlcmd
}

// parseSQLServerConnString parses các định dạng connection string phổ biến:
//   - "Server=host,port;Database=db;User Id=sa;Password=xxx"  (ADO.NET style)
//   - "server=host;port=1433;uid=sa;pwd=xxx"                  (ODBC style)
//   - "tcp:host,1433"                                          (sqlcmd style)
//   - "host,1433"                                              (sqlcmd style, no prefix)
func parseSQLServerConnString(connStr, defaultHost, defaultPort, defaultUser, defaultPass string) (host, port, user, password string) {
	host, port, user, password = defaultHost, defaultPort, defaultUser, defaultPass

	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return
	}

	// Dạng sqlcmd thuần: không có dấu "=" → "tcp:host,port" hoặc "host,port"
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

	// Dạng key=value (ADO.NET / ODBC style)
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
			// Có thể dạng "tcp:host,port" hoặc chỉ "host,port"
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
			// "database" và "initial catalog" bỏ qua vì cfg.Database dùng riêng
		}
	}
	return
}

// parseMySQLConnString parses MySQL URI hoặc DSN:
//   - "mysql://user:pass@host:port/dbname"
//   - "user:pass@tcp(host:port)/dbname"
func parseMySQLConnString(connStr, defaultHost, defaultPort, defaultUser, defaultPass string) (host, port, user, password string) {
	host, port, user, password = defaultHost, defaultPort, defaultUser, defaultPass

	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return
	}

	// URI scheme: mysql://user:pass@host:port/dbname
	raw := strings.TrimPrefix(connStr, "mysql://")
	if raw != connStr {
		// Có prefix mysql://
		// Tách userinfo@hostinfo
		atIdx := strings.LastIndex(raw, "@")
		if atIdx > 0 {
			userInfo := raw[:atIdx]
			hostInfo := raw[atIdx+1:]

			// Parse user:pass
			if colonIdx := strings.Index(userInfo, ":"); colonIdx >= 0 {
				user = userInfo[:colonIdx]
				password = userInfo[colonIdx+1:]
			} else {
				user = userInfo
			}

			// Bỏ /dbname
			if slashIdx := strings.Index(hostInfo, "/"); slashIdx >= 0 {
				hostInfo = hostInfo[:slashIdx]
			}

			// Parse host:port
			if colonIdx := strings.LastIndex(hostInfo, ":"); colonIdx >= 0 {
				host = hostInfo[:colonIdx]
				port = hostInfo[colonIdx+1:]
			} else {
				host = hostInfo
			}
		}
		return
	}

	// Go DSN style: user:pass@tcp(host:port)/dbname
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

		// tcp(host:port) → host:port
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

// parsePostgresConnString parses PostgreSQL URI hoặc DSN:
//   - "postgres://user:pass@host:port/dbname"
//   - "postgresql://user:pass@host:port/dbname"
//   - "host=... port=... user=... password=..."  (libpq style)
func parsePostgresConnString(connStr, defaultHost, defaultPort, defaultUser, defaultPass string) (host, port, user, password string) {
	host, port, user, password = defaultHost, defaultPort, defaultUser, defaultPass

	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return
	}

	// URI scheme
	raw := connStr
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if strings.HasPrefix(connStr, prefix) {
			raw = strings.TrimPrefix(connStr, prefix)

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

	// libpq key=value style: "host=... port=... user=... password=..."
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

func ExecuteBackup(cfg Config) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.sql", cfg.Database, timestamp)

	var cmd *exec.Cmd
	bin := cfg.BinaryPath

	switch cfg.Type {
	case "MySQL":
		if bin == "" {
			bin = "mysqldump"
		}

		host := cfg.Host
		port := cfg.Port
		user := cfg.User
		password := cfg.Password

		if cfg.ConnString != "" {
			host, port, user, password = parseMySQLConnString(cfg.ConnString, host, port, user, password)
		}

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}

		args := []string{
			"-h", host,
			"-P", port,
			"-u", user,
			fmt.Sprintf("-p%s", password),
			cfg.Database,
			"-r", filename,
		}

		cmd = exec.Command(bin, args...)

	case "PostgreSQL":
		if bin == "" {
			bin = "pg_dump"
		}

		host := cfg.Host
		port := cfg.Port
		user := cfg.User
		password := cfg.Password

		if cfg.ConnString != "" {
			host, port, user, password = parsePostgresConnString(cfg.ConnString, host, port, user, password)
		}

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}

		args := []string{
			"-h", host,
			"-p", port,
			"-U", user,
			"-f", filename,
			cfg.Database,
		}

		cmd = exec.Command(bin, args...)
		if password != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", password))
		}

	case "SQL Server":
		if bin == "" {
			bin = "sqlcmd"
		}
		filename = fmt.Sprintf("%s_%s.bak", cfg.Database, timestamp)
		backupQuery := fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = N'%s'", cfg.Database, filename)

		host := cfg.Host
		port := cfg.Port
		user := cfg.User
		password := cfg.Password

		if cfg.ConnString != "" {
			host, port, user, password = parseSQLServerConnString(cfg.ConnString, host, port, user, password)
		}

		if host == "" {
			host = "localhost"
		}

		var server string
		if port != "" {
			server = fmt.Sprintf("tcp:%s,%s", host, port)
		} else {
			server = fmt.Sprintf("tcp:%s", host)
		}

		// -C: trust server certificate (bắt buộc với ODBC Driver 18)
		args := []string{"-S", server, "-Q", backupQuery, "-l", "30", "-C"}
		if user == "" {
			args = append(args, "-E") // Windows integrated auth
		} else {
			args = append(args, "-U", user, "-P", password)
		}

		cmd = exec.Command(bin, args...)

	default:
		return "", fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return "", fmt.Errorf("tool '%s' not found. Please ensure it's in your PATH or specify the full path in the settings", bin)
		}
		return "", fmt.Errorf("backup failed: %v\nOutput: %s", err, string(output))
	}

	return filename, nil
}
