package handler

import (
	"SWAPIGo/internal/helpers"
	"net/http"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSON(w, http.StatusOK, map[string]any{
		"name":        "SWAPIGo",
		"description": "Simple REST microservice proxying the Star Wars API (swapi.info)",
		"endpoints":   []string{"GET /healthz", "GET /api/swapi?path=<resource>"},
		"version":     "v0.1.0",
	})
}
