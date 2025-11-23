package storage

// migrate crée les tables si elles n'existent pas
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_name TEXT NOT NULL,
			window_title TEXT,
			process_path TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			duration_seconds INTEGER NOT NULL,
			is_idle BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activities_start_time ON activities(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_activities_app_name ON activities(app_name)`,
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS browser_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			tab_title TEXT,
			browser_name TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			duration_seconds INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_browser_events_start_time ON browser_events(start_time)`,
		// Ajouter enriched_name si elle n'existe pas
		`ALTER TABLE activities ADD COLUMN enriched_name TEXT`,
		`CREATE INDEX IF NOT EXISTS idx_activities_enriched_name ON activities(enriched_name)`,
	}

	for _, migration := range migrations {
		// Ignorer les erreurs si la colonne existe déjà (ALTER TABLE)
		db.conn.Exec(migration)
	}

	return nil
}
