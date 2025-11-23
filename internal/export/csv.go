package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"trackmytime/internal/storage"
)

// ExportCSV exporte les activités au format CSV
func ExportCSV(activities []storage.Activity, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"ID", "App Name", "Window Title", "Process Path", "Start Time", "End Time", "Duration (seconds)", "Is Idle"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("erreur écriture header: %w", err)
	}

	// Données
	for _, activity := range activities {
		record := []string{
			fmt.Sprintf("%d", activity.ID),
			activity.AppName,
			activity.WindowTitle,
			activity.ProcessPath,
			activity.StartTime.Format(time.RFC3339),
			activity.EndTime.Format(time.RFC3339),
			fmt.Sprintf("%d", activity.DurationSecs),
			fmt.Sprintf("%t", activity.IsIdle),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("erreur écriture ligne: %w", err)
		}
	}

	return nil
}

// ExportJSON exporte les activités au format JSON
func ExportJSON(activities []storage.Activity, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(activities); err != nil {
		return fmt.Errorf("erreur encodage JSON: %w", err)
	}

	return nil
}

// StatsByApp représente les statistiques par application
type StatsByApp struct {
	AppName      string  `json:"app_name"`
	TotalSeconds int64   `json:"total_seconds"`
	TotalHours   float64 `json:"total_hours"`
}

// ExportStatsByAppCSV exporte les statistiques par application au format CSV
func ExportStatsByAppCSV(stats map[string]int64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"App Name", "Total Seconds", "Total Hours"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("erreur écriture header: %w", err)
	}

	// Données
	for appName, totalSeconds := range stats {
		totalHours := float64(totalSeconds) / 3600.0
		record := []string{
			appName,
			fmt.Sprintf("%d", totalSeconds),
			fmt.Sprintf("%.2f", totalHours),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("erreur écriture ligne: %w", err)
		}
	}

	return nil
}

// ExportStatsByAppJSON exporte les statistiques par application au format JSON
func ExportStatsByAppJSON(stats map[string]int64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	// Convertir en slice pour un meilleur format JSON
	var statsList []StatsByApp
	for appName, totalSeconds := range stats {
		statsList = append(statsList, StatsByApp{
			AppName:      appName,
			TotalSeconds: totalSeconds,
			TotalHours:   float64(totalSeconds) / 3600.0,
		})
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(statsList); err != nil {
		return fmt.Errorf("erreur encodage JSON: %w", err)
	}

	return nil
}

// DailyStats représente les statistiques journalières
type DailyStats struct {
	Date         string  `json:"date"`
	TotalSeconds int64   `json:"total_seconds"`
	TotalHours   float64 `json:"total_hours"`
}

// ExportDailyStatsCSV exporte les statistiques journalières au format CSV
func ExportDailyStatsCSV(activities []storage.Activity, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Regrouper par jour
	dailyStats := make(map[string]int64)
	for _, activity := range activities {
		if activity.IsIdle {
			continue
		}
		date := activity.StartTime.Format("2006-01-02")
		dailyStats[date] += activity.DurationSecs
	}

	// Header
	header := []string{"Date", "Total Seconds", "Total Hours"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("erreur écriture header: %w", err)
	}

	// Données
	for date, totalSeconds := range dailyStats {
		totalHours := float64(totalSeconds) / 3600.0
		record := []string{
			date,
			fmt.Sprintf("%d", totalSeconds),
			fmt.Sprintf("%.2f", totalHours),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("erreur écriture ligne: %w", err)
		}
	}

	return nil
}

// formatDuration convertit des secondes en format HH:MM:SS lisible
func formatDuration(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// AggregatedStat représente les statistiques agrégées par application
type AggregatedStat struct {
	AppName      string `json:"app_name"`
	TotalSeconds int64  `json:"total_seconds"`
	Duration     string `json:"duration"` // Format HH:MM:SS
	TotalHours   float64 `json:"total_hours"`
}

// ExportAggregatedCSV exporte les stats agrégées par app avec format HH:MM:SS
func ExportAggregatedCSV(stats map[string]int64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"Application", "Duration (HH:MM:SS)", "Total Hours", "Total Seconds"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("erreur écriture header: %w", err)
	}

	// Données triées par durée décroissante
	type appStat struct {
		name    string
		seconds int64
	}
	var sortedStats []appStat
	for appName, totalSeconds := range stats {
		sortedStats = append(sortedStats, appStat{appName, totalSeconds})
	}
	
	// Tri simple par durée décroissante
	for i := 0; i < len(sortedStats); i++ {
		for j := i + 1; j < len(sortedStats); j++ {
			if sortedStats[j].seconds > sortedStats[i].seconds {
				sortedStats[i], sortedStats[j] = sortedStats[j], sortedStats[i]
			}
		}
	}

	// Écrire les données
	var totalSeconds int64
	for _, stat := range sortedStats {
		totalHours := float64(stat.seconds) / 3600.0
		record := []string{
			stat.name,
			formatDuration(stat.seconds),
			fmt.Sprintf("%.2f", totalHours),
			fmt.Sprintf("%d", stat.seconds),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("erreur écriture ligne: %w", err)
		}
		totalSeconds += stat.seconds
	}

	// Ajouter une ligne de total
	if len(sortedStats) > 0 {
		totalHours := float64(totalSeconds) / 3600.0
		totalRecord := []string{
			"TOTAL",
			formatDuration(totalSeconds),
			fmt.Sprintf("%.2f", totalHours),
			fmt.Sprintf("%d", totalSeconds),
		}
		if err := writer.Write(totalRecord); err != nil {
			return fmt.Errorf("erreur écriture ligne totale: %w", err)
		}
	}

	return nil
}

// ExportAggregatedJSON exporte les stats agrégées au format JSON
func ExportAggregatedJSON(stats map[string]int64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer file.Close()

	// Convertir en slice
	var statsList []AggregatedStat
	var totalSeconds int64
	
	for appName, seconds := range stats {
		statsList = append(statsList, AggregatedStat{
			AppName:      appName,
			TotalSeconds: seconds,
			Duration:     formatDuration(seconds),
			TotalHours:   float64(seconds) / 3600.0,
		})
		totalSeconds += seconds
	}

	// Trier par durée décroissante
	for i := 0; i < len(statsList); i++ {
		for j := i + 1; j < len(statsList); j++ {
			if statsList[j].TotalSeconds > statsList[i].TotalSeconds {
				statsList[i], statsList[j] = statsList[j], statsList[i]
			}
		}
	}

	// Ajouter le total
	result := map[string]interface{}{
		"applications": statsList,
		"total": AggregatedStat{
			AppName:      "TOTAL",
			TotalSeconds: totalSeconds,
			Duration:     formatDuration(totalSeconds),
			TotalHours:   float64(totalSeconds) / 3600.0,
		},
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("erreur encodage JSON: %w", err)
	}

	return nil
}
