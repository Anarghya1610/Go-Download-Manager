package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/Anarghya1610/gdm/internal/manager"
)

func StartGUI(mgr *manager.Manager) {
	a := app.New()

	icon, _ := fyne.LoadResourceFromPath("icon.png")
	a.SetIcon(icon)

	w := a.NewWindow("Go Download Manager")

	w.SetIcon(icon)

	content := BuildDownloadUI(w, mgr)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(700, 500))
	w.ShowAndRun()
}
