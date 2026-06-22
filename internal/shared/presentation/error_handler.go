package presentation

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	authdomain "go-starter/internal/auth/domain"
	usersdomain "go-starter/internal/users/domain"
	videosdomain "go-starter/internal/videos/domain"
)

var (
	AuthErrorHandler  func(error) *echo.HTTPError
	UserErrorHandler  func(error) *echo.HTTPError
	VideoErrorHandler func(error) *echo.HTTPError
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var authErr authdomain.AuthError
	var userErr usersdomain.UserError
	var videoErr videosdomain.VideoError

	switch {
	case errors.As(err, &authErr):
		if AuthErrorHandler != nil {
			httpErr := AuthErrorHandler(err)
			_ = c.JSON(httpErr.Code, httpErr.Message)
			return
		}

	case errors.As(err, &userErr):
		if UserErrorHandler != nil {
			httpErr := UserErrorHandler(err)
			_ = c.JSON(httpErr.Code, httpErr.Message)
			return
		}

	case errors.As(err, &videoErr):
		if VideoErrorHandler != nil {
			httpErr := VideoErrorHandler(err)
			_ = c.JSON(httpErr.Code, httpErr.Message)
			return
		}
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		resp := he.Message
		if _, ok := resp.(ErrorResponse); !ok {
			resp = ErrorResponse{
				Message:    "HTTP Error",
				StatusCode: he.Code,
				Error:      he.Message.(string),
			}
		}
		_ = c.JSON(he.Code, resp)
		return
	}

	slog.Error("unhandled error", "error", err)
	_ = c.JSON(http.StatusInternalServerError, ErrorResponse{
		Message:    "Internal Server Error",
		StatusCode: 500,
		Error:      "An unexpected error occurred",
	})
}
