// Package auth — Auth error types и утилиты для HTTP-ответов с ошибками.
package auth

import (
	"encoding/json"
	"net/http"
)

// AuthError — структура ошибки аутентификации для JSON-ответа.
type AuthError struct {
	Error string `json:"error"`
}

// WriteAuthError пишет JSON-ответ с ошибкой аутентификации.
//
// Формат ответа: {"error": "<message>"}
// Content-Type: application/json
func WriteAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(AuthError{Error: message})
}

// WriteUnauthorizedError пишет 401 ответ.
func WriteUnauthorizedError(w http.ResponseWriter, message string) {
	WriteAuthError(w, http.StatusUnauthorized, message)
}

// WriteForbiddenError пишет 403 ответ.
func WriteForbiddenError(w http.ResponseWriter, message string) {
	WriteAuthError(w, http.StatusForbidden, message)
}
