package tracker

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// WindowInfo contient les informations de la fenêtre active
type WindowInfo struct {
	AppName     string
	WindowTitle string
	ProcessPath string
	Timestamp   time.Time
}

// GetActiveWindow retourne les informations de la fenêtre active selon l'OS
func GetActiveWindow() (*WindowInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return getActiveWindowMac()
	case "windows":
		return getActiveWindowWindows()
	case "linux":
		return getActiveWindowLinux()
	default:
		return nil, fmt.Errorf("OS non supporté: %s", runtime.GOOS)
	}
}

// getActiveWindowMac récupère la fenêtre active sur macOS
func getActiveWindowMac() (*WindowInfo, error) {
	// AppleScript pour récupérer l'app active et le titre de la fenêtre
	script := `tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	set frontWindow to ""
	try
		tell process frontApp
			set frontWindow to name of front window
		end tell
	end try
	return frontApp & "|" & frontWindow
end tell`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		// Essayer une approche alternative plus simple
		cmd2 := exec.Command("osascript", "-e", "tell application \"System Events\" to name of first application process whose frontmost is true")
		output2, err2 := cmd2.Output()
		if err2 != nil {
			return nil, fmt.Errorf("erreur osascript: %w", err2)
		}
		appName := strings.TrimSpace(string(output2))
		processPath, _ := getProcessPath(appName)
		return &WindowInfo{
			AppName:     appName,
			WindowTitle: "",
			ProcessPath: processPath,
			Timestamp:   time.Now(),
		}, nil
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	appName := parts[0]
	windowTitle := ""
	if len(parts) > 1 {
		windowTitle = parts[1]
	}

	// Récupérer le chemin du processus
	processPath, _ := getProcessPath(appName)

	return &WindowInfo{
		AppName:     appName,
		WindowTitle: windowTitle,
		ProcessPath: processPath,
		Timestamp:   time.Now(),
	}, nil
}

// getActiveWindowWindows récupère la fenêtre active sur Windows
func getActiveWindowWindows() (*WindowInfo, error) {
	// PowerShell script pour récupérer la fenêtre active
	script := `
		Add-Type @"
			using System;
			using System.Runtime.InteropServices;
			using System.Text;
			public class Win32 {
				[DllImport("user32.dll")]
				public static extern IntPtr GetForegroundWindow();
				[DllImport("user32.dll")]
				public static extern int GetWindowText(IntPtr hWnd, StringBuilder text, int count);
				[DllImport("user32.dll")]
				public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);
			}
"@
		$hwnd = [Win32]::GetForegroundWindow()
		$text = New-Object System.Text.StringBuilder 256
		[Win32]::GetWindowText($hwnd, $text, 256) | Out-Null
		$processId = 0
		[Win32]::GetWindowThreadProcessId($hwnd, [ref]$processId) | Out-Null
		$process = Get-Process -Id $processId -ErrorAction SilentlyContinue
		if ($process) {
			Write-Output "$($process.ProcessName)|$($text.ToString())|$($process.Path)"
		}
	`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erreur PowerShell: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("format de sortie invalide")
	}

	processPath := ""
	if len(parts) > 2 {
		processPath = parts[2]
	}

	return &WindowInfo{
		AppName:     parts[0],
		WindowTitle: parts[1],
		ProcessPath: processPath,
		Timestamp:   time.Now(),
	}, nil
}

// getActiveWindowLinux récupère la fenêtre active sur Linux (X11)
func getActiveWindowLinux() (*WindowInfo, error) {
	// Utiliser xdotool pour X11
	cmd := exec.Command("xdotool", "getactivewindow", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erreur xdotool (installer xdotool): %w", err)
	}

	windowTitle := strings.TrimSpace(string(output))

	// Récupérer le PID de la fenêtre active
	cmd = exec.Command("xdotool", "getactivewindow", "getwindowpid")
	pidOutput, err := cmd.Output()
	if err != nil {
		return &WindowInfo{
			AppName:     "Unknown",
			WindowTitle: windowTitle,
			Timestamp:   time.Now(),
		}, nil
	}

	var pid int32
	fmt.Sscanf(string(pidOutput), "%d", &pid)

	proc, err := process.NewProcess(pid)
	if err != nil {
		return &WindowInfo{
			AppName:     "Unknown",
			WindowTitle: windowTitle,
			Timestamp:   time.Now(),
		}, nil
	}

	appName, _ := proc.Name()
	processPath, _ := proc.Exe()

	return &WindowInfo{
		AppName:     appName,
		WindowTitle: windowTitle,
		ProcessPath: processPath,
		Timestamp:   time.Now(),
	}, nil
}

// getProcessPath tente de récupérer le chemin du processus par son nom
func getProcessPath(appName string) (string, error) {
	processes, err := process.Processes()
	if err != nil {
		return "", err
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.EqualFold(name, appName) || strings.Contains(strings.ToLower(name), strings.ToLower(appName)) {
			path, err := p.Exe()
			if err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("processus non trouvé: %s", appName)
}

// GetEnrichedName extrait un nom contextualisé depuis le titre de la fenêtre
func (w *WindowInfo) GetEnrichedName() string {
	// Navigateurs
	if isBrowser(w.AppName) {
		site := extractWebsiteName(w.WindowTitle)
		if site != "" {
			return site
		}
	}

	// Electron/VSCode
	if isElectronApp(w.AppName) {
		project := extractProjectName(w.WindowTitle)
		if project != "" {
			return project
		}
	}

	// Fallback : garder le nom de l'app
	return w.AppName
}

func isBrowser(appName string) bool {
	browsers := []string{"Brave Browser", "Google Chrome", "Safari",
		"Firefox", "Microsoft Edge", "Arc", "Zen Browser", "Zen",
		"Opera", "Vivaldi", "Chromium", "Waterfox", "LibreWolf"}
	for _, b := range browsers {
		if strings.Contains(appName, b) {
			return true
		}
	}
	return false
}

func extractWebsiteName(title string) string {
	if title == "" {
		return ""
	}

	// Enlever les suffixes navigateur courants
	suffixes := []string{" – Brave", " - Brave", " – Chrome", " - Chrome",
		" – Safari", " - Safari", " – Firefox", " - Firefox",
		" – Edge", " - Edge", " – Arc", " - Arc",
		" – Zen", " - Zen", " – Opera", " - Opera",
		" – Vivaldi", " - Vivaldi", " – Chromium", " - Chromium"}

	for _, suffix := range suffixes {
		if idx := strings.Index(title, suffix); idx > 0 {
			title = title[:idx]
		}
	}

	title = strings.TrimSpace(title)
	titleLower := strings.ToLower(title)

	// Détecter les sites populaires avec patterns spécifiques

	// X (Twitter) - plusieurs patterns possibles
	// "Username sur X : "tweet"" / X" → "X"
	// "Accueil / X" → "X"
	// "(123) X" → "X"
	if strings.Contains(titleLower, " sur x :") ||
		strings.Contains(titleLower, " sur x ") ||
		strings.Contains(titleLower, "accueil / x") ||
		strings.HasSuffix(titleLower, " / x") ||
		strings.Contains(titleLower, ") x") {
		return "X"
	}

	// YouTube - "Video Title - YouTube" → "YouTube"
	if strings.Contains(titleLower, "youtube") {
		return "YouTube"
	}

	// Twitch - "Username - Twitch" → "Twitch"
	if strings.Contains(titleLower, "twitch") {
		return "Twitch"
	}

	// TikTok
	if strings.Contains(titleLower, "tiktok") {
		return "TikTok"
	}

	// Gmail - plusieurs patterns
	if strings.Contains(titleLower, "gmail") || strings.Contains(titleLower, "inbox") {
		return "Gmail"
	}

	// GitHub
	if strings.Contains(titleLower, "github") {
		return "GitHub"
	}

	// LinkedIn
	if strings.Contains(titleLower, "linkedin") {
		return "LinkedIn"
	}

	// Reddit
	if strings.Contains(titleLower, "reddit") {
		return "Reddit"
	}

	// Instagram
	if strings.Contains(titleLower, "instagram") {
		return "Instagram"
	}

	// Facebook
	if strings.Contains(titleLower, "facebook") {
		return "Facebook"
	}

	// Discord
	if strings.Contains(titleLower, "discord") {
		return "Discord"
	}

	// Slack
	if strings.Contains(titleLower, "slack") {
		return "Slack"
	}

	// Notion
	if strings.Contains(titleLower, "notion") {
		return "Notion"
	}

	// Google Drive / Docs / Sheets
	if strings.Contains(titleLower, "google drive") ||
		strings.Contains(titleLower, "google docs") ||
		strings.Contains(titleLower, "google sheets") {
		return "Google Drive"
	}

	// Stack Overflow
	if strings.Contains(titleLower, "stack overflow") {
		return "Stack Overflow"
	}

	// ChatGPT
	if strings.Contains(titleLower, "chatgpt") {
		return "ChatGPT"
	}

	// Claude
	if strings.Contains(titleLower, "claude") {
		return "Claude"
	}

	// Netflix
	if strings.Contains(titleLower, "netflix") {
		return "Netflix"
	}

	// Spotify
	if strings.Contains(titleLower, "spotify") {
		return "Spotify"
	}

	// Tous les autres sites non reconnus → "Autres"
	// On retourne "Autres" au lieu d'essayer d'extraire un nom personnalisé
	return "Autres"
}

func isElectronApp(appName string) bool {
	electronApps := []string{"Electron", "Code", "Visual Studio Code",
		"Cursor", "VSCodium"}
	for _, app := range electronApps {
		if strings.Contains(appName, app) {
			return true
		}
	}
	return false
}

func extractProjectName(title string) string {
	if title == "" {
		return ""
	}

	// Patterns VSCode/Cursor :
	// "file.md — TrackMyTime — Perso" → "TrackMyTime" (3 parts)
	// "TrackMyTime — Perso" → "TrackMyTime" (2 parts, ignore "Perso")
	// "file.py — my-project" → "my-project"

	parts := strings.Split(title, " — ")

	// Si 3+ parties : prendre la partie du milieu (parts[1])
	// Exemple: ["file.md", "TrackMyTime", "Perso"] → "TrackMyTime"
	if len(parts) >= 3 {
		projectName := strings.TrimSpace(parts[1])
		if projectName != "" {
			return projectName
		}
	}

	// Si 2 parties : prendre la dernière sauf si générique
	// Exemple: ["TrackMyTime", "Perso"] → "TrackMyTime"
	if len(parts) == 2 {
		// Vérifier la dernière partie d'abord
		lastPart := strings.TrimSpace(parts[1])
		genericNames := []string{"Perso", "Workspace", "Visual Studio Code"}
		isLastGeneric := false
		for _, generic := range genericNames {
			if strings.EqualFold(lastPart, generic) {
				isLastGeneric = true
				break
			}
		}

		// Si dernière partie est générique, prendre la première
		if isLastGeneric {
			projectName := strings.TrimSpace(parts[0])
			if projectName != "" {
				return projectName
			}
		} else {
			// Sinon prendre la dernière
			if lastPart != "" {
				return lastPart
			}
		}
	}

	// Fallback : chercher entre []
	// "[TrackMyTime] file.md"
	if strings.HasPrefix(title, "[") {
		end := strings.Index(title, "]")
		if end > 1 {
			return title[1:end]
		}
	}

	return ""
}
