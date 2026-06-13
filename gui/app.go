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
