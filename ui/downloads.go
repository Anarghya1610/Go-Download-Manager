package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Anarghya1610/gdm/internal/manager"
)

func BuildDownloadUI(mgr *manager.Manager) fyne.CanvasObject {

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Download URL")

	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder("Output file")

	// LIST
	list := widget.NewList(
		func() int {
			return len(mgr.GetTasks())
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("file"),
				widget.NewProgressBar(),
				container.NewHBox(
					widget.NewButton("Pause", nil),
					widget.NewButton("Resume", nil),
					widget.NewButton("Cancel", nil),
				),
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {

			tasks := mgr.GetTasks()
			if i >= len(tasks) {
				return
			}
			task := tasks[i]

			box := obj.(*fyne.Container)

			label := box.Objects[0].(*widget.Label)
			progress := box.Objects[1].(*widget.ProgressBar)
			buttons := box.Objects[2].(*fyne.Container)

			pauseBtn := buttons.Objects[0].(*widget.Button)
			resumeBtn := buttons.Objects[1].(*widget.Button)
			cancelBtn := buttons.Objects[2].(*widget.Button)

			// TEXT
			label.SetText(
				fmt.Sprintf("%s (%s) - %.2f%% - %s",
					task.Output,
					task.GetStatus(),
					task.GetProgress()*100,
					task.GetSpeed(),
				),
			)

			progress.SetValue(task.GetProgress())

			status := task.GetStatus()

			// BUTTON STATES
			switch status {
			case "running":
				pauseBtn.Enable()
				resumeBtn.Disable()
			case "paused":
				pauseBtn.Disable()
				resumeBtn.Enable()
			default:
				pauseBtn.Disable()
				resumeBtn.Disable()
			}

			// ACTIONS

			// Pause = cancel but keep task
			pauseBtn.OnTapped = func() {
				mgr.Pause(task.ID)
			}

			// Resume = re-add same task (metadata resumes)
			resumeBtn.OnTapped = func() {
				mgr.Resume(task.ID)
			}

			// Cancel = remove task completely
			cancelBtn.OnTapped = func() {
				mgr.Cancel(task.ID)
			}
		},
	)

	// SUBMIT FUNCTION
	submit := func() {
		url := urlEntry.Text
		file := fileEntry.Text

		if url == "" || file == "" {
			return
		}

		mgr.AddTask(url, file)

		urlEntry.SetText("")
		fileEntry.SetText("")

		list.Refresh()
	}

	// ADD BUTTON
	addBtn := widget.NewButton("Add Download", submit)

	// ENTER SUPPORT
	urlEntry.OnSubmitted = func(string) {
		submit()
	}
	fileEntry.OnSubmitted = func(string) {
		submit()
	}

	form := widget.NewForm(
		widget.NewFormItem("URL", urlEntry),
		widget.NewFormItem("File", fileEntry),
	)

	// AUTO REFRESH (safe)
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			fyne.Do(func() {
				list.Refresh()
			})
		}
	}()

	return container.NewVBox(
		form,
		addBtn,
		list,
	)
}
