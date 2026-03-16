package ui

import (
	"fyne.io/fyne/v2/widget"

	"github.com/Anarghya1610/godownloader/internal/manager"
)

func BuildDownloadUI(mgr *manager.Manager) *widget.Form {

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Download URL")

	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder("Output file")

	addBtn := widget.NewButton("Add Download", func() {

		url := urlEntry.Text
		file := fileEntry.Text

		if url == "" || file == "" {
			return
		}

		mgr.AddTask(url, file)

		urlEntry.SetText("")
		fileEntry.SetText("")
	})

	return widget.NewForm(
		widget.NewFormItem("URL", urlEntry),
		widget.NewFormItem("File", fileEntry),
		widget.NewFormItem("", addBtn),
	)
}
