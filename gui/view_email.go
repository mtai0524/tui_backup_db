package gui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	return container.NewVBox(
		widget.NewLabelWithStyle("Email", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		sendBtn,
		progress,
		status,
	)
}
