package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Only allow HTTPS in production domains.
	// Localhost is allowed over HTTP for development.
	host := strings.ToLower(u.Hostname())

	// Local development
	if u.Scheme == "http" &&
		(host == "localhost" || host == "127.0.0.1") &&
		(u.Port() == "5173" || u.Port() == "5174") {
		return true
	}

	// Production
	if u.Scheme != "https" {
		return false
	}

	// nid.xyz itself + all subdomains
	if host == "nid.xyz" || strings.HasSuffix(host, ".nid.xyz") {
		return true
	}

	// vercel.app itself + all subdomains
	if host == "vercel.app" || strings.HasSuffix(host, ".vercel.app") {
		return true
	}

	return false
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isAllowedOrigin(origin) {
			w.Header().Set(
				"Access-Control-Allow-Origin",
				origin,
			)

			w.Header().Set(
				"Access-Control-Allow-Credentials",
				"true",
			)

			w.Header().Set(
				"Vary",
				"Origin",
			)
		}

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Accept, Authorization, Content-Type, X-CSRF-Token",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
