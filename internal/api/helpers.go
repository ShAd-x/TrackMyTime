package api

import (
	"fmt"
	"net/http"
	"time"

	"trackmytime/internal/storage"
)

// parseCustomPeriod parses custom start and end dates from query parameters
func (s *Server) parseCustomPeriod(r *http.Request) (start, end time.Time, err error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start and end dates required")
	}

	now := time.Now()
	start, err = time.ParseInLocation("2006-01-02", startStr, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format (YYYY-MM-DD)")
	}

	end, err = time.ParseInLocation("2006-01-02", endStr, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format (YYYY-MM-DD)")
	}

	// Add 1 day to end to include the entire end date
	end = end.Add(24 * time.Hour)

	if end.Before(start) || end.Equal(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end date must be after start date")
	}

	return start, end, nil
}

// getPeriodBounds calculates start and end time for a given period
func (s *Server) getPeriodBounds(period string, r *http.Request) (start, end time.Time, err error) {
	now := time.Now()

	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour)

	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		end = start.Add(7 * 24 * time.Hour)

	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)

	case "custom":
		return s.parseCustomPeriod(r)

	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period (today, week, month, custom)")
	}

	return start, end, nil
}

// calculateIdleTime calculates total idle seconds from activities
func calculateIdleTime(activities []storage.Activity) int64 {
	var totalIdleSeconds int64
	for _, activity := range activities {
		if activity.IsIdle {
			totalIdleSeconds += activity.DurationSecs
		}
	}
	return totalIdleSeconds
}
