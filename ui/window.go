package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/Anarghya1610/gdm/internal/manager"
)

func StartGUI(mgr *manager.Manager) {
	a := app.New()
	w := a.NewWindow("Go Download Manager")
	content := BuildDownloadUI(mgr)
	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(700, 500))
	w.ShowAndRun()
}
