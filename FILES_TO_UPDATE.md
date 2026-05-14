# 📝 Files to Update in Your Repository

Quick checklist of files that need to be replaced or updated.

---

## ✏️ Files to UPDATE (modify existing)

### 1. `install.sh` - Line 9
**Change from:**
```bash
REPO_URL="https://github.com/yourusername/bakdb"
```

**Change to:**
```bash
REPO_URL="https://github.com/mtai0524/tui_backup_db"
```

**Or use sed:**
```bash
sed -i 's|https://github.com/yourusername/bakdb|https://github.com/mtai0524/tui_backup_db|g' install.sh
```

---

### 2. `README.md` - Multiple lines
**Replace all occurrences:**

```bash
sed -i 's|https://github.com/yourusername/bakdb|https://github.com/mtai0524/tui_backup_db|g' README.md
sed -i 's|yourusername|mtai0524|g' README.md
```

---

### 3. `QUICKSTART.md` - Multiple lines
**Replace all occurrences:**

```bash
sed -i 's|https://github.com/yourusername/bakdb|https://github.com/mtai0524/tui_backup_db|g' QUICKSTART.md
sed -i 's|yourusername|mtai0524|g' QUICKSTART.md
```

---

### 4. `DEPLOYMENT.md` - Multiple lines
**Replace all occurrences:**

```bash
sed -i 's|https://github.com/yourusername/bakdb|https://github.com/mtai0524/tui_backup_db|g' DEPLOYMENT.md
sed -i 's|yourusername|mtai0524|g' DEPLOYMENT.md
```

---

## ➕ Files to ADD (new files)

Copy these new documentation files to your repository:

1. **`START_HERE.md`** - Quick reference guide
2. **`INSTALL.md`** - Detailed installation guide  
3. **`IMPROVEMENTS.md`** - Roadmap and future features
4. **`CONTRIBUTING.md`** - Contribution guidelines
5. **`SUMMARY.md`** - Complete package overview
6. **`FIXES.md`** - Summary of fixes applied

---

## 🚀 Quick Fix Script

Run this to fix everything automatically:

```bash
#!/bin/bash
# Replace all yourusername references
sed -i 's|https://github.com/yourusername/bakdb|https://github.com/mtai0524/tui_backup_db|g' install.sh README.md QUICKSTART.md DEPLOYMENT.md

# Verify
echo "✅ Fixes applied!"
echo ""
echo "Checking install.sh:"
grep "REPO_URL=" install.sh
echo ""
echo "Checking README.md:"
grep -c "mtai0524/tui_backup_db" README.md
echo "occurrences of mtai0524/tui_backup_db found"
```

Save as `fix_urls.sh` and run:
```bash
chmod +x fix_urls.sh
./fix_urls.sh
```

---

## 📋 Manual Fix Checklist

- [ ] Update `install.sh` line 9
- [ ] Update `README.md` GitHub URLs
- [ ] Update `QUICKSTART.md` GitHub URLs
- [ ] Update `DEPLOYMENT.md` GitHub URLs
- [ ] Add `START_HERE.md`
- [ ] Add `INSTALL.md`
- [ ] Add `IMPROVEMENTS.md`
- [ ] Add `CONTRIBUTING.md`
- [ ] Add `SUMMARY.md`
- [ ] Add `FIXES.md`

---

## ✅ Verification Commands

After making changes, run these to verify:

```bash
# Check install.sh
grep "REPO_URL=" install.sh
# Should output: REPO_URL="https://github.com/mtai0524/tui_backup_db"

# Count fixes in README
grep -c "mtai0524/tui_backup_db" README.md

# Count fixes in QUICKSTART
grep -c "mtai0524/tui_backup_db" QUICKSTART.md

# Count fixes in DEPLOYMENT
grep -c "mtai0524/tui_backup_db" DEPLOYMENT.md

# List new files
ls -l START_HERE.md INSTALL.md IMPROVEMENTS.md CONTRIBUTING.md SUMMARY.md FIXES.md
```

---

## 🎯 Git Commit

After making all changes:

```bash
git add install.sh README.md QUICKSTART.md DEPLOYMENT.md \
        START_HERE.md INSTALL.md IMPROVEMENTS.md CONTRIBUTING.md \
        SUMMARY.md FIXES.md FILES_TO_UPDATE.md

git commit -m "fix: correct GitHub repository URLs and add comprehensive documentation

- Fix install.sh hardcoded repo URL (yourusername → mtai0524/tui_backup_db)
- Update README.md with correct GitHub URLs
- Update QUICKSTART.md with correct GitHub URLs
- Update DEPLOYMENT.md with correct GitHub URLs
- Add START_HERE.md quick reference guide
- Add INSTALL.md detailed installation instructions
- Add IMPROVEMENTS.md roadmap and features
- Add CONTRIBUTING.md development guidelines
- Add SUMMARY.md package overview
- Add FIXES.md summary of fixes applied"

git push origin main
```

---

## 🔍 Final Checks

After pushing, verify on GitHub:

1. Open https://github.com/mtai0524/tui_backup_db
2. Click on `install.sh`
3. Find line with `REPO_URL`
4. Verify it shows: `https://github.com/mtai0524/tui_backup_db`
5. Check that new files appear in file list

---

## 📞 Common Issues

**Issue: URL not changing**
- Make sure you're editing the right files
- Check file encoding (should be UTF-8)
- Try using explicit sed syntax

**Issue: Missing new files**
- Make sure you copied all new files
- Check file permissions (should be readable)
- Verify files are in root directory

**Issue: Git commit fails**
- Run `git status` to see what's staged
- Run `git add` again if needed
- Check you have write permission

---

## ✨ You're Done!

Once all files are updated and pushed:

1. ✅ Users can clone your repo
2. ✅ `./install.sh` will work correctly
3. ✅ All documentation has correct URLs
4. ✅ All new guides are available

🎉 Ready to go!

---

## 📝 This File

This file (`FILES_TO_UPDATE.md`) is a quick reference. You can:
- Keep it for future reference
- Delete it after updating
- Share it with collaborators

---

Happy fixing! 🚀
