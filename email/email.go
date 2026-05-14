package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
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

	// Thông tin về backup (tuỳ chọn, để tạo email chuyên nghiệp)
	BackupFileName string
	BackupSize     int64  // kích thước file tính bằng bytes
	DatabaseName   string
	BackupFormat   string // .bak hoặc .sql
}

// buildPlainTextBody tạo plain text body cho email backup
func buildPlainTextBody(fileName string, fileSize int64, dbName, backupFormat string) string {
	sizeStr := formatBytes(fileSize)
	timestamp := time.Now().Format("02/01/2006 15:04:05")

	return fmt.Sprintf(`DATABASE BACKUP NOTIFICATION
================================

Xin chào,

Backup database của bạn đã hoàn tất thành công. File đã được đính kèm trong email này.

THÔNG TIN BACKUP
----------------
Database:    %s
File:        %s
Kích thước:  %s
Định dạng:   %s
Thời gian:   %s

HƯỚNG DẪN RESTORE
-----------------
1. SQL Server (.bak): Sử dụng SQL Server Management Studio → Restore Database → chọn file .bak
2. SQL Script (.sql): Mở file bằng SQL Server Management Studio hoặc sqlcmd và execute các lệnh
3. Kiểm tra dữ liệu trước khi sử dụng ở production

LƯU Ý BẢO MẬT
--------------
• Lưu trữ file này ở nơi an toàn
• Không chia sẻ email này với người không được phép
• Xóa file sau khi đã restore thành công

Nếu bạn có bất kỳ câu hỏi nào, vui lòng liên hệ với IT team.

Trân trọng,
Backup System

---
Email này được gửi tự động vào %s
Database Backup Manager - Enterprise Backup Solution
Confidential - For authorized recipients only`, dbName, fileName, sizeStr, backupFormat, timestamp, timestamp)
}

// buildHTMLBody tạo HTML body chuyên nghiệp cho email backup
func buildHTMLBody(fileName string, fileSize int64, dbName, backupFormat string) string {
	sizeStr := formatBytes(fileSize)
	timestamp := time.Now().Format("02/01/2006 15:04:05")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="vi">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style type="text/css">
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; }
		.header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px 20px; text-align: center; }
		.header h1 { margin: 0; font-size: 28px; font-weight: 600; }
		.content { padding: 40px 20px; background: #f9fafb; }
		.card { background: white; border-left: 4px solid #667eea; padding: 20px; margin-bottom: 20px; border-radius: 6px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
		.card h3 { margin: 0 0 15px 0; color: #667eea; font-size: 16px; text-transform: uppercase; letter-spacing: 0.5px; }
		.info-row { display: flex; justify-content: space-between; margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid #e5e7eb; }
		.info-row:last-child { border-bottom: none; margin-bottom: 0; padding-bottom: 0; }
		.label { color: #6b7280; font-weight: 600; font-size: 14px; }
		.value { color: #111827; font-weight: 500; }
		.highlight { color: #667eea; font-weight: 600; }
		.instructions { background: #eff6ff; border: 1px solid #bfdbfe; color: #1e40af; padding: 15px; border-radius: 6px; margin-bottom: 20px; }
		.instructions h4 { margin: 0 0 10px 0; font-size: 14px; }
		.instructions ol { margin: 0; padding-left: 20px; }
		.instructions li { margin-bottom: 8px; font-size: 13px; }
		.warning { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; border-radius: 6px; margin-bottom: 20px; }
		.warning h4 { margin: 0 0 8px 0; color: #d97706; font-size: 14px; }
		.warning p { margin: 0; color: #92400e; font-size: 13px; }
		.footer { background: #1f2937; color: #d1d5db; padding: 20px; text-align: center; font-size: 12px; }
		.footer p { margin: 5px 0; }
		.footer a { color: #93c5fd; text-decoration: none; }
		.timestamp { color: #9ca3af; font-size: 12px; margin-top: 20px; padding-top: 20px; border-top: 1px solid #e5e7eb; }
	</style>
</head>
<body>
	<div class="container">
		<!-- Header -->
		<div class="header">
			<h1>📦 Database Backup Notification</h1>
		</div>

		<!-- Content -->
		<div class="content">
			<p>Xin chào,</p>
			<p>Backup database của bạn đã hoàn tất thành công. File đã được đính kèm trong email này và sẵn sàng để sử dụng.</p>

			<!-- Thông tin Backup -->
			<div class="card">
				<h3>💾 Thông tin Backup</h3>
				<div class="info-row">
					<span class="label">Database:</span>
					<span class="value"><span class="highlight">%s</span></span>
				</div>
				<div class="info-row">
					<span class="label">File:</span>
					<span class="value">%s</span>
				</div>
				<div class="info-row">
					<span class="label">Kích thước:</span>
					<span class="value">%s</span>
				</div>
				<div class="info-row">
					<span class="label">Định dạng:</span>
					<span class="value">%s</span>
				</div>
				<div class="info-row">
					<span class="label">Thời gian:</span>
					<span class="value">%s</span>
				</div>
			</div>

			<!-- Hướng dẫn Restore -->
			<div class="instructions">
				<h4>📋 Hướng dẫn Restore</h4>
				<ol>
					<li><strong>SQL Server (.bak):</strong> Sử dụng SQL Server Management Studio → Restore Database → chọn file .bak</li>
					<li><strong>SQL Script (.sql):</strong> Mở file bằng SQL Server Management Studio hoặc sqlcmd và execute các lệnh</li>
					<li>Kiểm tra dữ liệu trước khi sử dụng ở production</li>
				</ol>
			</div>

			<!-- Cảnh báo Bảo mật -->
			<div class="warning">
				<h4>⚠️ Lưu ý Bảo mật</h4>
				<p>
					File backup chứa dữ liệu quan trọng của bạn. Vui lòng:<br>
					• Lưu trữ file này ở nơi an toàn<br>
					• Không chia sẻ email này với người không được phép<br>
					• Xóa file sau khi đã restore thành công
				</p>
			</div>

			<p>Nếu bạn có bất kỳ câu hỏi nào, vui lòng liên hệ với IT team.</p>
			<p>Trân trọng,<br><strong>Backup System</strong></p>

			<div class="timestamp">
				✉️ Email này được gửi tự động vào %s
			</div>
		</div>

		<!-- Footer -->
		<div class="footer">
			<p><strong>Database Backup Manager</strong></p>
			<p>Built with bakdb - Enterprise Backup Solution</p>
			<p style="margin-top: 15px; border-top: 1px solid #374151; padding-top: 15px;">
				🔒 Confidential - For authorized recipients only
			</p>
		</div>
	</div>
</body>
</html>`, dbName, fileName, sizeStr, backupFormat, timestamp, timestamp)
}

// formatBytes chuyển đổi số bytes thành readable format (KB, MB, GB)
func formatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
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

	// Lấy thông tin file
	fileInfo, _ := os.Stat(attachmentPath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Build MIME message
	var buf strings.Builder
	mixedBoundary := fmt.Sprintf("boundary_mixed_%d", time.Now().UnixNano())
	altBoundary := fmt.Sprintf("boundary_alt_%d", time.Now().UnixNano())

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", cfg.FromAddress))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(cfg.ToAddresses, ", ")))

	subject := cfg.Subject
	if subject == "" {
		subject = fmt.Sprintf("Database Backup: %s", cfg.DatabaseName)
		if cfg.DatabaseName == "" {
			subject = fmt.Sprintf("Database Backup - %s", fileName)
		}
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mixedBoundary))
	buf.WriteString("\r\n")

	// Part 1: multipart/alternative (plain text + HTML)
	buf.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", altBoundary))
	buf.WriteString("\r\n")

	// Plain text version
	plainBody := cfg.Body
	if plainBody == "" {
		plainBody = buildPlainTextBody(fileName, fileSize, cfg.DatabaseName, cfg.BackupFormat)
	}
	buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(plainBody)
	buf.WriteString("\r\n")

	// HTML version
	htmlBody := buildHTMLBody(fileName, fileSize, cfg.DatabaseName, cfg.BackupFormat)
	buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))

	// Part 2: attachment
	buf.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))

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

	buf.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))

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
