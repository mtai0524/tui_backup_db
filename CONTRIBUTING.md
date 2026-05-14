# 🤝 Contributing to bakdb

Thank you for considering contributing to bakdb! We welcome contributions from everyone.

---

## 📋 Code of Conduct

Be respectful, inclusive, and professional. We're building a welcoming community.

---

## 🐛 Reporting Bugs

Found a bug? Please open a GitHub issue with:

1. **Title**: Clear, descriptive title
2. **Description**: What happened and what did you expect?
3. **Steps to Reproduce**: How to replicate the issue
4. **Environment**: OS, Go version, database type
5. **Logs**: Relevant error messages

**Example:**
```
Title: Email fails with "connection refused"

Description:
Trying to send backup via Gmail but getting connection error.

Steps to Reproduce:
1. Set BAKDB_EMAIL_FROM and BAKDB_EMAIL_APP_PASSWORD
2. Run backup
3. Press 'e' to send email
4. Get error: "cannot connect to Gmail SMTP"

Environment:
- OS: Ubuntu 20.04
- Go: 1.21
- Database: MySQL

Error:
"gmail authentication failed (check App Password): connection refused"
```

---

## 💡 Suggesting Enhancements

Have an idea? Open a GitHub discussion or issue:

1. **Describe the enhancement**: What problem does it solve?
2. **Use case**: Why would this be useful?
3. **Alternatives**: How else could this be done?
4. **Additional context**: Anything else relevant?

**Example:**
```
Title: Add backup compression support

Describe the enhancement:
Allow automatic compression of .sql files to .gz format

Use case:
Large databases produce huge .sql files. Compression would:
- Save storage space
- Faster email transmission
- Better for cloud backups

Alternatives:
- Manual compression (inconvenient)
- External script (extra work)
- Already handled in other tools (we can do it too)
```

---

## 🚀 Getting Started

### 1. Fork & Clone

```bash
# Fork on GitHub, then:
git clone https://github.com/YOUR_USERNAME/tui_backup_db.git
cd tui_backup_db
```

### 2. Create Feature Branch

```bash
git checkout -b feature/my-awesome-feature
```

Branch naming:
- `feature/` for new features
- `bugfix/` for bug fixes
- `docs/` for documentation
- `refactor/` for code improvements

### 3. Setup Development Environment

```bash
# Install dev tools
go mod download
go get -u golang.org/x/lint/golint
go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Verify build
make build
```

### 4. Make Changes

- Write clean, readable code
- Add comments for complex logic
- Follow existing code style
- Use meaningful variable names

**Code Style Tips:**
- Use `gofmt` to format code
- Follow Go idioms (simplicity, clarity)
- Keep functions small and focused
- Add error handling

### 5. Test Your Changes

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code
golangci-lint run

# Build for current platform
make build

# Manual testing
./build/bakdb
```

### 6. Commit & Push

```bash
# Commit with clear message
git commit -m "feature: add backup compression support

- Add -compress flag
- Support .gz and .zip formats
- Update documentation
- Add tests"

# Push to your fork
git push origin feature/my-awesome-feature
```

**Commit Message Guidelines:**
- Use imperative mood ("add" not "added")
- First line should be summary (50 chars max)
- Blank line, then detailed description
- Reference issues: "Fixes #123"

### 7. Open Pull Request

1. Go to GitHub
2. Click "New Pull Request"
3. Fill in title and description
4. Link to any related issues
5. Submit!

**PR Description Template:**
```markdown
## Description
What does this PR do?

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation
- [ ] Refactoring

## Related Issues
Fixes #123

## Testing
How was this tested?

## Checklist
- [ ] Code follows style guidelines
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes
```

---

## 📝 Development Workflow

### Project Structure

```
bakdb/
├── main.go              # Entry point
├── ui/                  # TUI components
│   ├── model.go        # Bubble Tea state
│   ├── views.go        # Rendering
│   ├── updates.go      # Event handling
│   ├── email_modal.go  # Email UI
│   └── styles.go       # Styling
├── backup/              # Core logic
│   └── engine.go       # Backup implementation
├── email/               # Email utilities
│   └── email.go        # SMTP & formatting
├── config/              # Configuration
│   └── env.go          # .env parsing
└── [tests, docs, build configs...]
```

### Key Files to Know

1. **engine.go** - Database backup logic
   - `ExecuteBackup()` - Main function
   - Database-specific implementations
   - File handling

2. **email.go** - Email sending
   - `Send()` - Main function
   - HTML/text formatting
   - SMTP connection

3. **model.go** - TUI state
   - `Model` struct - App state
   - `InitialModel()` - Setup

4. **updates.go** - Event handling
   - `updateEnterDetails()` - Form handling
   - `updateBackingUp()` - Backup process
   - `startBackupCmd()` - Execute backup

---

## 🧪 Testing

### Run Tests

```bash
go test ./...
```

### Write Tests

```go
// example_test.go
package main

import "testing"

func TestBackup(t *testing.T) {
    // Arrange
    cfg := backup.Config{
        Type:     "MySQL",
        Host:     "localhost",
        Database: "test_db",
    }
    
    // Act
    result, err := backup.ExecuteBackup(cfg)
    
    // Assert
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if result == "" {
        t.Error("Expected backup file path, got empty string")
    }
}
```

---

## 📚 Documentation

### Update Documentation When:
- Adding new features
- Changing behavior
- Fixing unclear instructions
- Improving examples

### Documentation Files

- **README.md** - Overview & quick start
- **INSTALL.md** - Installation instructions
- **QUICKSTART.md** - 5-minute guide
- **DEPLOYMENT.md** - Production setup
- **IMPROVEMENTS.md** - Roadmap & ideas

### Adding to Documentation

1. Find relevant file
2. Add/update section
3. Keep format consistent
4. Test markdown rendering
5. Include in PR description

---

## 🎯 Areas for Contribution

### Easy (Good for First-Time Contributors)
- Documentation improvements
- Fix typos
- Add examples
- Improve error messages
- Update screenshots/demos

### Medium
- Bug fixes
- Code refactoring
- Add tests
- Performance improvements
- Small features

### Advanced
- New database support
- Cloud storage integration
- Scheduling system
- Web UI
- Architecture improvements

---

## ❓ Questions?

- 📖 Check existing documentation
- 🔍 Search existing issues
- 💬 Open a discussion
- 🙋 Ask in pull request comments

---

## 🎉 Thank You!

Your contributions make bakdb better for everyone! 🚀

---

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

## 🔐 Security

Found a security issue? **Don't** open a public issue. Email: [security contact]

---

## 📊 Contribution Tips

- Start with small changes
- One feature per PR
- Keep PRs focused
- Discuss large changes first
- Be open to feedback
- Help review other PRs

---

## ✨ Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in CHANGELOG.md
- Thanked in release notes

Again, thanks for contributing! 🙏
