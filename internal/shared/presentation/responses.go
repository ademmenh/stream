package presentation

import "fmt"

type Response[T any] struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Data       T      `json:"data"`
}

type PaginationMeta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type PaginatedResponse[T any] struct {
	Message    string         `json:"message"`
	StatusCode int            `json:"statusCode"`
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

type TokensData struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthApiResponse[T any] struct {
	Message    string     `json:"message"`
	StatusCode int        `json:"statusCode"`
	Data       T          `json:"data"`
	Tokens     TokensData `json:"tokens"`
}

type ApiResponse[T any] struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Data       T      `json:"data"`
}

type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

func NewResponse[T any](msg string, data T) Response[T] {
	return Response[T]{Message: msg, StatusCode: 200, Data: data}
}

func NewErrorResponse(status int, errType, msg string) ErrorResponse {
	return ErrorResponse{
		Message:    msg,
		StatusCode: status,
		Error:      errType,
	}
}

type AppError struct {
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}
