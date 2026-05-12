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
		
		args := []string{
			"-h", cfg.Host,
			"-P", cfg.Port,
			"-u", cfg.User,
			fmt.Sprintf("-p%s", cfg.Password),
			cfg.Database,
			"-r", filename,
		}

		// If connection string is provided, try to use it (MySQL 8.0+ supports --uri)
		if cfg.ConnString != "" {
			args = []string{"--uri=" + cfg.ConnString, "-r", filename, cfg.Database}
		}

		cmd = exec.Command(bin, args...)

	case "PostgreSQL":
		if bin == "" {
			bin = "pg_dump"
		}
		
		var args []string
		if cfg.ConnString != "" {
			args = []string{"-d", cfg.ConnString, "-f", filename}
		} else {
			args = []string{
				"-h", cfg.Host,
				"-p", cfg.Port,
				"-U", cfg.User,
				"-f", filename,
				cfg.Database,
			}
		}
		
		cmd = exec.Command(bin, args...)
		if cfg.Password != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Password))
		}

	case "SQL Server":
		if bin == "" {
			bin = "sqlcmd"
		}
		filename = fmt.Sprintf("%s_%s.bak", cfg.Database, timestamp)
		backupQuery := fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = '%s'", cfg.Database, filename)
		
		server := cfg.ConnString
		if server == "" {
			if cfg.Port != "" {
				server = fmt.Sprintf("%s,%s", cfg.Host, cfg.Port)
			} else {
				server = cfg.Host
			}
		}

		if server == "" {
			server = "localhost"
		}
		
		args := []string{"-S", server, "-Q", backupQuery, "-l", "30"}
		if cfg.User == "" {
			args = append(args, "-E")
		} else {
			args = append(args, "-U", cfg.User, "-P", cfg.Password)
		}
		
		cmd = exec.Command(bin, args...)
	default:
		return "", fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return "", fmt.Errorf("tool '%s' not found. Please ensure it's in your PATH or specify the full path in the settings.", bin)
		}
		return "", fmt.Errorf("backup failed: %v\nOutput: %s", err, string(output))
	}

	return filename, nil
}
