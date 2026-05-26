package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"syncspace/backend/internal/api"
	"syncspace/backend/internal/auth"
	"syncspace/backend/internal/config"
	"syncspace/backend/internal/service"
	"syncspace/backend/internal/store"
	"syncspace/backend/internal/websocket"
)

func main() {
	cfg := config.Load()
	auth.SetJWTSecret(cfg.JWTSecret)
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()

	svc := service.New(st, cfg.UploadDir)
	h := api.New(svc)
	mux := http.NewServeMux()

	// WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()
	mux.HandleFunc("/ws", hub.HandleWebSocket)

	h.Register(mux)

	// Serve uploaded files
	os.MkdirAll("uploads", 0755)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	server := &http.Server{Addr: cfg.Addr, Handler: withCORS(mux)}
	log.Printf("syncspace backend listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow same-origin (no Origin header, e.g., through nginx proxy)
		// or localhost origins for development
		if origin == "" || strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "https://localhost") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
