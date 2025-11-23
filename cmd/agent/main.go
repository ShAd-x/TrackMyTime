package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trackmytime/config"
	"trackmytime/internal/api"
	"trackmytime/internal/storage"
	"trackmytime/internal/tracker"
)

func main() {
	log.Println("🚀 TrackMyTime Agent démarrage...")

	// Charger la configuration
	cfg := config.DefaultConfig()
	log.Printf("📁 Base de données: %s", cfg.DBPath)
	log.Printf("⏱️  Intervalle de vérification: %v", cfg.CheckInterval)
	log.Printf("💤 Seuil d'inactivité: %v", cfg.IdleThreshold)

	// Connexion à la base de données
	db, err := storage.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Erreur connexion DB: %v", err)
	}
	defer db.Close()
	log.Println("✅ Base de données initialisée")

	// Créer le détecteur d'inactivité
	idleDetector := tracker.NewIdleDetector(cfg.IdleThreshold)

	// Démarrer le serveur API en arrière-plan
	if cfg.EnableAPI {
		apiServer := api.NewServer(db, cfg.APIPort)
		go func() {
			if err := apiServer.Start(); err != nil {
				log.Printf("⚠️  Erreur serveur API: %v", err)
			}
		}()
	}

	// Variables pour le tracking
	var currentWindow *tracker.WindowInfo
	var activityStartTime time.Time
	var wasIdle bool
	var idleStartTime time.Time

	// Ticker pour vérifier la fenêtre active
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	// Gérer l'arrêt propre
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("🎯 Agent démarré - tracking en cours...")

	// Boucle principale
	for {
		select {
		case <-ticker.C:
			// Vérifier si l'utilisateur est idle
			isIdle, err := idleDetector.IsIdle()
			if err != nil {
				log.Printf("⚠️  Erreur détection idle: %v", err)
				continue
			}

			// Si l'utilisateur devient idle
			if isIdle && !wasIdle {
				if currentWindow != nil {
					// Enregistrer l'activité avant l'idle
					endTime := time.Now()
					duration := endTime.Sub(activityStartTime)

					activity := &storage.Activity{
						AppName:      currentWindow.AppName,
						WindowTitle:  currentWindow.WindowTitle,
					EnrichedName: currentWindow.GetEnrichedName(),
						ProcessPath:  currentWindow.ProcessPath,
						StartTime:    activityStartTime,
						EndTime:      endTime,
						DurationSecs: int64(duration.Seconds()),
						IsIdle:       false,
					}

					if err := db.InsertActivity(activity); err != nil {
						log.Printf("❌ Erreur sauvegarde activité: %v", err)
					} else {
						log.Printf("💾 Activité sauvegardée: %s (%s) - %.0fs",
							activity.AppName,
							activity.WindowTitle,
							duration.Seconds())
					}

					currentWindow = nil
				}
				wasIdle = true
				idleStartTime = time.Now()
				log.Println("💤 Utilisateur inactif")
				continue
			}

			// Si l'utilisateur était idle et redevient actif
			if !isIdle && wasIdle {
				// Enregistrer la période d'inactivité
				endTime := time.Now()
				idleDuration := endTime.Sub(idleStartTime)
				
				idleActivity := &storage.Activity{
					AppName:      "IDLE",
					WindowTitle:  "Inactif",
					ProcessPath:  "",
					StartTime:    idleStartTime,
					EndTime:      endTime,
					DurationSecs: int64(idleDuration.Seconds()),
					IsIdle:       true,
				}
				
				if err := db.InsertActivity(idleActivity); err != nil {
					log.Printf("❌ Erreur sauvegarde période idle: %v", err)
				} else {
					log.Printf("💾 Période idle sauvegardée: %.0fs", idleDuration.Seconds())
				}
				
				wasIdle = false
				log.Println("👋 Utilisateur de retour")
			}

			// Si l'utilisateur devient idle
			if isIdle && !wasIdle {
				if currentWindow != nil {
					// Enregistrer l'activité avant l'idle
					endTime := time.Now()
					duration := endTime.Sub(activityStartTime)

					activity := &storage.Activity{
						AppName:      currentWindow.AppName,
						WindowTitle:  currentWindow.WindowTitle,
					EnrichedName: currentWindow.GetEnrichedName(),
						ProcessPath:  currentWindow.ProcessPath,
						StartTime:    activityStartTime,
						EndTime:      endTime,
						DurationSecs: int64(duration.Seconds()),
						IsIdle:       false,
					}

					if err := db.InsertActivity(activity); err != nil {
						log.Printf("❌ Erreur sauvegarde activité: %v", err)
					} else {
						log.Printf("💾 Activité sauvegardée: %s (%s) - %.0fs",
							activity.AppName,
							activity.WindowTitle,
							duration.Seconds())
					}

					currentWindow = nil
				}
				wasIdle = true
				log.Println("💤 Utilisateur inactif")
				continue
			}

			// Si l'utilisateur était idle et redevient actif
			if !isIdle && wasIdle {
				wasIdle = false
				log.Println("👋 Utilisateur de retour")
			}

			// Si l'utilisateur n'est pas idle, vérifier la fenêtre active
			if !isIdle {
				window, err := tracker.GetActiveWindow()
				if err != nil {
					log.Printf("⚠️  Erreur récupération fenêtre: %v", err)
					continue
				}

				// Si la fenêtre a changé
				if currentWindow == nil ||
					window.AppName != currentWindow.AppName ||
					window.WindowTitle != currentWindow.WindowTitle {

					// Sauvegarder l'activité précédente
					if currentWindow != nil {
						endTime := time.Now()
						duration := endTime.Sub(activityStartTime)

					activity := &storage.Activity{
						AppName:      currentWindow.AppName,
						EnrichedName: currentWindow.GetEnrichedName(),
						WindowTitle:  currentWindow.WindowTitle,
						ProcessPath:  currentWindow.ProcessPath,
						StartTime:    activityStartTime,
						EndTime:      endTime,
						DurationSecs: int64(duration.Seconds()),
						IsIdle:       false,
					}

						if err := db.InsertActivity(activity); err != nil {
							log.Printf("❌ Erreur sauvegarde activité: %v", err)
						} else {
							log.Printf("💾 Activité sauvegardée: %s (%s) - %.0fs",
								activity.AppName,
								activity.WindowTitle,
								duration.Seconds())
						}
					}

					// Commencer le tracking de la nouvelle activité
					currentWindow = window
					activityStartTime = time.Now()
					log.Printf("🔄 Changement d'activité: %s - %s",
						window.AppName,
						window.WindowTitle)

					// Mettre à jour l'API avec l'activité courante
					if cfg.EnableAPI {
						// Note: Pour accéder à apiServer ici, il faudrait le passer via un channel
						// ou le rendre accessible globalement. Pour simplifier, on skip cette partie.
					}
				}
			}

		case <-sigChan:
			log.Println("\n👋 Arrêt de l'agent...")

			// Sauvegarder l'activité courante avant de quitter
			if currentWindow != nil && !wasIdle {
				endTime := time.Now()
				duration := endTime.Sub(activityStartTime)

				activity := &storage.Activity{
					AppName:      currentWindow.AppName,
					EnrichedName: currentWindow.GetEnrichedName(),
					WindowTitle:  currentWindow.WindowTitle,
					ProcessPath:  currentWindow.ProcessPath,
					StartTime:    activityStartTime,
					EndTime:      endTime,
					DurationSecs: int64(duration.Seconds()),
					IsIdle:       false,
				}

				if err := db.InsertActivity(activity); err != nil {
					log.Printf("❌ Erreur sauvegarde activité finale: %v", err)
				} else {
					log.Printf("💾 Activité finale sauvegardée: %s (%s) - %.0fs",
						activity.AppName,
						activity.WindowTitle,
						duration.Seconds())
				}
			}

			log.Println("✅ Agent arrêté proprement")
			return
		}
	}
}
