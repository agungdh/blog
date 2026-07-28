package middleware

import (
	"crypto/tls"
	"net/http"
	"strings"
)

func ForwardedProto(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			for _, fwd := range r.Header.Values("Forwarded") {
				for _, part := range strings.Split(fwd, ";") {
					kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
					if len(kv) == 2 && strings.TrimSpace(kv[0]) == "proto" {
						proto = strings.Trim(kv[1], `"`)
					}
				}
			}
		}
		if proto == "https" {
			r.TLS = &tls.ConnectionState{}
		}
		next.ServeHTTP(w, r)
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
