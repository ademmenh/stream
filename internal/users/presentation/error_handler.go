package presentation

import (
	"errors"
	"net/http"

	sharedpres "go-starter/internal/shared/presentation"
	usersdomain "go-starter/internal/users/domain"

	"github.com/labstack/echo/v4"
)

func UserErrorHandler(err error) *echo.HTTPError {
	var notFound *usersdomain.UserNotFoundError
	if errors.As(err, &notFound) {
		return echo.NewHTTPError(http.StatusNotFound, sharedpres.ErrorResponse{
			Message:    "Not Found",
			StatusCode: 404,
			Error:      notFound.Error(),
		})
	}

	var conflict *usersdomain.UserEmailAlreadyExistsError
	if errors.As(err, &conflict) {
		return echo.NewHTTPError(http.StatusConflict, sharedpres.ErrorResponse{
			Message:    "Conflict",
			StatusCode: 409,
			Error:      conflict.Error(),
		})
	}

	return echo.NewHTTPError(http.StatusInternalServerError, sharedpres.ErrorResponse{
		Message:    "Internal Server Error",
		StatusCode: 500,
		Error:      err.Error(),
	})
}
