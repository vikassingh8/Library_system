package handlers

import (
	"net/http"

	"github.com/library_system/internal/utils/response"
)

// Simple health check endpoint
func HealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.ApiResponse(w, http.StatusOK, map[string]string{
			"status": "healthy",
		})
	}
}
