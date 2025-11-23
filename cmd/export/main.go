package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"trackmytime/config"
	"trackmytime/internal/export"
	"trackmytime/internal/storage"
)

func main() {
	// Définir les flags
	period := flag.String("period", "today", "Période à exporter (today, week)")
	format := flag.String("format", "csv", "Format d'export (csv, json)")
	output := flag.String("output", "", "Fichier de sortie (par défaut: trackmytime_PERIOD_YYYYMMDD.FORMAT)")
	aggregated := flag.Bool("aggregated", false, "Export agrégé avec temps combiné par application")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExemples:\n")
		fmt.Fprintf(os.Stderr, "  # Export agrégé du jour en CSV\n")
		fmt.Fprintf(os.Stderr, "  %s -aggregated\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Export agrégé de la semaine en JSON\n")
		fmt.Fprintf(os.Stderr, "  %s -aggregated -period week -format json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Export détaillé avec nom de fichier personnalisé\n")
		fmt.Fprintf(os.Stderr, "  %s -output mon_export.csv\n\n", os.Args[0])
	}
	
	flag.Parse()

	// Charger la configuration
	cfg := config.DefaultConfig()

	// Connexion à la base de données
	db, err := storage.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Erreur connexion DB: %v", err)
	}
	defer db.Close()

	// Déterminer la période
	var start, end time.Time
	now := time.Now()

	switch *period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour)
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		end = start.Add(7 * 24 * time.Hour)
	default:
		log.Fatalf("❌ Période invalide: %s (utilisez 'today' ou 'week')", *period)
	}

	// Nom de fichier par défaut
	if *output == "" {
		dateStr := now.Format("20060102")
		if *aggregated {
			*output = fmt.Sprintf("trackmytime_aggregated_%s_%s.%s", *period, dateStr, *format)
		} else {
			*output = fmt.Sprintf("trackmytime_detailed_%s_%s.%s", *period, dateStr, *format)
		}
	}

	// Export agrégé ou détaillé
	if *aggregated {
		// Export agrégé avec temps combiné
		stats, err := db.GetStatsByApp(start, end)
		if err != nil {
			log.Fatalf("❌ Erreur récupération stats: %v", err)
		}

		if len(stats) == 0 {
			log.Printf("⚠️  Aucune donnée à exporter pour la période '%s'", *period)
			return
		}

		if *format == "json" {
			err = export.ExportAggregatedJSON(stats, *output)
		} else {
			err = export.ExportAggregatedCSV(stats, *output)
		}

		if err != nil {
			log.Fatalf("❌ Erreur export agrégé: %v", err)
		}

		log.Printf("✅ Export agrégé créé: %s", *output)
		log.Printf("📊 %d applications trackées", len(stats))
		
		// Afficher un aperçu
		var total int64
		for _, seconds := range stats {
			total += seconds
		}
		hours := total / 3600
		minutes := (total % 3600) / 60
		secs := total % 60
		log.Printf("⏱️  Temps total: %02d:%02d:%02d (%.2f heures)", hours, minutes, secs, float64(total)/3600.0)

	} else {
		// Export détaillé
		activities, err := db.GetActivitiesByDateRange(start, end)
		if err != nil {
			log.Fatalf("❌ Erreur récupération activités: %v", err)
		}

		if len(activities) == 0 {
			log.Printf("⚠️  Aucune activité à exporter pour la période '%s'", *period)
			return
		}

		if *format == "json" {
			err = export.ExportJSON(activities, *output)
		} else {
			err = export.ExportCSV(activities, *output)
		}

		if err != nil {
			log.Fatalf("❌ Erreur export détaillé: %v", err)
		}

		log.Printf("✅ Export détaillé créé: %s", *output)
		log.Printf("📊 %d activités exportées", len(activities))
	}
}
