package main

import (
	"github.com/Anarghya1610/godownloader/internal/manager"
	"github.com/Anarghya1610/godownloader/ui"
)

func main() {
	mgr := manager.NewManager(3)
	ui.StartGUI(mgr)
}
