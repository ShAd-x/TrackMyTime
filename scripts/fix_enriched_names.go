package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Chemin vers la base de données
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	
	dbPath := filepath.Join(homeDir, ".trackmytime", "activities.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	log.Println("🔧 Nettoyage des enriched_name...")
	
	// Récupérer toutes les activités
	rows, err := db.Query("SELECT id, app_name, window_title FROM activities")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	updated := 0
	for rows.Next() {
		var id int64
		var appName, windowTitle string
		
		if err := rows.Scan(&id, &appName, &windowTitle); err != nil {
			log.Printf("Erreur scan: %v", err)
			continue
		}
		
		// Calculer le nouveau enriched_name
		enrichedName := getEnrichedName(appName, windowTitle)
		
		// Mettre à jour
		_, err := db.Exec("UPDATE activities SET enriched_name = ? WHERE id = ?", enrichedName, id)
		if err != nil {
			log.Printf("Erreur update ID %d: %v", id, err)
			continue
		}
		
		updated++
		if updated%100 == 0 {
			log.Printf("✅ %d entrées mises à jour...", updated)
		}
	}
	
	log.Printf("✅ Terminé ! %d entrées nettoyées", updated)
}

func getEnrichedName(appName, windowTitle string) string {
	if isBrowser(appName) {
		site := extractWebsiteName(windowTitle)
		if site != "" {
			return site
		}
	}
	
	if isElectronApp(appName) {
		project := extractProjectName(windowTitle)
		if project != "" {
			return project
		}
	}
	
	return appName
}

func isBrowser(appName string) bool {
	browsers := []string{"Brave Browser", "Google Chrome", "Safari", 
						"Firefox", "Microsoft Edge", "Arc"}
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
	
	// Enlever les suffixes navigateur
	suffixes := []string{" – Brave", " - Brave", " – Chrome", " - Chrome", 
					    " – Safari", " - Safari", " – Firefox", " - Firefox",
					    " – Edge", " - Edge", " – Arc", " - Arc"}
	
	for _, suffix := range suffixes {
		if idx := strings.Index(title, suffix); idx > 0 {
			title = title[:idx]
		}
	}
	
	title = strings.TrimSpace(title)
	titleLower := strings.ToLower(title)
	
	// Sites populaires
	if strings.Contains(titleLower, " sur x :") || 
	   strings.Contains(titleLower, " sur x ") ||
	   strings.Contains(titleLower, "accueil / x") ||
	   strings.HasSuffix(titleLower, " / x") ||
	   strings.Contains(titleLower, ") x") {
		return "X"
	}
	
	if strings.Contains(titleLower, "youtube") {
		return "YouTube"
	}
	
	if strings.Contains(titleLower, "twitch") {
		return "Twitch"
	}
	
	if strings.Contains(titleLower, "tiktok") {
		return "TikTok"
	}
	
	if strings.Contains(titleLower, "gmail") || strings.Contains(titleLower, "inbox") {
		return "Gmail"
	}
	
	if strings.Contains(titleLower, "github") {
		return "GitHub"
	}
	
	if strings.Contains(titleLower, "linkedin") {
		return "LinkedIn"
	}
	
	if strings.Contains(titleLower, "reddit") {
		return "Reddit"
	}
	
	if strings.Contains(titleLower, "instagram") {
		return "Instagram"
	}
	
	if strings.Contains(titleLower, "facebook") {
		return "Facebook"
	}
	
	if strings.Contains(titleLower, "discord") {
		return "Discord"
	}
	
	if strings.Contains(titleLower, "slack") {
		return "Slack"
	}
	
	if strings.Contains(titleLower, "notion") {
		return "Notion"
	}
	
	if strings.Contains(titleLower, "google drive") || 
	   strings.Contains(titleLower, "google docs") ||
	   strings.Contains(titleLower, "google sheets") {
		return "Google Drive"
	}
	
	if strings.Contains(titleLower, "stack overflow") {
		return "Stack Overflow"
	}
	
	if strings.Contains(titleLower, "chatgpt") {
		return "ChatGPT"
	}
	
	if strings.Contains(titleLower, "claude") {
		return "Claude"
	}
	
	if strings.Contains(titleLower, "netflix") {
		return "Netflix"
	}
	
	if strings.Contains(titleLower, "spotify") {
		return "Spotify"
	}
	
	// Tous les autres sites non reconnus → "Autres"
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
	
	parts := strings.Split(title, " — ")
	
	if len(parts) >= 3 {
		projectName := strings.TrimSpace(parts[1])
		if projectName != "" {
			return projectName
		}
	}
	
	if len(parts) == 2 {
		lastPart := strings.TrimSpace(parts[1])
		genericNames := []string{"Perso", "Workspace", "Visual Studio Code"}
		isLastGeneric := false
		for _, generic := range genericNames {
			if strings.EqualFold(lastPart, generic) {
				isLastGeneric = true
				break
			}
		}
		
		if isLastGeneric {
			projectName := strings.TrimSpace(parts[0])
			if projectName != "" {
				return projectName
			}
		} else {
			if lastPart != "" {
				return lastPart
			}
		}
	}
	
	if strings.HasPrefix(title, "[") {
		end := strings.Index(title, "]")
		if end > 1 {
			return title[1:end]
		}
	}
	
	return ""
}
