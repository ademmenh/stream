package presentation

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go-starter/internal/auth/application"
	config "go-starter/internal/config/domain"
	presentation "go-starter/internal/shared/presentation"
)

type AuthHandlers struct {
	loginUseCase    *application.Login
	registerUseCase *application.Register
	refreshUseCase  *application.RefreshToken
	config          config.IConfig
}

func NewAuthHandlers(
	loginUseCase *application.Login,
	registerUseCase *application.Register,
	refreshUseCase *application.RefreshToken,
	cfg config.IConfig,
) *AuthHandlers {
	return &AuthHandlers{
		loginUseCase:    loginUseCase,
		registerUseCase: registerUseCase,
		refreshUseCase:  refreshUseCase,
		config:          cfg,
	}
}

func (h *AuthHandlers) setAuthCookies(c echo.Context, accessToken, refreshToken string) {
	accessCookie := new(http.Cookie)
	accessCookie.Name = "access_token"
	accessCookie.Value = accessToken
	accessCookie.HttpOnly = true
	accessCookie.SameSite = http.SameSiteLaxMode
	accessCookie.Secure = h.config.CookiesSecure()
	accessCookie.Path = "/"
	c.SetCookie(accessCookie)

	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = refreshToken
	refreshCookie.HttpOnly = true
	refreshCookie.SameSite = http.SameSiteLaxMode
	refreshCookie.Secure = h.config.CookiesSecure()
	refreshCookie.Path = "/"
	c.SetCookie(refreshCookie)
}

func (h *AuthHandlers) Login(c echo.Context) error {
	var dto LoginDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.loginUseCase.Execute(c.Request().Context(), application.LoginInput{
		Email:    dto.Email,
		Password: dto.Password,
	})
	if err != nil {
		return err
	}

	h.setAuthCookies(c, result.Tokens.AccessToken, result.Tokens.RefreshToken)

	return c.JSON(http.StatusOK, presentation.AuthApiResponse[application.LoginUserOutput]{
		Message:    "Login successful",
		StatusCode: 200,
		Data:       result.User,
		Tokens: presentation.TokensData{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
		},
	})
}

func (h *AuthHandlers) Register(c echo.Context) error {
	var dto RegisterDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.registerUseCase.Execute(c.Request().Context(), application.RegisterInput{
		Name:     dto.Name,
		Email:    dto.Email,
		Password: dto.Password,
		Phone:    dto.Phone,
	})
	if err != nil {
		return err
	}

	h.setAuthCookies(c, result.Tokens.AccessToken, result.Tokens.RefreshToken)

	return c.JSON(http.StatusCreated, presentation.AuthApiResponse[application.LoginUserOutput]{
		Message:    "Registration successful",
		StatusCode: 201,
		Data:       result.User,
		Tokens: presentation.TokensData{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
		},
	})
}

func (h *AuthHandlers) Refresh(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Missing refresh token")
	}

	result, err := h.refreshUseCase.Execute(c.Request().Context(), application.RefreshTokenInput{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		return err
	}

	h.setAuthCookies(c, result.AccessToken, result.RefreshToken)

	return c.JSON(http.StatusOK, presentation.ApiResponse[application.TokensOutput]{
		Message:    "Token refreshed",
		StatusCode: 200,
		Data:       *result,
	})
}
