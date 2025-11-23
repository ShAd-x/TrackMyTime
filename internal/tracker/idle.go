package tracker

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// IdleDetector détecte l'inactivité de l'utilisateur
type IdleDetector struct {
	threshold time.Duration
}

// NewIdleDetector crée un nouveau détecteur d'inactivité
func NewIdleDetector(threshold time.Duration) *IdleDetector {
	return &IdleDetector{
		threshold: threshold,
	}
}

// IsIdle retourne true si l'utilisateur est inactif
func (id *IdleDetector) IsIdle() (bool, error) {
	idleTime, err := id.GetIdleTime()
	if err != nil {
		return false, err
	}

	return idleTime >= id.threshold, nil
}

// GetIdleTime retourne le temps d'inactivité en secondes
func (id *IdleDetector) GetIdleTime() (time.Duration, error) {
	switch runtime.GOOS {
	case "darwin":
		return getIdleTimeMac()
	case "windows":
		return getIdleTimeWindows()
	case "linux":
		return getIdleTimeLinux()
	default:
		return 0, fmt.Errorf("OS non supporté: %s", runtime.GOOS)
	}
}

// getIdleTimeMac récupère le temps d'inactivité sur macOS
func getIdleTimeMac() (time.Duration, error) {
	cmd := exec.Command("ioreg", "-c", "IOHIDSystem")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("erreur ioreg: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "HIDIdleTime") {
			// Format: "HIDIdleTime" = 12345678901
			parts := strings.Split(line, "=")
			if len(parts) < 2 {
				continue
			}

			idleStr := strings.TrimSpace(parts[1])
			idleNano, err := strconv.ParseInt(idleStr, 10, 64)
			if err != nil {
				continue
			}

			// HIDIdleTime est en nanosecondes
			return time.Duration(idleNano), nil
		}
	}

	return 0, fmt.Errorf("impossible de trouver HIDIdleTime")
}

// getIdleTimeWindows récupère le temps d'inactivité sur Windows
func getIdleTimeWindows() (time.Duration, error) {
	script := `
		Add-Type @"
			using System;
			using System.Runtime.InteropServices;
			public struct LASTINPUTINFO {
				public uint cbSize;
				public uint dwTime;
			}
			public class Win32 {
				[DllImport("user32.dll")]
				public static extern bool GetLastInputInfo(ref LASTINPUTINFO plii);
				[DllImport("kernel32.dll")]
				public static extern uint GetTickCount();
			}
"@
		$lastInputInfo = New-Object LASTINPUTINFO
		$lastInputInfo.cbSize = [System.Runtime.InteropServices.Marshal]::SizeOf($lastInputInfo)
		[Win32]::GetLastInputInfo([ref]$lastInputInfo) | Out-Null
		$idleTime = ([Win32]::GetTickCount() - $lastInputInfo.dwTime)
		Write-Output $idleTime
	`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("erreur PowerShell: %w", err)
	}

	idleMs, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("erreur de parsing: %w", err)
	}

	return time.Duration(idleMs) * time.Millisecond, nil
}

// getIdleTimeLinux récupère le temps d'inactivité sur Linux (X11)
func getIdleTimeLinux() (time.Duration, error) {
	// Utiliser xprintidle (doit être installé)
	cmd := exec.Command("xprintidle")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("erreur xprintidle (installer xprintidle): %w", err)
	}

	idleMs, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("erreur de parsing: %w", err)
	}

	return time.Duration(idleMs) * time.Millisecond, nil
}
