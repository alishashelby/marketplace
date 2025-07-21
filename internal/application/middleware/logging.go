package middleware

import (
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf(
			"loggingMiddleware: new request with: remote_addr %s "+
				"method %s url %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
		)
	})
}
