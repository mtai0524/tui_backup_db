package gui

import (
	"path/filepath"

	"bakdb/backup"
	"bakdb/email"
)

// backupForm is the plain-data view of the Backup form fields.
type backupForm struct {
	Type, Host, Port, User, Password, Database string
	ConnString, OutputDir, BackupFormat        string
}

func (f backupForm) toConfig() backup.Config {
	return backup.Config{
		Type:         f.Type,
		Host:         f.Host,
		Port:         f.Port,
		User:         f.User,
		Password:     f.Password,
		Database:     f.Database,
		ConnString:   f.ConnString,
		OutputDir:    f.OutputDir,
		BackupFormat: f.BackupFormat,
	}
}

// emailForm is the plain-data view of the Email form fields.
type emailForm struct {
	From, AppPassword, To, Subject string
}

func (f emailForm) toConfig(attachmentPath string, size int64, dbName, format string) email.Config {
	return email.Config{
		FromAddress:    f.From,
		AppPassword:    f.AppPassword,
		ToAddresses:    splitAddresses(f.To),
		Subject:        f.Subject,
		BackupFileName: filepath.Base(attachmentPath),
		BackupSize:     size,
		DatabaseName:   dbName,
		BackupFormat:   format,
	}
}
