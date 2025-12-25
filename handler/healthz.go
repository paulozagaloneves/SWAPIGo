package handler

import (
	"SWAPIGo/internal/helpers"
	"net/http"
	"time"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "SWAPIGo",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}
