package main

import (
	"github.com/Anarghya1610/gdm/internal/manager"
	"github.com/Anarghya1610/gdm/ui"
)

func main() {
	mgr := manager.NewManager(3)
	mgr.LoadTasks()
	ui.StartGUI(mgr)
}
