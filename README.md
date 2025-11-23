# TrackMyTime ⏱️

Agent desktop pour tracker automatiquement votre temps d'activité avec un dashboard web moderne.

## 🚀 Quick Start

```bash
# Build
make build

# Lancer
./trackmytime

# Dashboard
open http://localhost:8787/
```

## ✨ Fonctionnalités

- 🎯 **Tracking automatique** - Détecte la fenêtre active et l'application utilisée
- 📊 **Dashboard temps réel** - Interface web moderne avec graphiques interactifs
- 🔍 **Vue groupée intelligente** - Reconnaissance de 20+ sites populaires (X, YouTube, GitHub, etc.)
- 📥 **Export CSV/JSON** - Exportez vos données facilement
- 💤 **Détection inactivité** - Track uniquement quand vous êtes actif
- 🎨 **Design moderne** - Interface glassmorphism avec animations fluides
- 🔒 **100% local** - Aucune donnée envoyée en ligne, tout reste sur votre machine

## 📦 Installation

### Prérequis

- **Go** 1.21+
- **macOS** : Aucune dépendance externe
- **Linux** : `xdotool`, `xprintidle`
  ```bash
  sudo apt-get install xdotool xprintidle
  ```
- **Windows** : PowerShell (inclus)

### Build

```bash
# Compiler tout
make build

# Ou manuellement
go build -o trackmytime ./cmd/agent
go build -o trackmytime-export ./cmd/export
```

## 📖 Usage

### Agent

```bash
# Démarrer (foreground)
./trackmytime

# Démarrer (background)
nohup ./trackmytime > /dev/null 2>&1 &

# Arrêter
pkill trackmytime
```

L'agent démarre automatiquement :
- 🌐 **Dashboard web** sur http://localhost:8787/
- 📡 **API REST** sur http://localhost:8787/
- 💾 **Base SQLite** dans `~/.trackmytime/activities.db`

### Dashboard Web

Ouvrez `http://localhost:8787/` pour accéder au dashboard.

**Fonctionnalités :**
- 📊 Vue Today/Week avec switch
- 🎯 Vue groupée (sites reconnus automatiquement)
- 📈 Graphiques donut + timeline 24h
- 🏆 Top applications avec classement
- 📥 Export CSV/JSON en un clic
- 🔄 Auto-refresh toutes les 5s

### Export CLI

```bash
# Export today en CSV
./trackmytime-export

# Export en JSON
./trackmytime-export -format json

# Export de la semaine
./trackmytime-export -period week

# Export agrégé
./trackmytime-export -aggregated
```

## 📁 Structure du projet

```
TrackMyTime/
├── cmd/
│   ├── agent/              # Agent principal
│   └── export/             # Tool d'export CLI
├── internal/
│   ├── api/                # Serveur HTTP + endpoints
│   ├── storage/            # SQLite + migrations
│   ├── tracker/            # Détection fenêtre active
│   └── export/             # Logique d'export
├── web/
│   ├── index.html          # Dashboard
│   └── static/
│       └── js/app.js       # Logique frontend
├── config/                 # Configuration
├── docs/                   # Documentation
│   ├── API.md             # Documentation API REST
│   └── TROUBLESHOOTING.md # Guide de dépannage
├── scripts/                # Scripts utilitaires
├── Makefile               # Commandes de build
└── README.md              # Ce fichier
```

## 🎨 Sites reconnus

L'agent reconnaît automatiquement 20+ sites populaires et les regroupe intelligemment :

**Réseaux sociaux :** X (Twitter) • TikTok • Instagram • Facebook • LinkedIn • Reddit

**Vidéo/Streaming :** YouTube • Twitch • Netflix • Spotify

**Productivité :** Gmail • GitHub • Slack • Discord • Notion • Google Drive • Stack Overflow

**IA :** ChatGPT • Claude

*Tous les autres sites sont regroupés sous "Autres"*

## 🗄️ Base de données

**Location :** `~/.trackmytime/activities.db` (SQLite)

**Backup :**
```bash
cp ~/.trackmytime/activities.db ~/backup/activities_$(date +%Y%m%d).db
```

**Structure :**
- `activities` - Historique complet des activités
- `config` - Configuration persistée
- `browser_events` - Préparé pour extension navigateur future

## 🔧 Configuration

Fichier `config/config.go` :

```go
type Config struct {
    APIPort        string        // "8787"
    CheckInterval  time.Duration // 2s
    IdleThreshold  time.Duration // 60s
    DBPath         string        // ~/.trackmytime/activities.db
}
```

Modifier et recompiler : `make build`

## 🔌 API

Documentation complète : [docs/API.md](docs/API.md)

**Endpoints principaux :**
```
GET /health                              # Status
GET /activity/current                    # Activité en cours
GET /stats/today                         # Stats du jour
GET /stats/week                          # Stats de la semaine
GET /api/stats/hourly?period=today       # Timeline 24h
GET /api/stats/grouped?period=today      # Vue groupée
GET /export/aggregated?period=today&format=csv
```

## 🐛 Troubleshooting

**Dashboard ne charge pas :**
```bash
ps aux | grep trackmytime    # Vérifier si l'agent tourne
pkill trackmytime            # Arrêter
./trackmytime                # Relancer
```

**Pas de données :**
- Attendre 2-3 secondes après le démarrage
- Changer de fenêtre pour déclencher un event

**Dashboard blanc :**
- Hard refresh : `Cmd+Shift+R` (Mac) ou `Ctrl+Shift+R` (Win/Linux)

Guide complet : [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)

## 📝 Makefile

```bash
make build        # Compiler agent + export
make clean        # Nettoyer les binaires
make run          # Compiler et lancer l'agent
make export       # Build seulement l'outil export
make help         # Afficher l'aide
```

## 🔒 Sécurité

- ✅ Tout fonctionne en local (localhost uniquement)
- ✅ Aucune donnée envoyée en ligne
- ✅ Pas de télémétrie ni tracking externe
- ✅ Base SQLite locale et chiffrable si besoin
- ⚠️ API non authentifiée (usage local uniquement)

## 🤝 Contribution

Les contributions sont les bienvenues ! Pour contribuer :

1. Forkez le projet
2. Créez une branche (`git checkout -b feature/amazing`)
3. Committez vos changements (`git commit -m 'Add amazing feature'`)
4. Pushez vers la branche (`git push origin feature/amazing`)
5. Ouvrez une Pull Request

## 📄 License

MIT License - Voir [LICENSE](LICENSE)

## 🔗 Documentation

- [API REST](docs/API.md) - Documentation complète de l'API
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Guide de dépannage

## 🙏 Crédits

Construit avec :
- **Go** - Backend
- **SQLite** - Base de données
- **Chart.js** - Graphiques interactifs
- **Tailwind CSS** - Styling
- **AppleScript** (macOS) - Détection fenêtre active

---

**Version :** 1.3.0  
**Go Version :** 1.21+ requis
