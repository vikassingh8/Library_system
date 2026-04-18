package response

import (
	"encoding/json"
	"net/http"
	"github.com/go-playground/validator/v10"
)




type ErrorResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func ApiResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func ApiErrorResponse(w http.ResponseWriter, status int, err error) error {
	msg := http.StatusText(status)
	if err != nil {
		msg = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(ErrorResponse{Status: status, Error: msg})
}


func ValidateErrs(err validator.ValidationErrors   )ErrorResponse  {
	var errorMsg string
	for _, fieldErr := range err {
		errorMsg += fieldErr.Field() + ": " + fieldErr.Tag() 
	}
	return ErrorResponse{Status: http.StatusBadRequest, Error: errorMsg}
}