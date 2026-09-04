package main

import (
	"context"
	"crypto/sha256"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type apiDoc struct {
	Name    string `json:"name"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

type ctxKey string

const routeKey ctxKey = "route"

type routeStats struct {
	count       atomic.Uint64
	errors      atomic.Uint64
	totalMicros  atomic.Uint64
	maxMicros    atomic.Uint64
	status2xx    atomic.Uint64
	status3xx    atomic.Uint64
	status4xx    atomic.Uint64
	status5xx    atomic.Uint64
}

type metricsRegistry struct {
	startTime time.Time
	inFlight  atomic.Int64
	totalReq  atomic.Uint64
	mu        sync.Mutex
	routes    map[string]*routeStats
}

func newMetricsRegistry() *metricsRegistry {
	return &metricsRegistry{
		startTime: time.Now().UTC(),
		routes:    make(map[string]*routeStats),
	}
}

func (m *metricsRegistry) getRoute(name string) *routeStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.routes[name]; ok {
		return s
	}
	s := &routeStats{}
	m.routes[name] = s
	return s
}

var metrics = newMetricsRegistry()

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
		{Name: "product", Method: http.MethodGet, Path: "/open/v1/products/{skuId}", Summary: "商品读接口（经 Gateway → product-service）"},
		{Name: "sign", Method: http.MethodPost, Path: "/open/v1/debug/sign", Summary: "签名预览"},
		{Name: "execute", Method: http.MethodPost, Path: "/open/v1/debug/execute", Summary: "执行调试请求"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /open/v1/health", observe("health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"service":  "open-server",
			"status":   "UP",
			"gateway":  gateway,
			"time":     time.Now().UTC().Format(time.RFC3339),
			"version":  "0.1.0",
			"note":     "Open API debug shell is running",
		})
	}))
	mux.HandleFunc("GET /open/v1/ping", observe("ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"message": "open platform pong",
			"gateway": gateway,
		})
	}))
	mux.HandleFunc("GET /open/v1/apis", observe("apis", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"count": len(apis),
			"items": apis,
		})
	}))
	mux.HandleFunc("GET /open/v1/products/{skuId}", observe("product", func(w http.ResponseWriter, r *http.Request) {
		skuId := r.PathValue("skuId")
		if skuId == "" {
			writeJSONStatus(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "missing skuId"})
			return
		}
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, gateway+"/api/products/"+skuId, nil)
		if err != nil {
			writeJSONStatus(w, http.StatusBadGateway, map[string]any{"code": 502, "message": err.Error()})
			return
		}
		propagateTrace(r, req)
		upstream, err := http.DefaultClient.Do(req)
		if err != nil {
			writeJSONStatus(w, http.StatusBadGateway, map[string]any{"code": 502, "message": err.Error()})
			return
		}
		defer upstream.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(upstream.Body).Decode(&payload); err != nil {
			writeJSONStatus(w, http.StatusBadGateway, map[string]any{"code": 502, "message": "invalid upstream response"})
			return
		}
		writeJSONStatus(w, upstream.StatusCode, map[string]any{
			"source":  "open-server",
			"gateway": gateway,
			"upstream": payload,
		})
	}))

	mux.HandleFunc("POST /open/v1/debug/sign", observe("debug_sign", func(w http.ResponseWriter, r *http.Request) {
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
	}))
	mux.HandleFunc("POST /open/v1/debug/execute", observe("debug_execute", func(w http.ResponseWriter, r *http.Request) {
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
	}))
	mux.HandleFunc("GET /metrics", observe("metrics", func(w http.ResponseWriter, r *http.Request) {
		writeMetrics(w)
	}))

	handler := withObservability(withCORS(mux, allowedOrigins))
	addr := "127.0.0.1:" + port
	log.Printf("open-server listening on http://%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func observe(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), routeKey, route)
		next(w, r.WithContext(ctx))
	}
}

func routeName(r *http.Request) string {
	if v, ok := r.Context().Value(routeKey).(string); ok && v != "" {
		return v
	}
	if route := r.Header.Get("X-Route-Name"); route != "" {
		return route
	}
	return r.URL.Path
}

func withObservability(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = newHexID(12)
		}
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = requestID
		}
		start := time.Now()
		metrics.inFlight.Add(1)
		metrics.totalReq.Add(1)

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		w.Header().Set("X-Request-Id", requestID)
		w.Header().Set("X-Trace-Id", traceID)

		next.ServeHTTP(rw, r)

		metrics.inFlight.Add(-1)
		durationMicros := uint64(time.Since(start).Microseconds())
		stats := metrics.getRoute(routeName(r))
		stats.count.Add(1)
		stats.totalMicros.Add(durationMicros)
		for {
			old := stats.maxMicros.Load()
			if durationMicros <= old || stats.maxMicros.CompareAndSwap(old, durationMicros) {
				break
			}
		}
		if rw.status >= 500 {
			stats.errors.Add(1)
		}
		switch {
		case rw.status >= 200 && rw.status < 300:
			stats.status2xx.Add(1)
		case rw.status >= 300 && rw.status < 400:
			stats.status3xx.Add(1)
		case rw.status >= 400 && rw.status < 500:
			stats.status4xx.Add(1)
		default:
			stats.status5xx.Add(1)
		}

		log.Printf(
			`{"ts":"%s","request_id":"%s","trace_id":"%s","method":"%s","path":"%s","route":"%s","status":%d,"duration_ms":%.2f,"remote_addr":"%s"}`,
			time.Now().UTC().Format(time.RFC3339Nano),
			requestID,
			traceID,
			r.Method,
			r.URL.Path,
			routeName(r),
			rw.status,
			float64(durationMicros)/1000.0,
			r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(p)
}

func propagateTrace(r *http.Request, req *http.Request) {
	if rid := r.Header.Get("X-Request-Id"); rid != "" {
		req.Header.Set("X-Request-Id", rid)
	}
	if tid := r.Header.Get("X-Trace-Id"); tid != "" {
		req.Header.Set("X-Trace-Id", tid)
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

func writeMetrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP minimall_open_server_uptime_seconds Server uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE minimall_open_server_uptime_seconds gauge\n")
	fmt.Fprintf(w, "minimall_open_server_uptime_seconds %.0f\n", time.Since(metrics.startTime).Seconds())
	fmt.Fprintf(w, "# HELP minimall_open_server_in_flight_requests Current in-flight requests\n")
	fmt.Fprintf(w, "# TYPE minimall_open_server_in_flight_requests gauge\n")
	fmt.Fprintf(w, "minimall_open_server_in_flight_requests %d\n", metrics.inFlight.Load())
	fmt.Fprintf(w, "# HELP minimall_open_server_total_requests Total requests handled\n")
	fmt.Fprintf(w, "# TYPE minimall_open_server_total_requests counter\n")
	fmt.Fprintf(w, "minimall_open_server_total_requests %d\n", metrics.totalReq.Load())

	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	for name, s := range metrics.routes {
		count := s.count.Load()
		avgMs := 0.0
		if count > 0 {
			avgMs = float64(s.totalMicros.Load()) / 1000.0 / float64(count)
		}
		fmt.Fprintf(w, "# HELP minimall_open_server_route_requests_total Requests per route\n")
		fmt.Fprintf(w, "# TYPE minimall_open_server_route_requests_total counter\n")
		fmt.Fprintf(w, "minimall_open_server_route_requests_total{route=%q} %d\n", name, count)
		fmt.Fprintf(w, "# HELP minimall_open_server_route_errors_total Errors per route\n")
		fmt.Fprintf(w, "# TYPE minimall_open_server_route_errors_total counter\n")
		fmt.Fprintf(w, "minimall_open_server_route_errors_total{route=%q} %d\n", name, s.errors.Load())
		fmt.Fprintf(w, "minimall_open_server_route_duration_avg_ms{route=%q} %.3f\n", name, avgMs)
		fmt.Fprintf(w, "minimall_open_server_route_duration_max_ms{route=%q} %.3f\n", name, float64(s.maxMicros.Load())/1000.0)
		fmt.Fprintf(w, "minimall_open_server_route_status_2xx_total{route=%q} %d\n", name, s.status2xx.Load())
		fmt.Fprintf(w, "minimall_open_server_route_status_3xx_total{route=%q} %d\n", name, s.status3xx.Load())
		fmt.Fprintf(w, "minimall_open_server_route_status_4xx_total{route=%q} %d\n", name, s.status4xx.Load())
		fmt.Fprintf(w, "minimall_open_server_route_status_5xx_total{route=%q} %d\n", name, s.status5xx.Load())
	}
}

func newHexID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, body map[string]any) {
	writeJSONStatus(w, http.StatusOK, body)
}

func writeJSONStatus(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
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
