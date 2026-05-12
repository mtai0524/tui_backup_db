package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	FromAddress string // Gmail address của người gửi
	AppPassword string // Gmail App Password (không phải password thường)
	ToAddresses []string
	Subject     string
	Body        string
}

// Send gửi email với file đính kèm qua Gmail SMTP (TLS port 465)
func Send(cfg Config, attachmentPath string) error {
	if len(cfg.ToAddresses) == 0 {
		return fmt.Errorf("no recipient email address provided")
	}

	host := "smtp.gmail.com"
	port := "465"
	addr := host + ":" + port

	auth := smtp.PlainAuth("", cfg.FromAddress, cfg.AppPassword, host)

	// Đọc file đính kèm
	fileData, err := os.ReadFile(attachmentPath)
	if err != nil {
		return fmt.Errorf("cannot read backup file: %w", err)
	}
	fileName := filepath.Base(attachmentPath)

	// Build MIME message
	var buf strings.Builder
	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", cfg.FromAddress))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(cfg.ToAddresses, ", ")))

	subject := cfg.Subject
	if subject == "" {
		subject = fmt.Sprintf("Database Backup - %s", fileName)
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")

	// Part 1: body text
	mw := multipart.NewWriter(&buf)
	_ = mw // dùng manual boundary để control format

	body := cfg.Body
	if body == "" {
		body = fmt.Sprintf(
			"Xin chào,\r\n\r\nFile backup database đã được đính kèm trong email này.\r\n\r\nFile: %s\r\nThời gian: %s\r\n\r\nTrân trọng.",
			fileName,
			time.Now().Format("02/01/2006 15:04:05"),
		)
	}

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	buf.WriteString("\r\n")

	// Part 2: attachment
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))

	// Detect content type dựa vào extension
	contentType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".sql":
		contentType = "application/sql"
	case ".bak":
		contentType = "application/octet-stream"
	case ".gz":
		contentType = "application/gzip"
	case ".zip":
		contentType = "application/zip"
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", fmt.Sprintf("%s; name=\"%s\"", contentType, fileName))
	h.Set("Content-Transfer-Encoding", "base64")
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	for k, vs := range h {
		for _, v := range vs {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	buf.WriteString("\r\n")

	// Encode file thành base64, 76 ký tự mỗi dòng (chuẩn MIME)
	encoded := base64.StdEncoding.EncodeToString(fileData)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString(encoded[i:end])
		buf.WriteString("\r\n")
	}

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Kết nối TLS (port 465 dùng implicit TLS, khác với STARTTLS)
	tlsCfg := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("cannot connect to Gmail SMTP: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("cannot create SMTP client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("Gmail authentication failed (check App Password): %w", err)
	}

	if err = client.Mail(cfg.FromAddress); err != nil {
		return fmt.Errorf("SMTP MAIL FROM error: %w", err)
	}

	for _, to := range cfg.ToAddresses {
		to = strings.TrimSpace(to)
		if to == "" {
			continue
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP RCPT TO <%s> error: %w", to, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA error: %w", err)
	}

	_, err = fmt.Fprint(w, buf.String())
	if err != nil {
		return fmt.Errorf("cannot write email body: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("cannot finalize email: %w", err)
	}

	return client.Quit()
}
