package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8092"
	}
	gateway := strings.TrimRight(os.Getenv("MINIMALL_GATEWAY_BASE_URL"), "/")
	if gateway == "" {
		gateway = "http://127.0.0.1:8080"
	}

	allowedOrigins := map[string]struct{}{
		"http://127.0.0.1:5300": {},
		"http://localhost:5300": {},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /open/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"service": "open-server",
			"status":  "UP",
			"gateway": gateway,
			"note":    "Go shell — add ISV auth and scoped APIs here",
		})
	})
	mux.HandleFunc("GET /open/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"message": "open platform pong",
			"gateway": gateway,
		})
	})

	handler := withCORS(mux, allowedOrigins)
	addr := "127.0.0.1:" + port
	log.Printf("open-server listening on http://%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler, allowed map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-App-Key")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, body map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
