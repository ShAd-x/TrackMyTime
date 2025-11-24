package storage

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Activity représente une activité trackée
type Activity struct {
	ID           int64
	AppName      string
	EnrichedName string
	WindowTitle  string
	ProcessPath  string
	StartTime    time.Time
	EndTime      time.Time
	DurationSecs int64
	IsIdle       bool
}

// DB gère la connexion à la base de données
type DB struct {
	conn *sql.DB
}

// NewDB crée une nouvelle connexion à la base de données
func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}

	// Exécuter les migrations
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close ferme la connexion à la base de données
func (db *DB) Close() error {
	return db.conn.Close()
}

// InsertActivity insère une nouvelle activité
func (db *DB) InsertActivity(activity *Activity) error {
	query := `
		INSERT INTO activities (app_name, enriched_name, window_title, process_path, start_time, end_time, duration_seconds, is_idle)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		activity.AppName,
		activity.EnrichedName,
		activity.WindowTitle,
		activity.ProcessPath,
		activity.StartTime,
		activity.EndTime,
		activity.DurationSecs,
		activity.IsIdle,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	activity.ID = id
	return nil
}

// GetActivitiesByDateRange retourne les activités dans une plage de dates
func (db *DB) GetActivitiesByDateRange(start, end time.Time) ([]Activity, error) {
	query := `
		SELECT id, app_name, window_title, process_path, start_time, end_time, duration_seconds, is_idle
		FROM activities
		WHERE start_time >= ? AND start_time < ?
		ORDER BY start_time DESC
	`

	rows, err := db.conn.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var a Activity
		err := rows.Scan(
			&a.ID,
			&a.AppName,
			&a.WindowTitle,
			&a.ProcessPath,
			&a.StartTime,
			&a.EndTime,
			&a.DurationSecs,
			&a.IsIdle,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}

	return activities, nil
}

// GetTodayActivities retourne les activités du jour
func (db *DB) GetTodayActivities() ([]Activity, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return db.GetActivitiesByDateRange(startOfDay, endOfDay)
}

// GetWeekActivities retourne les activités de la semaine
func (db *DB) GetWeekActivities() ([]Activity, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Dimanche = 7
	}
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	return db.GetActivitiesByDateRange(startOfWeek, endOfWeek)
}

// GetMonthActivities retourne les activités du mois
func (db *DB) GetMonthActivities() ([]Activity, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	return db.GetActivitiesByDateRange(startOfMonth, endOfMonth)
}

// GetConfig récupère une valeur de configuration
func (db *DB) GetConfig(key string) (string, error) {
	var value string
	query := `SELECT value FROM config WHERE key = ?`
	err := db.conn.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig définit une valeur de configuration
func (db *DB) SetConfig(key, value string) error {
	query := `
		INSERT INTO config (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`
	_, err := db.conn.Exec(query, key, value)
	return err
}

// GetStatsByApp retourne les statistiques groupées par application
func (db *DB) GetStatsByApp(start, end time.Time) (map[string]int64, error) {
	query := `
		SELECT app_name, SUM(duration_seconds) as total_duration
		FROM activities
		WHERE start_time >= ? AND start_time <= ? AND is_idle = 0
		GROUP BY app_name
		ORDER BY total_duration DESC
	`

	rows, err := db.conn.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var appName string
		var duration int64
		if err := rows.Scan(&appName, &duration); err != nil {
			return nil, err
		}
		stats[appName] = duration
	}

	return stats, nil
}

// GetHourlyStats retourne les statistiques par heure pour une période donnée
func (db *DB) GetHourlyStats(start, end time.Time) ([]int64, error) {
	// Initialiser un tableau de 24 heures à zéro
	hourlyStats := make([]int64, 24)

	query := `
		SELECT 
			CAST(strftime('%H', start_time, 'localtime') AS INTEGER) as hour,
			SUM(duration_seconds) as total_duration
		FROM activities
		WHERE start_time >= ? AND start_time < ? AND is_idle = 0
		GROUP BY hour
		ORDER BY hour
	`

	rows, err := db.conn.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hour int
		var duration int64
		if err := rows.Scan(&hour, &duration); err != nil {
			return nil, err
		}
		if hour >= 0 && hour < 24 {
			hourlyStats[hour] = duration
		}
	}

	return hourlyStats, nil
}

// GetDailyStats retourne les statistiques par jour pour une période donnée
func (db *DB) GetDailyStats(start, end time.Time) ([]int64, error) {
	// Calculer le nombre de jours dans la période
	days := int(end.Sub(start).Hours() / 24)
	if days <= 0 {
		days = 1
	}

	dailyStats := make([]int64, days)

	query := `
		SELECT 
			date(start_time, 'localtime') as day,
			SUM(duration_seconds) as total_duration
		FROM activities
		WHERE start_time >= ? AND start_time < ? AND is_idle = 0
		GROUP BY day
		ORDER BY day
	`

	rows, err := db.conn.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dayStr string
		var duration int64
		if err := rows.Scan(&dayStr, &duration); err != nil {
			return nil, err
		}

		// Parser la date et calculer l'index
		dayTime, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			continue
		}

		// Calculer le nombre de jours depuis le début de la période
		dayIndex := int(dayTime.Sub(start).Hours() / 24)
		if dayIndex >= 0 && dayIndex < days {
			dailyStats[dayIndex] = duration
		}
	}

	return dailyStats, nil
}

// GetGroupedStats retourne les stats groupées par app puis par enriched_name
func (db *DB) GetGroupedStats(start, end time.Time) (map[string]map[string]int64, error) {
	query := `
		SELECT 
			app_name,
			COALESCE(enriched_name, app_name) as enriched,
			SUM(duration_seconds) as total_duration
		FROM activities
		WHERE start_time >= ? AND start_time <= ? AND is_idle = 0
		GROUP BY app_name, enriched
		ORDER BY app_name, total_duration DESC
	`

	rows, err := db.conn.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Structure: map[AppName]map[EnrichedName]Duration
	grouped := make(map[string]map[string]int64)

	for rows.Next() {
		var appName, enrichedName string
		var duration int64

		if err := rows.Scan(&appName, &enrichedName, &duration); err != nil {
			return nil, err
		}

		if grouped[appName] == nil {
			grouped[appName] = make(map[string]int64)
		}
		grouped[appName][enrichedName] = duration
	}

	return grouped, nil
}
