package config

import (
	"os"
	"path/filepath"
	"time"
)

// Config contient la configuration globale de l'application
type Config struct {
	// Intervalle de vérification de la fenêtre active (en secondes)
	CheckInterval time.Duration

	// Délai d'inactivité avant de considérer l'utilisateur comme idle (en secondes)
	IdleThreshold time.Duration

	// Chemin de la base de données SQLite
	DBPath string

	// Port du serveur HTTP local
	APIPort string

	// Activer ou non l'API HTTP
	EnableAPI bool
}

// DefaultConfig retourne la configuration par défaut
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".trackmytime", "activities.db")

	// Créer le dossier s'il n'existe pas
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	return &Config{
		CheckInterval: 2 * time.Second,
		IdleThreshold: 60 * time.Second,
		DBPath:        dbPath,
		APIPort:       "8787",
		EnableAPI:     true,
	}
}
