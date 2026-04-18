package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/library_system/internal/config"
	"github.com/library_system/internal/models"
	"github.com/library_system/internal/storage"
	"github.com/library_system/internal/utils/response"
	"github.com/library_system/utils"
)

func Register(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var userRegister models.User
		err := json.NewDecoder(r.Body).Decode(&userRegister)

		// ✅ empty body
		if errors.Is(err, io.EOF) {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("empty body"))
			return
		}

		// ✅ any other decode error
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		if err := validator.New().Struct(userRegister); err != nil {
			validationError := err.(validator.ValidationErrors)
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New(response.ValidateErrs(validationError).Error))
			return
		}

		// Normalize email (lowercase and trim)
		userRegister.Email = strings.ToLower(strings.TrimSpace(userRegister.Email))

		hashedPassword, err := utils.HashPassword(userRegister.Password)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}
		userRegister.Password = hashedPassword

		id, err := db.CreateUser(
			userRegister.Name,
			userRegister.Email,
			userRegister.Password,
		)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusConflict, errors.New("user already exist"))
			return
		}

		response.ApiResponse(w, http.StatusOK, map[string]interface{}{"userId": id, "message": "user created successfully"})
	})
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func Login(db storage.Storage, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var req LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if errors.Is(err, io.EOF) {
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New("empty body"))
			return
		}
		if err != nil {
			response.ApiErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		if err := validator.New().Struct(req); err != nil {
			validationError := err.(validator.ValidationErrors)
			response.ApiErrorResponse(w, http.StatusBadRequest, errors.New(response.ValidateErrs(validationError).Error))
			return
		}

		// Normalize email (lowercase and trim)
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		userDetails, err := db.GetUserByEmail(req.Email)

		if err != nil || userDetails == nil {
			response.ApiErrorResponse(w, http.StatusUnauthorized, errors.New("invalid credentials"))
			return
		}
		err = utils.CheckPassword(req.Password, userDetails.Password)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusUnauthorized, errors.New("invalid credentials"))
			return
		}

		id := userDetails.ID

		expiryDuration, err := time.ParseDuration(cfg.JWT.Expiry)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, errors.New("invalid JWT expiry configuration"))
			return
		}

		// Generate JWT token (includes role)
		token, err := utils.GenerateJWT(id, userDetails.Role, cfg.JWT.Secret, expiryDuration)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		// Set cookie only after successful token generation
		// SameSite=None REQUIRES Secure=true (mandatory by browser)
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // REQUIRED when using SameSite=None
			SameSite: http.SameSiteNoneMode,
			MaxAge:   int(expiryDuration.Seconds()),
		})

		response.ApiResponse(w, http.StatusOK, map[string]string{
			"message": "login successful",
			"name":    userDetails.Name,
			"email":   userDetails.Email,
			"role":    userDetails.Role,
		})

	})

}

// Me returns the current user's info — used by the frontend to restore session from cookie on page load.
func Me(db storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)
		user, err := db.GetUserByID(userId)
		if err != nil {
			response.ApiErrorResponse(w, http.StatusUnauthorized, errors.New("user not found"))
			return
		}
		response.ApiResponse(w, http.StatusOK, map[string]string{
			"id":    fmt.Sprintf("%d", user.ID),
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		})
	})
}

func Logout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Clear the auth_token cookie by setting MaxAge to -1
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // REQUIRED when using SameSite=None
			SameSite: http.SameSiteNoneMode,
			MaxAge:   -1, // This deletes the cookie
		})

		response.ApiResponse(w, http.StatusOK, map[string]string{"message": "logout successful"})
	})
}
