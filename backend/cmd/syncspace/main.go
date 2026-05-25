package main

import (
	"log"
	"net/http"
	"os"

	"syncspace/backend/internal/api"
	"syncspace/backend/internal/config"
	"syncspace/backend/internal/service"
	"syncspace/backend/internal/store"
)

func main() {
	cfg := config.Load()
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	h := api.New(svc)
	mux := http.NewServeMux()
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
