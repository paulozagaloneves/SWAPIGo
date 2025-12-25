package main

import (
	"SWAPIGo/handler"
	"SWAPIGo/internal/cache"
	"SWAPIGo/internal/helpers"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// HTTP client with sane timeout for upstream calls
	httpClient := &http.Client{Timeout: 10 * time.Second}

	c := cache.New(500, 5*time.Minute) // capacity 500, TTL 5m

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.InfoHandler)
	mux.HandleFunc("/healthz", handler.HealthHandler)
	mux.HandleFunc("/api/swapi", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/swapi/", http.StatusMovedPermanently)
	})
	mux.Handle("/api/swapi/", handler.SwapiProxyHandler(httpClient, c))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      helpers.Logging(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown setup
	go func() {
		log.Printf("SWAPIGo listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
