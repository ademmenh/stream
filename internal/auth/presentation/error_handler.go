package presentation

import (
	"net/http"

	presentation "go-starter/internal/shared/presentation"

	"github.com/labstack/echo/v4"
)

func AuthErrorHandler(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusUnauthorized, presentation.ErrorResponse{
		Message:    "Unauthorized",
		StatusCode: 401,
		Error:      err.Error(),
	})
}
