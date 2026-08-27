package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type apiDoc struct {
	Name    string `json:"name"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

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
		"http://localhost:5300":  {},
	}

	apis := []apiDoc{
		{Name: "health", Method: http.MethodGet, Path: "/open/v1/health", Summary: "健康检查"},
		{Name: "apis", Method: http.MethodGet, Path: "/open/v1/apis", Summary: "API 目录"},
		{Name: "sign", Method: http.MethodPost, Path: "/open/v1/debug/sign", Summary: "签名预览"},
		{Name: "execute", Method: http.MethodPost, Path: "/open/v1/debug/execute", Summary: "执行调试请求"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /open/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"service":  "open-server",
			"status":   "UP",
			"gateway":  gateway,
			"time":     time.Now().UTC().Format(time.RFC3339),
			"version":  "0.1.0",
			"note":     "Open API debug shell is running",
		})
	})
	mux.HandleFunc("GET /open/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"message": "open platform pong",
			"gateway": gateway,
		})
	})
	mux.HandleFunc("GET /open/v1/apis", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"count": len(apis),
			"items": apis,
		})
	})
	mux.HandleFunc("POST /open/v1/debug/sign", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		keys := make([]string, 0, len(payload))
		for k := range payload {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var raw strings.Builder
		for _, key := range keys {
			raw.WriteString(key)
			raw.WriteString("=")
			raw.WriteString(toString(payload[key]))
			raw.WriteString("&")
		}
		sum := sha256.Sum256([]byte(raw.String()))
		writeJSON(w, map[string]any{
			"raw":      strings.TrimSuffix(raw.String(), "&"),
			"signature": hex.EncodeToString(sum[:]),
		})
	})
	mux.HandleFunc("POST /open/v1/debug/execute", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		writeJSON(w, map[string]any{
			"ok":        true,
			"request":   payload,
			"debugMode": true,
			"echo": map[string]any{
				"gateway": gateway,
				"headers": map[string]any{
					"content-type": r.Header.Get("Content-Type"),
				},
			},
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-App-Key, X-App-Secret")
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

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case nil:
		return ""
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
