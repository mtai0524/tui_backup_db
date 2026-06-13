package gui

import "testing"

func TestBuildBackupConfig(t *testing.T) {
	f := backupForm{
		Type: "MySQL", Host: "h", Port: "3306", User: "u",
		Password: "p", Database: "db", OutputDir: "~/out",
		ConnString: "", BackupFormat: ".sql",
	}
	cfg := f.toConfig()
	if cfg.Type != "MySQL" || cfg.Database != "db" || cfg.Port != "3306" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestBuildEmailConfig(t *testing.T) {
	f := emailForm{
		From: "me@gmail.com", AppPassword: "pw",
		To: "a@x.com, b@y.com", Subject: "S",
	}
	cfg := f.toConfig("/tmp/file.sql", 1234, "db", ".sql")
	if cfg.FromAddress != "me@gmail.com" || len(cfg.ToAddresses) != 2 {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.BackupFileName != "file.sql" || cfg.BackupSize != 1234 {
		t.Fatalf("attachment metadata not set: %+v", cfg)
	}
}
