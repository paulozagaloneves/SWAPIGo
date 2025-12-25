package handler

import (
	"io"
	"log"
	"net/http"
	"strings"

	"SWAPIGo/internal/cache"
	"SWAPIGo/internal/helpers"
)

// SwapiProxyHandler proxies read-only GET requests to https://swapi.info/api/
// Usage: GET /api/swapi/people/1 -> https://swapi.info/api/people/1
func SwapiProxyHandler(client *http.Client, c cache.Cache) http.HandlerFunc {
	const base = "https://swapi.info/api/"
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			helpers.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", "only GET is supported")
			return
		}

		// Derive path from URI by trimming the route prefix
		prefix := "/api/swapi/"
		path := strings.TrimPrefix(r.URL.Path, prefix)
		// Normalize path and avoid leading slashes duplication
		path = strings.TrimPrefix(path, "/")

		// Validate resource and optional numeric id
		allowed := map[string]bool{
			"people": true, "films": true, "planets": true, "species": true, "starships": true, "vehicles": true,
			"": true, // allow upstream root
		}
		segs := strings.Split(strings.Trim(path, "/"), "/")
		resource := ""
		if len(segs) > 0 {
			resource = segs[0]
		}
		if !allowed[resource] {
			helpers.WriteError(w, http.StatusBadRequest, "invalid resource", "allowed: people, films, planets, species, starships, vehicles")
			return
		}
		if len(segs) >= 2 {
			id := segs[1]
			for i := 0; i < len(id); i++ {
				if id[i] < '0' || id[i] > '9' {
					helpers.WriteError(w, http.StatusBadRequest, "invalid id", "id must be numeric")
					return
				}
			}
		}

		target := base
		if path != "" {
			target = base + path
		}

		// Cache lookup
		if status, ctype, body, ok := c.Get(target); ok {
			if ctype == "" {
				ctype = "application/json; charset=utf-8"
			}
			w.Header().Set("Content-Type", ctype)
			w.WriteHeader(status)
			if _, err := w.Write(body); err != nil {
				log.Printf("cache write error: %v", err)
			}
			return
		}

		// Forward request
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "failed to build upstream request", err.Error())
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			helpers.WriteError(w, http.StatusBadGateway, "upstream request failed", err.Error())
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			helpers.WriteError(w, http.StatusBadGateway, "failed to read upstream response", err.Error())
			return
		}
		ctype := resp.Header.Get("Content-Type")
		if ctype == "" {
			ctype = "application/json; charset=utf-8"
		}

		// Cache store
		c.Set(target, resp.StatusCode, ctype, data)

		// Write response
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(resp.StatusCode)
		if _, err := w.Write(data); err != nil {
			log.Printf("proxy body write error: %v", err)
		}
	}
}
