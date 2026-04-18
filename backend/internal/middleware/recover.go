package middleware

import (
	"log"
	"net/http"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("PANIC:", err)
				http.Error(w, "Internal Server Error", 500)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
