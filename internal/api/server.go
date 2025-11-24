package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"trackmytime/internal/storage"
	"trackmytime/internal/tracker"
)

// Server représente le serveur HTTP de l'API
type Server struct {
	db      *storage.DB
	port    string
	tracker *ActivityTracker
}

// ActivityTracker contient l'état actuel du tracking
type ActivityTracker struct {
	CurrentWindow *tracker.WindowInfo
	StartTime     time.Time
}

// NewServer crée un nouveau serveur API
func NewServer(db *storage.DB, port string) *Server {
	return &Server{
		db:      db,
		port:    port,
		tracker: &ActivityTracker{},
	}
}

// SetCurrentActivity met à jour l'activité courante
func (s *Server) SetCurrentActivity(window *tracker.WindowInfo, startTime time.Time) {
	s.tracker.CurrentWindow = window
	s.tracker.StartTime = startTime
}

// Start démarre le serveur HTTP
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Web UI routes
	mux.HandleFunc("/", s.handleWebUI)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// API routes
	mux.HandleFunc("/stats/today", s.handleStatsToday)
	mux.HandleFunc("/stats/week", s.handleStatsWeek)
	mux.HandleFunc("/stats/month", s.handleStatsMonth)
	mux.HandleFunc("/stats/custom", s.handleStatsCustom)
	mux.HandleFunc("/export/csv", s.handleExportCSV)
	mux.HandleFunc("/activity/current", s.handleCurrentActivity)
	mux.HandleFunc("/browser/event", s.handleBrowserEvent)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/export/aggregated", s.handleExportAggregated)
	mux.HandleFunc("/api/stats/hourly", s.handleStatsHourly)
	mux.HandleFunc("/api/stats/grouped", s.handleStatsGrouped)

	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("API HTTP démarrée sur http://localhost%s", addr)
	log.Printf("Dashboard web disponible sur http://localhost%s/", addr)

	return http.ListenAndServe(addr, mux)
}

// handleHealth vérifie l'état du serveur
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleStatsToday retourne les statistiques du jour
func (s *Server) handleStatsToday(w http.ResponseWriter, r *http.Request) {
	activities, err := s.db.GetTodayActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	stats, err := s.db.GetStatsByApp(startOfDay, endOfDay)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"date":                 now.Format("2006-01-02"),
		"total_activities":     len(activities),
		"stats_by_app":         stats,
		"total_active_seconds": sumStats(stats),
		"total_active_hours":   float64(sumStats(stats)) / 3600.0,
		"total_idle_seconds":   calculateIdleTime(activities),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatsWeek retourne les statistiques de la semaine
func (s *Server) handleStatsWeek(w http.ResponseWriter, r *http.Request) {
	activities, err := s.db.GetWeekActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	stats, err := s.db.GetStatsByApp(startOfWeek, endOfWeek)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"week_start":           startOfWeek.Format("2006-01-02"),
		"week_end":             endOfWeek.Format("2006-01-02"),
		"total_activities":     len(activities),
		"stats_by_app":         stats,
		"total_active_seconds": sumStats(stats),
		"total_active_hours":   float64(sumStats(stats)) / 3600.0,
		"total_idle_seconds":   calculateIdleTime(activities),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatsMonth retourne les statistiques du mois
func (s *Server) handleStatsMonth(w http.ResponseWriter, r *http.Request) {
	activities, err := s.db.GetMonthActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	stats, err := s.db.GetStatsByApp(startOfMonth, endOfMonth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"month_start":          startOfMonth.Format("2006-01-02"),
		"month_end":            endOfMonth.Format("2006-01-02"),
		"total_activities":     len(activities),
		"stats_by_app":         stats,
		"total_active_seconds": sumStats(stats),
		"total_active_hours":   float64(sumStats(stats)) / 3600.0,
		"total_idle_seconds":   calculateIdleTime(activities),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatsCustom retourne les statistiques pour une période personnalisée
func (s *Server) handleStatsCustom(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start, end, err := s.parseCustomPeriod(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer les activités
	activities, err := s.db.GetActivitiesByDateRange(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats, err := s.db.GetStatsByApp(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"start_date":           startStr,
		"end_date":             endStr,
		"total_activities":     len(activities),
		"stats_by_app":         stats,
		"total_active_seconds": sumStats(stats),
		"total_active_hours":   float64(sumStats(stats)) / 3600.0,
		"total_idle_seconds":   calculateIdleTime(activities),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleExportCSV génère un export CSV des activités
func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "today"
	}

	var activities []storage.Activity
	var err error

	switch period {
	case "today":
		activities, err = s.db.GetTodayActivities()
	case "week":
		activities, err = s.db.GetWeekActivities()
	case "month":
		activities, err = s.db.GetMonthActivities()
	case "custom":
		var start, end time.Time
		start, end, err = s.parseCustomPeriod(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		activities, err = s.db.GetActivitiesByDateRange(start, end)
	default:
		http.Error(w, "period invalide (today, week, month, custom)", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=activities_%s.csv", period))

	// Écrire le CSV directement dans la réponse
	w.Write([]byte("ID,App Name,Window Title,Process Path,Start Time,End Time,Duration (seconds),Is Idle\n"))
	for _, activity := range activities {
		line := fmt.Sprintf("%d,%s,%s,%s,%s,%s,%d,%t\n",
			activity.ID,
			activity.AppName,
			activity.WindowTitle,
			activity.ProcessPath,
			activity.StartTime.Format(time.RFC3339),
			activity.EndTime.Format(time.RFC3339),
			activity.DurationSecs,
			activity.IsIdle,
		)
		w.Write([]byte(line))
	}
}

// handleCurrentActivity retourne l'activité en cours
func (s *Server) handleCurrentActivity(w http.ResponseWriter, r *http.Request) {
	if s.tracker.CurrentWindow == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "no activity",
		})
		return
	}

	duration := time.Since(s.tracker.StartTime)

	response := map[string]any{
		"app_name":         s.tracker.CurrentWindow.AppName,
		"window_title":     s.tracker.CurrentWindow.WindowTitle,
		"process_path":     s.tracker.CurrentWindow.ProcessPath,
		"start_time":       s.tracker.StartTime.Format(time.RFC3339),
		"current_duration": int64(duration.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// BrowserEvent représente un événement du navigateur
type BrowserEvent struct {
	URL         string `json:"url"`
	TabTitle    string `json:"tab_title"`
	BrowserName string `json:"browser_name"`
	Timestamp   string `json:"timestamp"`
}

// handleBrowserEvent gère les événements du navigateur (pour future extension)
func (s *Server) handleBrowserEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var event BrowserEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Pour l'instant, on log juste l'événement
	// Dans le futur, on pourra le stocker dans browser_events
	log.Printf("Browser event received: %s - %s (%s)", event.BrowserName, event.TabTitle, event.URL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "received",
		"message": "Browser event logged successfully",
	})
}

// sumStats calcule la somme totale des statistiques
func sumStats(stats map[string]int64) int64 {
	var total int64
	for _, duration := range stats {
		total += duration
	}
	return total
}

// handleExportAggregated exporte les stats agrégées avec format HH:MM:SS
func (s *Server) handleExportAggregated(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "today"
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	start, end, err := s.getPeriodBounds(period, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stats, err := s.db.GetStatsByApp(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Type pour le tri
	type appStat struct {
		name    string
		seconds int64
	}

	// Convertir map en slice pour le tri
	var sortedStats []appStat
	var totalSeconds int64
	for appName, seconds := range stats {
		sortedStats = append(sortedStats, appStat{appName, seconds})
		totalSeconds += seconds
	}

	// Tri par durée décroissante avec sort.Slice
	sort.Slice(sortedStats, func(i, j int) bool {
		return sortedStats[i].seconds > sortedStats[j].seconds
	})

	if format == "json" {
		// Format JSON avec durées HH:MM:SS
		type AggregatedStat struct {
			AppName      string  `json:"app_name"`
			Duration     string  `json:"duration"`
			TotalHours   float64 `json:"total_hours"`
			TotalSeconds int64   `json:"total_seconds"`
		}

		var statsList []AggregatedStat

		for _, stat := range sortedStats {
			hours := stat.seconds / 3600
			minutes := (stat.seconds % 3600) / 60
			secs := stat.seconds % 60
			duration := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)

			statsList = append(statsList, AggregatedStat{
				AppName:      stat.name,
				Duration:     duration,
				TotalHours:   float64(stat.seconds) / 3600.0,
				TotalSeconds: stat.seconds,
			})
		}

		// Calculer le total
		totalHours := totalSeconds / 3600
		totalMinutes := (totalSeconds % 3600) / 60
		totalSecs := totalSeconds % 60
		totalDuration := fmt.Sprintf("%02d:%02d:%02d", totalHours, totalMinutes, totalSecs)

		result := map[string]any{
			"period":       period,
			"applications": statsList,
			"total": map[string]any{
				"app_name":      "TOTAL",
				"duration":      totalDuration,
				"total_hours":   float64(totalSeconds) / 3600.0,
				"total_seconds": totalSeconds,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	} else {
		// Format CSV
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=aggregated_%s.csv", period))

		// Header CSV
		w.Write([]byte("Application,Duration (HH:MM:SS),Total Hours,Total Seconds\n"))

		// Écrire les données
		for _, stat := range sortedStats {
			hours := stat.seconds / 3600
			minutes := (stat.seconds % 3600) / 60
			secs := stat.seconds % 60
			duration := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
			totalHours := float64(stat.seconds) / 3600.0

			line := fmt.Sprintf("%s,%s,%.2f,%d\n", stat.name, duration, totalHours, stat.seconds)
			w.Write([]byte(line))
		}

		// Ligne de total
		if len(sortedStats) > 0 {
			totalHours := totalSeconds / 3600
			totalMinutes := (totalSeconds % 3600) / 60
			totalSecs := totalSeconds % 60
			totalDuration := fmt.Sprintf("%02d:%02d:%02d", totalHours, totalMinutes, totalSecs)
			totalHoursFloat := float64(totalSeconds) / 3600.0

			totalLine := fmt.Sprintf("TOTAL,%s,%.2f,%d\n", totalDuration, totalHoursFloat, totalSeconds)
			w.Write([]byte(totalLine))
		}
	}
}

// handleWebUI sert l'interface web
func (s *Server) handleWebUI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

// handleStatsHourly retourne les statistiques par heure
func (s *Server) handleStatsHourly(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "today"
	}

	start, end, err := s.getPeriodBounds(period, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Pour today, utiliser les stats horaires, sinon les stats journalières
	var timelineData []int64
	var labels []string

	// Utiliser les stats horaires pour une seule journée, sinon par jour
	durationDays := int(end.Sub(start).Hours() / 24)

	if period == "today" || (period == "custom" && durationDays == 1) {
		timelineData, err = s.db.GetHourlyStats(start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Labels pour les heures
		labels = make([]string, 24)
		for i := range labels {
			labels[i] = fmt.Sprintf("%02d:00", i)
		}
	} else {
		timelineData, err = s.db.GetDailyStats(start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Labels pour les jours
		labels = make([]string, len(timelineData))
		for i := range labels {
			day := start.AddDate(0, 0, i)
			if period == "week" {
				// Pour la semaine, afficher Lun, Mar, Mer, etc.
				weekdays := []string{"Dim", "Lun", "Mar", "Mer", "Jeu", "Ven", "Sam"}
				labels[i] = weekdays[day.Weekday()]
			} else {
				// Pour le mois, afficher le numéro du jour
				labels[i] = fmt.Sprintf("%d", day.Day())
			}
		}
	}

	response := map[string]any{
		"period":        period,
		"timeline_data": timelineData,
		"labels":        labels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatsGrouped retourne les statistiques groupées par app et enriched_name
func (s *Server) handleStatsGrouped(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "today"
	}

	start, end, err := s.getPeriodBounds(period, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	grouped, err := s.db.GetGroupedStats(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Formatter la réponse
	type EnrichedApp struct {
		Name     string `json:"name"`
		Duration int64  `json:"duration"`
	}

	type AppGroup struct {
		AppName      string        `json:"app_name"`
		TotalSeconds int64         `json:"total_seconds"`
		Children     []EnrichedApp `json:"children"`
	}

	var result []AppGroup

	for appName, children := range grouped {
		var totalSeconds int64
		var childrenList []EnrichedApp

		for enrichedName, duration := range children {
			totalSeconds += duration
			childrenList = append(childrenList, EnrichedApp{
				Name:     enrichedName,
				Duration: duration,
			})
		}

		// Trier children par durée décroissante avec sort.Slice
		sort.Slice(childrenList, func(i, j int) bool {
			return childrenList[i].Duration > childrenList[j].Duration
		})

		result = append(result, AppGroup{
			AppName:      appName,
			TotalSeconds: totalSeconds,
			Children:     childrenList,
		})
	}

	// Trier result par durée totale décroissante avec sort.Slice
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalSeconds > result[j].TotalSeconds
	})

	response := map[string]any{
		"period": period,
		"groups": result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
