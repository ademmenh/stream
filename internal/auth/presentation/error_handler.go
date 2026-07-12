package presentation

import (
	"errors"
	"net/http"

	authdomain "go-starter/internal/auth/domain"
	presentation "go-starter/internal/shared/presentation"

	"github.com/labstack/echo/v4"
)

func AuthErrorHandler(err error) *echo.HTTPError {
	var banned *authdomain.UserBannedError
	if errors.As(err, &banned) {
		return echo.NewHTTPError(http.StatusForbidden, presentation.ErrorResponse{
			Message:    "Forbidden",
			StatusCode: 403,
			Error:      err.Error(),
		})
	}

	return echo.NewHTTPError(http.StatusUnauthorized, presentation.ErrorResponse{
		Message:    "Unauthorized",
		StatusCode: 401,
		Error:      err.Error(),
	})
}
