# bakdb GUI Desktop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace bakdb's terminal UI with a cross-platform Fyne desktop GUI (sidebar + 4 views: Backup, Email, History, Settings), reusing the existing `backup`/`email`/`config` engine unchanged.

**Architecture:** A new `gui/` package built on Fyne renders a single resizable window with a left sidebar that routes to four content views. Each view collects form input, builds a `backup.Config` or `email.Config`, and calls the existing exported engine functions on a background goroutine (results marshalled back to the UI thread via `fyne.Do`). The `backup/`, `email/`, `config/` packages are not modified. The old `ui/` (Bubble Tea) package is deleted only after the GUI is verified.

**Tech Stack:** Go 1.26, Fyne v2 (`fyne.io/fyne/v2`), existing engine packages (module path `bakdb`).

---

## Reference: engine API the GUI calls (already exists, do NOT change)

- `config.Load() config.Defaults` — fields: `Type, Host, Port, User, Password, Database, ConnString, BinaryPath, OutputDir, BackupFormat, EmailFrom, EmailAppPassword, EmailTo, EmailSubject` (all `string`).
- `config.NormalizeType(string) string` — returns "MySQL" | "PostgreSQL" | "SQL Server" | "".
- `backup.ExecuteBackup(backup.Config) (string, error)` — returns output file path. `backup.Config` fields: `Host, Port, User, Password, Database, Type, ConnString, BinaryPath, OutputDir, BackupFormat` (all `string`).
- `email.Send(email.Config, attachmentPath string) error`. `email.Config` fields: `FromAddress, AppPassword string; ToAddresses []string; Subject, Body string; BackupFileName string; BackupSize int64; DatabaseName, BackupFormat string`.

Module name is `bakdb`, so imports are `bakdb/backup`, `bakdb/email`, `bakdb/config`, `bakdb/gui`.

---

## Task 0: Install Fyne build dependencies (Linux) — PREREQUISITE

**This task is run by the human, not an agent — it needs sudo.** Fyne uses CGo + OpenGL; without these libs the GUI will not compile on Linux.

- [ ] **Step 1: Install system libraries**

Run:
```bash
sudo apt-get update
sudo apt-get install -y gcc pkg-config libgl1-mesa-dev xorg-dev \
  libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
```
Expected: packages install without error.

- [ ] **Step 2: Verify**

Run: `pkg-config --exists gl && echo OK`
Expected: prints `OK`.

> If you cannot install these, the GUI cannot be built on this machine. Stop and resolve before continuing.

---

## Task 1: Add Fyne dependency and a placeholder gui package

**Files:**
- Create: `gui/app.go`
- Modify: `go.mod`, `go.sum` (via `go get`)

- [ ] **Step 1: Add Fyne to the module**

Run:
```bash
go get fyne.io/fyne/v2@v2.7.4
```
Expected: `go.mod` gains `fyne.io/fyne/v2`.

- [ ] **Step 2: Create a minimal gui package that opens an empty window**

Create `gui/app.go`:
```go
package gui

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// Run launches the bakdb desktop application.
func Run() {
	a := app.New()
	w := a.NewWindow("bakdb — Database Backup Manager")
	w.SetContent(widget.NewLabel("bakdb"))
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}
```
NOTE: this references `fyne.NewSize`; add the import. Replace the import block with:
```go
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)
```

- [ ] **Step 3: Build to verify Fyne compiles**

Run: `go build ./gui/`
Expected: no errors. (If it fails on missing C libs, Task 0 was not completed.)

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum gui/app.go
git commit -m "feat(gui): add Fyne dependency and placeholder window"
```

---

## Task 2: AppState + pure helpers (with tests)

This task holds the only non-widget logic, so it is unit-tested. Helpers: split the To string into addresses, resolve an output dir (with `~` expansion), and list backup files newest-first.

**Files:**
- Create: `gui/state.go`
- Test: `gui/state_test.go`

- [ ] **Step 1: Write the failing test**

Create `gui/state_test.go`:
```go
package gui

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSplitAddresses(t *testing.T) {
	got := splitAddresses(" a@x.com, b@y.com ;c@z.com ")
	want := []string{"a@x.com", "b@y.com", "c@z.com"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if len(splitAddresses("   ")) != 0 {
		t.Fatalf("blank should yield no addresses")
	}
}

func TestResolveDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := resolveDir("~/backups"); got != filepath.Join(home, "backups") {
		t.Fatalf("tilde not expanded: %q", got)
	}
	if got := resolveDir(""); got == "" {
		t.Fatalf("empty should fall back to a non-empty default dir")
	}
}

func TestListBackups(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.sql"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.bak"), []byte("xx"), 0o644)
	files, err := listBackups(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 backup files, got %d (%v)", len(files), files)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./gui/ -run 'TestSplitAddresses|TestResolveDir|TestListBackups' -v`
Expected: FAIL — `undefined: splitAddresses` (etc.).

- [ ] **Step 3: Write minimal implementation**

Create `gui/state.go`:
```go
package gui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bakdb/config"
)

// AppState holds values loaded once at startup and the most recent results,
// shared across views.
type AppState struct {
	Defaults       config.Defaults
	LastBackupFile string // path of the last successful backup, "" if none
}

// loadState reads .env defaults into a fresh AppState.
func loadState() *AppState {
	return &AppState{Defaults: config.Load()}
}

// splitAddresses splits a comma/semicolon separated address list, trimming
// blanks and dropping empties.
func splitAddresses(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if s := strings.TrimSpace(f); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// resolveDir expands a leading "~" and falls back to the user's home dir when
// the input is empty.
func resolveDir(dir string) string {
	dir = strings.TrimSpace(dir)
	home, _ := os.UserHomeDir()
	if dir == "" {
		return home
	}
	if dir == "~" {
		return home
	}
	if strings.HasPrefix(dir, "~/") {
		return filepath.Join(home, dir[2:])
	}
	return dir
}

var backupExts = map[string]bool{".sql": true, ".bak": true, ".gz": true, ".zip": true}

// listBackups returns backup files in dir, newest first.
func listBackups(dir string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(resolveDir(dir))
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !backupExts[strings.ToLower(filepath.Ext(e.Name()))] {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].ModTime().After(infos[j].ModTime()) })
	return infos, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./gui/ -run 'TestSplitAddresses|TestResolveDir|TestListBackups' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add gui/state.go gui/state_test.go
git commit -m "feat(gui): add AppState and pure helpers with tests"
```

---

## Task 3: Config builders (with tests)

Pure functions that turn form field values into engine structs. Unit-tested because they encode the field mapping the whole app depends on.

**Files:**
- Create: `gui/builders.go`
- Test: `gui/builders_test.go`

- [ ] **Step 1: Write the failing test**

Create `gui/builders_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./gui/ -run 'TestBuildBackupConfig|TestBuildEmailConfig' -v`
Expected: FAIL — `undefined: backupForm` / `emailForm`.

- [ ] **Step 3: Write minimal implementation**

Create `gui/builders.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./gui/ -run 'TestBuildBackupConfig|TestBuildEmailConfig' -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add gui/builders.go gui/builders_test.go
git commit -m "feat(gui): add backup/email config builders with tests"
```

---

## Task 4: Main window shell with sidebar routing

**Files:**
- Modify: `gui/app.go`

- [ ] **Step 1: Replace app.go with the window shell**

Replace the entire contents of `gui/app.go`:
```go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Run launches the bakdb desktop application.
func Run() {
	a := app.New()
	w := a.NewWindow("bakdb — Database Backup Manager")
	state := loadState()

	content := container.NewStack() // holds the active view

	show := func(o fyne.CanvasObject) {
		content.Objects = []fyne.CanvasObject{o}
		content.Refresh()
	}

	// Forward declaration: views are built lazily so they can switch to each other.
	var showBackup, showEmail, showHistory, showSettings func()
	showBackup = func() { show(newBackupView(w, state, func() { showEmail() })) }
	showEmail = func() { show(newEmailView(w, state)) }
	showHistory = func() { show(newHistoryView(w, state, func() { showEmail() })) }
	showSettings = func() { show(newSettingsView(w, state)) }

	sidebar := container.NewVBox(
		widget.NewButtonWithIcon("Backup", nil, func() { showBackup() }),
		widget.NewButtonWithIcon("Email", nil, func() { showEmail() }),
		widget.NewButtonWithIcon("History", nil, func() { showHistory() }),
		widget.NewButtonWithIcon("Settings", nil, func() { showSettings() }),
	)

	showBackup() // default view

	split := container.NewBorder(nil, nil, sidebar, nil, content)
	w.SetContent(split)
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}
```

- [ ] **Step 2: Build (expected to fail — views not defined yet)**

Run: `go build ./gui/`
Expected: FAIL — `undefined: newBackupView` etc. This is expected; the next tasks define them. Do not commit yet.

> Note: Tasks 5–8 each add one view. After Task 8 the package builds. To keep commits green, this task's commit happens at the end of Task 5 once at least the backup view exists; for Tasks 6–8 stub the others first. To avoid a broken intermediate, create stub views now:

- [ ] **Step 3: Create stub views so the package compiles**

Create `gui/views_stub.go`:
```go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func newBackupView(w fyne.Window, s *AppState, onEmail func()) fyne.CanvasObject {
	return widget.NewLabel("Backup (todo)")
}
func newEmailView(w fyne.Window, s *AppState) fyne.CanvasObject {
	return widget.NewLabel("Email (todo)")
}
func newHistoryView(w fyne.Window, s *AppState, onEmail func()) fyne.CanvasObject {
	return widget.NewLabel("History (todo)")
}
func newSettingsView(w fyne.Window, s *AppState) fyne.CanvasObject {
	return widget.NewLabel("Settings (todo)")
}
```

- [ ] **Step 4: Build to verify the shell compiles**

Run: `go build ./gui/`
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add gui/app.go gui/views_stub.go
git commit -m "feat(gui): add main window with sidebar routing and view stubs"
```

---

## Task 5: Backup view

Replace the backup stub with the real form + background run + result.

**Files:**
- Create: `gui/view_backup.go`
- Modify: `gui/views_stub.go` (remove the `newBackupView` stub)

- [ ] **Step 1: Remove the backup stub**

In `gui/views_stub.go`, delete the `newBackupView` function (keep the other three stubs).

- [ ] **Step 2: Create the backup view**

Create `gui/view_backup.go`:
```go
package gui

import (
	"fmt"
	"os/exec"
	"runtime"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"bakdb/backup"
)

func newBackupView(w fyne.Window, s *AppState, onEmail func()) fyne.CanvasObject {
	d := s.Defaults

	dbType := widget.NewSelect([]string{"MySQL", "PostgreSQL", "SQL Server"}, nil)
	if d.Type != "" {
		dbType.SetSelected(d.Type)
	} else {
		dbType.SetSelected("MySQL")
	}

	host := widget.NewEntry()
	host.SetText(d.Host)
	host.SetPlaceHolder("localhost")
	port := widget.NewEntry()
	port.SetText(d.Port)
	user := widget.NewEntry()
	user.SetText(d.User)
	pass := widget.NewPasswordEntry()
	pass.SetText(d.Password)
	database := widget.NewEntry()
	database.SetText(d.Database)
	conn := widget.NewEntry()
	conn.SetText(d.ConnString)
	conn.SetPlaceHolder("optional — overrides host/port/user/password")
	outDir := widget.NewEntry()
	outDir.SetText(d.OutputDir)
	outDir.SetPlaceHolder("~ (home) if empty")

	format := widget.NewRadioGroup([]string{".bak", ".sql"}, nil)
	if d.BackupFormat != "" {
		format.SetSelected(d.BackupFormat)
	}
	formatItem := widget.NewFormItem("Format (SQL Server)", format)

	form := widget.NewForm(
		widget.NewFormItem("DB type", dbType),
		widget.NewFormItem("Host", host),
		widget.NewFormItem("Port", port),
		widget.NewFormItem("User", user),
		widget.NewFormItem("Password", pass),
		widget.NewFormItem("Database *", database),
		widget.NewFormItem("Connection string", conn),
		widget.NewFormItem("Output dir", outDir),
		formatItem,
	)

	status := widget.NewLabel("")
	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	var sendBtn *widget.Button
	sendBtn = widget.NewButton("Send Email", func() { onEmail() })
	sendBtn.Hide()
	openBtn := widget.NewButton("Open folder", nil)
	openBtn.Hide()

	startBtn := widget.NewButton("⬇ Start Backup", nil)
	startBtn.OnTapped = func() {
		if database.Text == "" {
			status.SetText("⚠ Database name is required")
			return
		}
		cfg := backupForm{
			Type: dbType.Selected, Host: host.Text, Port: port.Text,
			User: user.Text, Password: pass.Text, Database: database.Text,
			ConnString: conn.Text, OutputDir: outDir.Text, BackupFormat: format.Selected,
		}.toConfig()

		startBtn.Disable()
		sendBtn.Hide()
		openBtn.Hide()
		progress.Show()
		status.SetText("⏳ Running backup...")

		go func() {
			path, err := backup.ExecuteBackup(cfg)
			fyne.Do(func() {
				progress.Hide()
				startBtn.Enable()
				if err != nil {
					status.SetText("✖ Backup failed: " + err.Error())
					return
				}
				s.LastBackupFile = path
				status.SetText("✔ Backup successful: " + path)
				sendBtn.Show()
				openBtn.Show()
				openBtn.OnTapped = func() { openInFileManager(filepath.Dir(path)) }
			})
		}()
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("Backup", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		startBtn,
		progress,
		status,
		container.NewHBox(sendBtn, openBtn),
	)
}

// openInFileManager opens dir in the OS file manager.
func openInFileManager(dir string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", dir)
	case "windows":
		cmd = exec.Command("explorer", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	_ = cmd.Start()
	_ = fmt.Sprint("") // keep fmt imported if unused elsewhere
}
```
NOTE: remove the `fmt` import and the `_ = fmt.Sprint("")` line if `go vet` flags `fmt` as unused — they are only there as a guard. Prefer deleting both.

- [ ] **Step 3: Build**

Run: `go build ./gui/`
Expected: no errors. (If `fmt` is unused, delete the import and the guard line, rebuild.)

- [ ] **Step 4: Run existing tests (ensure nothing broke)**

Run: `go test ./gui/ -v`
Expected: PASS (5 tests from Tasks 2–3).

- [ ] **Step 5: Commit**

```bash
git add gui/view_backup.go gui/views_stub.go
git commit -m "feat(gui): implement Backup view with background run and result"
```

---

## Task 6: Email view

**Files:**
- Create: `gui/view_email.go`
- Modify: `gui/views_stub.go` (remove the `newEmailView` stub)

- [ ] **Step 1: Remove the email stub**

In `gui/views_stub.go`, delete the `newEmailView` function.

- [ ] **Step 2: Create the email view**

Create `gui/view_email.go`:
```go
package gui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"bakdb/email"
)

func newEmailView(w fyne.Window, s *AppState) fyne.CanvasObject {
	d := s.Defaults

	from := widget.NewEntry()
	from.SetText(d.EmailFrom)
	from.SetPlaceHolder("you@gmail.com")
	appPass := widget.NewPasswordEntry()
	appPass.SetText(d.EmailAppPassword)
	to := widget.NewEntry()
	to.SetText(d.EmailTo)
	to.SetPlaceHolder("a@x.com, b@y.com")
	subject := widget.NewEntry()
	subject.SetText(d.EmailSubject)

	attach := widget.NewEntry()
	attach.SetText(s.LastBackupFile)
	attach.SetPlaceHolder("path to backup file")

	pickBtn := widget.NewButton("Choose file…", func() {
		dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			attach.SetText(rc.URI().Path())
			_ = rc.Close()
		}, w)
	})

	status := widget.NewLabel("")
	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	sendBtn := widget.NewButton("✉ Send", nil)
	sendBtn.OnTapped = func() {
		if from.Text == "" || appPass.Text == "" || to.Text == "" {
			status.SetText("⚠ From, App Password and To are required")
			return
		}
		if attach.Text == "" {
			status.SetText("⚠ No attachment selected")
			return
		}
		var size int64
		if fi, err := os.Stat(attach.Text); err == nil {
			size = fi.Size()
		}
		cfg := emailForm{
			From: from.Text, AppPassword: appPass.Text,
			To: to.Text, Subject: subject.Text,
		}.toConfig(attach.Text, size, "", "")
		path := attach.Text

		sendBtn.Disable()
		progress.Show()
		status.SetText("⏳ Sending...")
		go func() {
			err := email.Send(cfg, path)
			fyne.Do(func() {
				progress.Hide()
				sendBtn.Enable()
				if err != nil {
					status.SetText("✖ Send failed: " + err.Error())
					return
				}
				status.SetText("✔ Email sent")
			})
		}()
	}

	form := widget.NewForm(
		widget.NewFormItem("Gmail from", from),
		widget.NewFormItem("App Password", appPass),
		widget.NewFormItem("To", to),
		widget.NewFormItem("Subject", subject),
		widget.NewFormItem("Attachment", container.NewBorder(nil, nil, nil, pickBtn, attach)),
	)
	_ = storage.NewFileURI // keep storage import available; remove if unused

	return container.NewVBox(
		widget.NewLabelWithStyle("Email", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		sendBtn,
		progress,
		status,
	)
}
```
NOTE: The `storage` import and the `_ = storage.NewFileURI` guard line are only there to avoid an unused-import error if the file picker path changes. If `go build` reports `storage` unused, delete both the import and the guard line.

- [ ] **Step 3: Build**

Run: `go build ./gui/`
Expected: no errors (delete the `storage` import + guard line if flagged unused).

- [ ] **Step 4: Commit**

```bash
git add gui/view_email.go gui/views_stub.go
git commit -m "feat(gui): implement Email view with attachment picker"
```

---

## Task 7: History view

**Files:**
- Create: `gui/view_history.go`
- Modify: `gui/views_stub.go` (remove the `newHistoryView` stub)

- [ ] **Step 1: Remove the history stub**

In `gui/views_stub.go`, delete the `newHistoryView` function.

- [ ] **Step 2: Create the history view**

Create `gui/view_history.go`:
```go
package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newHistoryView(w fyne.Window, s *AppState, onEmail func()) fyne.CanvasObject {
	dir := s.Defaults.OutputDir
	resolved := resolveDir(dir)

	files, err := listBackups(dir)
	header := widget.NewLabelWithStyle("History — "+resolved, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	if err != nil || len(files) == 0 {
		return container.NewVBox(header, widget.NewLabel("No backup files found in this directory."))
	}

	rows := container.NewVBox()
	for _, fi := range files {
		fi := fi
		full := filepath.Join(resolved, fi.Name())
		label := widget.NewLabel(fmt.Sprintf("%s — %s", fi.Name(), humanSize(fi.Size())))
		sendBtn := widget.NewButton("Send", func() {
			s.LastBackupFile = full
			onEmail()
		})
		openBtn := widget.NewButton("Open", func() { openInFileManager(resolved) })
		rows.Add(container.NewBorder(nil, nil, label, container.NewHBox(sendBtn, openBtn)))
	}

	return container.NewVBox(header, container.NewVScroll(rows))
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
```
NOTE: `container.NewVScroll` is the Fyne v2 scroll container. If the compiler reports it undefined, use `container.NewScroll` instead.

- [ ] **Step 3: Build**

Run: `go build ./gui/`
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add gui/view_history.go gui/views_stub.go
git commit -m "feat(gui): implement History view listing backups newest-first"
```

---

## Task 8: Settings view + remove stub file

**Files:**
- Create: `gui/view_settings.go`
- Create: `gui/envfile.go`
- Test: `gui/envfile_test.go`
- Delete: `gui/views_stub.go` (last stub removed)

- [ ] **Step 1: Write the failing test for .env serialization**

Create `gui/envfile_test.go`:
```go
package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	err := writeEnv(path, map[string]string{
		"BAKDB_OUTPUT_DIR":  "~/backups",
		"BAKDB_BINARY_PATH": "/usr/bin/sqlcmd",
	})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if !strings.Contains(s, "BAKDB_OUTPUT_DIR=~/backups") {
		t.Fatalf("missing output dir line:\n%s", s)
	}
	if !strings.Contains(s, "BAKDB_BINARY_PATH=/usr/bin/sqlcmd") {
		t.Fatalf("missing binary path line:\n%s", s)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./gui/ -run TestWriteEnv -v`
Expected: FAIL — `undefined: writeEnv`.

- [ ] **Step 3: Implement writeEnv**

Create `gui/envfile.go`:
```go
package gui

import (
	"os"
	"sort"
	"strings"
)

// writeEnv writes key=value lines (sorted for stable output) to path.
func writeEnv(path string, values map[string]string) error {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# bakdb config — written by the GUI\n")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(values[k])
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./gui/ -run TestWriteEnv -v`
Expected: PASS.

- [ ] **Step 5: Create the settings view**

Create `gui/view_settings.go`:
```go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"bakdb/config"
)

func newSettingsView(w fyne.Window, s *AppState) fyne.CanvasObject {
	d := s.Defaults

	outDir := widget.NewEntry()
	outDir.SetText(d.OutputDir)
	outDir.SetPlaceHolder("~/backups")
	binPath := widget.NewEntry()
	binPath.SetText(d.BinaryPath)
	binPath.SetPlaceHolder("optional: sqlcmd / mysqldump / pg_dump")

	status := widget.NewLabel("")

	saveBtn := widget.NewButton("💾 Save to .env", func() {
		err := writeEnv(".env", map[string]string{
			"BAKDB_OUTPUT_DIR":  outDir.Text,
			"BAKDB_BINARY_PATH": binPath.Text,
		})
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		s.Defaults = config.Load() // reload so other views see new defaults
		status.SetText("✔ Saved to ./.env")
	})

	form := widget.NewForm(
		widget.NewFormItem("Default output dir", outDir),
		widget.NewFormItem("Binary path", binPath),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		saveBtn,
		status,
	)
}
```

- [ ] **Step 6: Delete the now-empty stub file**

Run: `rm gui/views_stub.go`
(All four `newXView` functions are now defined in their own files.)

- [ ] **Step 7: Build and test**

Run: `go build ./gui/ && go test ./gui/ -v`
Expected: build OK; tests PASS (6 tests total).

- [ ] **Step 8: Commit**

```bash
git add gui/view_settings.go gui/envfile.go gui/envfile_test.go
git rm gui/views_stub.go
git commit -m "feat(gui): implement Settings view and .env writer"
```

---

## Task 9: Switch main.go to the GUI

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Replace main.go**

Replace the entire contents of `main.go`:
```go
package main

import "bakdb/gui"

func main() {
	gui.Run()
}
```

- [ ] **Step 2: Build the whole app**

Run: `go build -o bakdb .`
Expected: no errors, produces `bakdb` binary.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: launch GUI from main instead of the TUI"
```

---

## Task 10: Manual end-to-end verification

No automated test can drive the window; verify by hand **before** deleting the TUI.

- [ ] **Step 1: Launch the app**

Run: `./bakdb`
Expected: a window opens (~900×600) with a left sidebar (Backup/Email/History/Settings) and the Backup view showing, form pre-filled from `.env` if present.

- [ ] **Step 2: Run a backup**

Fill connection details for a reachable DB, click Start Backup.
Expected: progress bar shows, then a success line with the file path and [Open folder]/[Send Email] buttons (or a clear error message if the DB/tool is unavailable).

- [ ] **Step 3: Email + History + Settings**

- Click Send Email → Email view opens with the attachment pre-filled.
- Click History → lists files in the output dir.
- Click Settings → change output dir → Save to .env → confirm a `.env` file is written.

- [ ] **Step 4: Record the result**

If all steps work, the GUI is verified — proceed to remove the TUI. If not, fix issues before Task 11.

---

## Task 11: Remove the old TUI package

The Bubble Tea UI is now unused and the GUI is verified (Task 10). Removing it keeps the build lean.

**Files:**
- Delete: `ui/` (all files)
- Modify: `go.mod` / `go.sum` (tidy)

- [ ] **Step 1: Confirm nothing imports ui**

Run: `grep -rn "bakdb/ui" --include="*.go" .`
Expected: no matches.

- [ ] **Step 2: Delete the package and tidy modules**

Run:
```bash
git rm -r ui/
go mod tidy
```
Expected: `go.mod` drops bubbletea/bubbles/lipgloss requires.

- [ ] **Step 3: Build and test**

Run: `go build -o bakdb . && go test ./...`
Expected: build OK; all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove unused Bubble Tea TUI package"
```

---

## Task 12: Update packaging & docs

**Files:**
- Modify: `bakdb.desktop`, `README.md`, `Makefile`

- [ ] **Step 1: Mark the desktop entry as a windowed app**

In `bakdb.desktop`, change `Terminal=true` to `Terminal=false`.

- [ ] **Step 2: Document build-time system libs and the GUI nature in README**

Add to the README Requirements section (under the existing Go requirement):
```markdown
### Building the GUI (Linux)

bakdb is a desktop app built with [Fyne](https://fyne.io), which needs C/OpenGL
dev libraries to compile on Linux:

```bash
sudo apt-get install -y gcc pkg-config libgl1-mesa-dev xorg-dev \
  libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
```

Cross-compiling Fyne (e.g. building Windows binaries from Linux) requires
[`fyne-cross`](https://github.com/fyne-io/fyne-cross) because of CGo; build on
each target OS for the simplest path.
```

- [ ] **Step 3: Note Fyne cross-compile caveat in the Makefile help**

In the `help:` target of `Makefile`, add after the `make package` line:
```makefile
	@echo "  (note: cross-platform GUI builds need fyne-cross — see README)"
```

- [ ] **Step 4: Commit**

```bash
git add bakdb.desktop README.md Makefile
git commit -m "docs: document GUI build deps and update desktop entry"
```

---

## Notes for the implementer

- **Fyne UI-thread rule:** any update to a widget from a goroutine MUST be wrapped in `fyne.Do(func(){ ... })`. The Backup and Email views already do this; follow the same pattern for any new async work.
- **Unused-import guards:** Tasks 5 and 6 include small `_ = ...` guard lines flagged with NOTEs. If `go build` reports the import unused, delete both the import and the guard line — these are intentional escape hatches, not required code.
- **`container.NewVScroll` vs `NewScroll`:** Fyne renamed scroll constructors across versions. If one is undefined at v2.7.4, use the other.
- **Per the project rule, do NOT run `git commit` automatically.** The commit blocks above are the suggested messages for the user to run themselves.
```
