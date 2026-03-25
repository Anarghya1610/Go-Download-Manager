package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Anarghya1610/gdm/internal/manager"
)

func getFileName(rawURL string) string {

	// 1️⃣ Try GET request (more reliable than HEAD)
	req, err := http.NewRequest("GET", rawURL, nil)
	if err == nil {
		req.Header.Set("Range", "bytes=0-0") // minimal data

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()

			cd := resp.Header.Get("Content-Disposition")

			if cd != "" {

				// 🔥 Handle UTF-8 encoded filename*
				if strings.Contains(cd, "filename*=") {
					parts := strings.Split(cd, "filename*=")
					if len(parts) > 1 {
						val := parts[1]
						// Example: UTF-8''filename.rar
						idx := strings.Index(val, "''")
						if idx != -1 {
							encoded := val[idx+2:]
							decoded, err := url.QueryUnescape(encoded)
							if err == nil {
								return decoded
							}
						}
					}
				}

				// 🔹 Normal filename=
				if strings.Contains(cd, "filename=") {
					parts := strings.Split(cd, "filename=")
					if len(parts) > 1 {
						name := strings.Trim(parts[1], "\"")
						return name
					}
				}
			}
		}
	}

	// 2️⃣ Fallback to URL path
	parsed, err := url.Parse(rawURL)
	if err == nil {
		base := path.Base(parsed.Path)
		if base != "" && base != "/" && base != "." {
			return base
		}
	}

	// 3️⃣ Final fallback
	return "file_" + time.Now().Format("150405")
}

func BuildDownloadUI(w fyne.Window, mgr *manager.Manager) fyne.CanvasObject {

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Download URL")

	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("Select download folder")

	// 🔹 Browse button (Folder picker)
	browseBtn := widget.NewButton("Browse", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			folderEntry.SetText(uri.Path())
		}, w).Show()
	})

	// 🔹 LIST
	list := widget.NewList(
		func() int {
			return len(mgr.GetTasks())
		},
		func() fyne.CanvasObject {
			title := widget.NewLabel("file")
			progress := widget.NewProgressBar()

			pause := widget.NewButton("Pause", nil)
			resume := widget.NewButton("Resume", nil)
			cancel := widget.NewButton("Cancel", nil)

			buttons := container.NewHBox(pause, resume, cancel)

			stats := widget.NewLabel("progress")
			stats.TextStyle = fyne.TextStyle{Italic: true}

			return container.NewVBox(
				title,
				progress,
				stats, // 👈 NEW (below progress bar)
				buttons,
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
			stats := box.Objects[2].(*widget.Label)
			buttons := box.Objects[3].(*fyne.Container)

			pauseBtn := buttons.Objects[0].(*widget.Button)
			resumeBtn := buttons.Objects[1].(*widget.Button)
			cancelBtn := buttons.Objects[2].(*widget.Button)

			// 🔹 TEXT
			fileName := filepath.Base(task.Output)
			fileName = truncate(fileName, 40)

			label.SetText(fmt.Sprintf("%s (%s)", fileName, task.GetStatus()))

			stats.SetText(fmt.Sprintf(
				"%.2f%% • %s",
				task.GetProgress()*100,
				task.GetSpeed(),
			))

			stats.TextStyle = fyne.TextStyle{Italic: true}

			progress.SetValue(task.GetProgress())

			status := task.GetStatus()

			// 🔹 BUTTON STATES
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

			// 🔹 ACTIONS
			id := task.ID

			pauseBtn.OnTapped = func() {
				mgr.Pause(id)
			}

			resumeBtn.OnTapped = func() {
				mgr.Resume(id)
			}

			cancelBtn.OnTapped = func() {
				mgr.Cancel(id)
			}
		},
	)

	// 🔹 SUBMIT FUNCTION
	submit := func() {
		urlStr := urlEntry.Text
		folder := folderEntry.Text

		if urlStr == "" || folder == "" {
			return
		}

		filename := getFileName(urlStr)
		output := folder + "/" + filename

		mgr.AddTask(urlStr, output, true)

		urlEntry.SetText("")
		folderEntry.SetText("")

		list.Refresh()
	}

	addBtn := widget.NewButton("Add Download", submit)

	urlEntry.OnSubmitted = func(string) { submit() }

	// 🔹 FORM
	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("URL", urlEntry),
		),
		container.NewBorder(
			nil, nil, nil, browseBtn,
			folderEntry,
		),
		addBtn,
	)

	// 🔹 AUTO REFRESH
	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			fyne.Do(func() {
				list.Refresh()
			})
		}
	}()

	return container.NewBorder(
		form,
		nil,
		nil,
		nil,
		list,
	)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
