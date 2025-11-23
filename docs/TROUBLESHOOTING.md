# Troubleshooting

## Dashboard

### Dashboard blanc / pas de style

**Symptôme:** Page blanche ou sans CSS

**Solutions:**

1. **Hard refresh**
   - Mac: `Cmd+Shift+R`
   - Windows/Linux: `Ctrl+Shift+R`

2. **Vérifier le CSS**
   ```bash
   curl http://localhost:8787/static/js/app.js
   ```

3. **Redémarrer l'agent**
   ```bash
   pkill trackmytime
   ./trackmytime
   ```

---

### Status "Disconnected" rouge

**Symptôme:** Point rouge dans le header, "Hors ligne"

**Cause:** Agent pas démarré ou crash

**Solution:**
```bash
# Vérifier si l'agent tourne
ps aux | grep trackmytime

# Relancer
./trackmytime
```

---

### Charts ne s'affichent pas

**Symptôme:** Espaces vides à la place des graphiques

**Solutions:**

1. **Vérifier Chart.js**
   - Ouvrir Console (F12)
   - Taper: `typeof Chart`
   - Doit retourner `"function"`

2. **Vérifier l'API**
   ```bash
   curl http://localhost:8787/api/stats/hourly?period=today
   ```

3. **Vérifier la console**
   - F12 → Console
   - Chercher erreurs JavaScript

---

### Données ne se mettent pas à jour

**Symptôme:** Stats figées, compteur ne bouge pas

**Solutions:**

1. **Vérifier le compteur**
   - Doit afficher: "Prochain refresh dans 5s... 4s... 3s..."
   - Si figé → problème JavaScript

2. **Refresh manuel**
   - Cliquer sur le bouton "Actualiser"

3. **Vérifier le tracking**
   - Changer de fenêtre plusieurs fois
   - Attendre 2-3 secondes
   - Vérifier la base: `ls -lh ~/.trackmytime/activities.db`

---

### Pas de données / "Aucune donnée disponible"

**Causes possibles:**

1. **Agent vient d'être lancé**
   - Attendre 5-10 secondes
   - Changer de fenêtre pour déclencher un event

2. **Base de données vide**
   ```bash
   sqlite3 ~/.trackmytime/activities.db "SELECT COUNT(*) FROM activities"
   ```

3. **Permissions**
   ```bash
   ls -la ~/.trackmytime/
   # Le fichier doit être writable
   ```

---

### Vue groupée ne fonctionne pas

**Symptôme:** Clic sur "Vue groupée" ne change rien

**Solutions:**

1. **Vérifier l'API groupée**
   ```bash
   curl http://localhost:8787/api/stats/grouped?period=today
   ```

2. **Console navigateur**
   - F12 → Console
   - Chercher erreurs lors du clic

3. **Vider le cache**
   - Hard refresh: `Cmd+Shift+R`

---

### Accordéons se replient automatiquement

**Symptôme:** Dépliement se ferme après 5 secondes

**Cause:** Bug corrigé en v1.3.0

**Solution:** Mettre à jour vers la dernière version
```bash
git pull
make build
pkill trackmytime
./trackmytime
```

---

## Agent

### L'agent ne démarre pas

**Erreur:** `port already in use`

**Solution:**
```bash
# Trouver le processus sur le port 8787
lsof -i :8787

# Tuer le processus
kill -9 <PID>

# Relancer
./trackmytime
```

---

### L'agent crash au démarrage

**Erreur:** `database locked`

**Solution:**
```bash
# Fermer tous les processus utilisant la DB
lsof ~/.trackmytime/activities.db

# Supprimer les lock files
rm ~/.trackmytime/activities.db-shm
rm ~/.trackmytime/activities.db-wal

# Relancer
./trackmytime
```

---

### Tracking ne fonctionne pas

**Symptôme:** Aucune activité détectée

**macOS:**
```bash
# Vérifier permissions Accessibility
Préférences Système → Sécurité → Accessibilité
Ajouter Terminal (ou votre terminal)
```

**Linux:**
```bash
# Vérifier xdotool installé
which xdotool
sudo apt-get install xdotool xprintidle
```

**Windows:**
```powershell
# Vérifier PowerShell
$PSVersionTable.PSVersion
# Doit être >= 5.0
```

---

### Haute consommation CPU/RAM

**Symptôme:** Agent utilise beaucoup de ressources

**Normal:**
- CPU: 0.5-2%
- RAM: 10-20MB

**Si plus élevé:**

1. **Vérifier l'intervalle**
   ```go
   // config/config.go
   CheckInterval: 2 * time.Second  // Augmenter à 5s si besoin
   ```

2. **Vérifier la DB**
   ```bash
   du -h ~/.trackmytime/activities.db
   # Si > 100MB → peut-être nettoyer les vieilles données
   ```

---

## Export

### Export CSV vide

**Symptôme:** Fichier téléchargé mais vide

**Solutions:**

1. **Vérifier les données**
   ```bash
   curl http://localhost:8787/stats/today
   ```

2. **Essayer JSON**
   ```bash
   ./trackmytime-export -format json
   ```

3. **Vérifier la période**
   ```bash
   # Utiliser week si today est vide
   ./trackmytime-export -period week
   ```

---

## Base de données

### Erreur "database disk image is malformed"

**Solution:**
```bash
# Backup
cp ~/.trackmytime/activities.db ~/activities_backup.db

# Réparer
sqlite3 ~/.trackmytime/activities.db "PRAGMA integrity_check;"

# Si échec, recréer
rm ~/.trackmytime/activities.db
./trackmytime  # Créera une nouvelle DB
```

---

### Réinitialiser toutes les données

```bash
# Backup d'abord !
cp ~/.trackmytime/activities.db ~/backup_$(date +%Y%m%d).db

# Supprimer
rm -rf ~/.trackmytime/

# Relancer (créera une nouvelle DB)
./trackmytime
```

---

## Logs

### Où trouver les logs ?

```bash
# Si lancé avec nohup
cat nohup.out

# Si lancé avec redirection
cat trackmytime.log

# Logs temps réel
tail -f trackmytime.log
```

### Activer debug verbose

```go
// cmd/agent/main.go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

Recompiler: `make build`

---

## Performance

### Dashboard lent

**Solutions:**

1. **Réduire l'auto-refresh**
   ```javascript
   // web/static/js/app.js
   const REFRESH_INTERVAL = 10000; // 10s au lieu de 5s
   ```

2. **Limiter les apps affichées**
   ```javascript
   // Dans updateDonutChart
   .slice(0, 5);  // Top 5 au lieu de 8
   ```

3. **Désactiver les animations**
   ```css
   /* web/static/css/style.css */
   * {
     animation: none !important;
     transition: none !important;
   }
   ```

---

## Besoin d'aide ?

Si le problème persiste :

1. **Consulter les logs**
   ```bash
   tail -50 trackmytime.log
   ```

2. **Vérifier la console navigateur**
   - F12 → Console
   - Copier les erreurs

3. **Tester l'API manuellement**
   ```bash
   curl http://localhost:8787/health
   curl http://localhost:8787/stats/today
   ```

4. **Ouvrir une issue GitHub** avec :
   - Version OS
   - Version Go (`go version`)
   - Logs d'erreur
   - Steps to reproduce
