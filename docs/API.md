# API Documentation

L'agent TrackMyTime expose une API REST sur `http://localhost:8787`

## Endpoints

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

---

### Activité Courante

```http
GET /activity/current
```

**Response:**
```json
{
  "app_name": "Brave Browser",
  "window_title": "GitHub - TrackMyTime",
  "start_time": "2024-01-01T10:30:00Z",
  "duration_seconds": 120
}
```

**Response (no activity):**
```json
{
  "status": "no activity"
}
```

---

### Stats du Jour

```http
GET /stats/today
```

**Response:**
```json
{
  "total_active_seconds": 28800,
  "total_idle_seconds": 3600,
  "stats_by_app": {
    "Brave Browser": 10800,
    "Visual Studio Code": 7200,
    "Terminal": 3600
  }
}
```

---

### Stats de la Semaine

```http
GET /stats/week
```

**Response:** Même format que `/stats/today`

---

### Stats Horaires

```http
GET /api/stats/hourly?period={today|week}
```

**Parameters:**
- `period` (required): `today` ou `week`

**Response:**
```json
{
  "hourly_data": [
    0, 120, 1800, 3600, 5400, ..., 0
  ]
}
```

Array de 24 éléments (0-23h), valeurs en secondes.

---

### Stats Groupées

```http
GET /api/stats/grouped?period={today|week}
```

**Parameters:**
- `period` (required): `today` ou `week`

**Response:**
```json
{
  "groups": [
    {
      "app_name": "Brave Browser",
      "total_seconds": 10800,
      "children": [
        {
          "name": "X",
          "duration": 5400
        },
        {
          "name": "YouTube",
          "duration": 3600
        },
        {
          "name": "Autres",
          "duration": 1800
        }
      ]
    }
  ]
}
```

Les sites reconnus sont groupés par nom (X, YouTube, GitHub, etc.), les autres sous "Autres".

---

### Export

```http
GET /export/aggregated?period={today|week}&format={csv|json}
```

**Parameters:**
- `period` (required): `today` ou `week`
- `format` (required): `csv` ou `json`

**Response CSV:**
```csv
Application,Duration,Percentage
Brave Browser,03:00:00,37.5%
Visual Studio Code,02:00:00,25.0%
```

**Response JSON:**
```json
[
  {
    "app_name": "Brave Browser",
    "duration_seconds": 10800,
    "duration_formatted": "03:00:00",
    "percentage": 37.5
  }
]
```

---

## CORS

Par défaut, l'API accepte uniquement les requêtes depuis `localhost`.

## Rate Limiting

Aucune limite de taux pour l'instant (usage local uniquement).

## Authentification

Aucune authentification requise (API locale).

⚠️ **Sécurité:** N'exposez pas cette API publiquement sur internet.

## Codes d'erreur

- `200 OK` - Succès
- `400 Bad Request` - Paramètres invalides
- `404 Not Found` - Endpoint inexistant
- `500 Internal Server Error` - Erreur serveur
