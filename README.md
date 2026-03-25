<p align="center">
	<h1 align="center">GDM - Go Download Manager</h1>
	<p align="center">High-Speed desktop download manager in Go.</p>
</p>

<p align="center">
	<a href="https://github.com/Anarghya1610/gdm"><img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white" alt="Go Version" /></a>
	<a href="https://github.com/Anarghya1610/gdm"><img src="https://img.shields.io/badge/Platform-Windows-2ea44f" alt="Platforms" /></a>
</p>

GDM helps you download files faster with parallel chunks, and pause/resume controls in a desktop app.

## Features

- Parallel chunk downloads that provide up to 3-5x faster downloads
- Pause, resume, and cancel per download
- Resume interrupted downloads automatically
- Live progress and speed display
- Auto fallback to single-stream mode if server does not support ranges

## Quick Start

```bash
git clone https://github.com/Anarghya1610/Go-Download-Manager.git
cd Go-Download-Manager
./GDM
```

In the app:

1. Paste download URL
2. Choose destination folder
3. Click Add Download
4. Manage with Pause / Resume / Cancel

## Screenshots

<img width="910" height="687" alt="img1" src="https://github.com/user-attachments/assets/d5f55e4e-c5a9-45b7-847a-d2236b3df716" />
<img width="910" height="684" alt="img2" src="https://github.com/user-attachments/assets/35d5f97e-9e04-4dfa-874c-0dfdc3005e81" />

## Advanced (Optional)

- Environment variable: `GDM_WORKERS`
- Use this only if you want to manually control worker count

PowerShell example:

```powershell
$env:GDM_WORKERS=12
./GDM
```

## Resume Data

- `tasks.json`: saved task list
- `<output>.meta`: chunk progress for resume

`.meta` is removed after a successful download.

## Troubleshooting

- Download is not parallel: server likely does not support range requests
- Too many retries: lower `GDM_WORKERS` or let the program decide worker count
