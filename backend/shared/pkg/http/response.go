// Package http предоставляет общие HTTP-утилиты для LKFL.
//
// Включает:
//   - WriteJSON / WriteJSONError — унифицированный формат ответов
//   - ErrorResponse — структура ошибки API
//   - SuccessResponse — структура успешного ответа с метаданными
package http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse — унифицированный формат ошибки API.
// Все endpoint'ы возвращают ошибку в этом формате.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse — унифицированный формат успешного ответа.
// Используется когда нужно вернуть метаданные (пагинация, count и т.д.).
type SuccessResponse[T any] struct {
	Data       T                `json:"data"`
	Pagination *PaginationMeta  `json:"pagination,omitempty"`
	Meta       *ResponseMeta    `json:"meta,omitempty"`
}

// PaginationMeta — метаданные пагинации.
type PaginationMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int  `json:"total_pages"`
}

// ResponseMeta — дополнительные метаданные ответа.
type ResponseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

// WriteJSON отправляет JSON-ответ с указанным статусом.
// Если data == nil, тело ответа пустое.
// Content-Type всегда application/json.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteJSONError отправляет JSON-ответ с ошибкой.
// Формат: {"error": "<message>"}
func WriteJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// WriteSuccess отправляет унифицированный успешный ответ с метаданными.
func WriteSuccess[T any](w http.ResponseWriter, data T) {
	WriteJSON(w, http.StatusOK, SuccessResponse[T]{Data: data})
}

// WritePaginated отправляет ответ с данными и пагинацией.
func WritePaginated[T any](w http.ResponseWriter, data T, pagination PaginationMeta) {
	WriteJSON(w, http.StatusOK, SuccessResponse[T]{
		Data:       data,
		Pagination: &pagination,
	})
}
