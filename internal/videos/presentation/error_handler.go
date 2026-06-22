package presentation

import (
	"errors"
	"net/http"

	sharedpres "go-starter/internal/shared/presentation"
	videosdomain "go-starter/internal/videos/domain"

	"github.com/labstack/echo/v4"
)

func VideoErrorHandler(err error) *echo.HTTPError {
	var notFound *videosdomain.VideoNotFoundError
	if errors.As(err, &notFound) {
		return echo.NewHTTPError(http.StatusNotFound, sharedpres.ErrorResponse{
			Message:    "Not Found",
			StatusCode: 404,
			Error:      notFound.Error(),
		})
	}

	var notReady *videosdomain.VideoNotReadyError
	if errors.As(err, &notReady) {
		return echo.NewHTTPError(http.StatusConflict, sharedpres.ErrorResponse{
			Message:    "Video Not Ready",
			StatusCode: 409,
			Error:      notReady.Error(),
		})
	}

	var invalidTransition *videosdomain.VideoInvalidStatusTransitionError
	if errors.As(err, &invalidTransition) {
		return echo.NewHTTPError(http.StatusConflict, sharedpres.ErrorResponse{
			Message:    "Invalid Status Transition",
			StatusCode: 409,
			Error:      invalidTransition.Error(),
		})
	}

	var invalidType *videosdomain.VideoInvalidTypeError
	if errors.As(err, &invalidType) {
		return echo.NewHTTPError(http.StatusBadRequest, sharedpres.ErrorResponse{
			Message:    "Bad Request",
			StatusCode: 400,
			Error:      invalidType.Error(),
		})
	}

	var rawNotFound *videosdomain.VideoRawNotFoundError
	if errors.As(err, &rawNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, sharedpres.ErrorResponse{
			Message:    "Raw File Not Found",
			StatusCode: 404,
			Error:      rawNotFound.Error(),
		})
	}

	return echo.NewHTTPError(http.StatusInternalServerError, sharedpres.ErrorResponse{
		Message:    "Internal Server Error",
		StatusCode: 500,
		Error:      err.Error(),
	})
}
