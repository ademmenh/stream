package presentation

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func extractToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:], nil
	}

	return "", echo.NewHTTPError(http.StatusUnauthorized, "Missing authentication token")
}

type JWTTokenPayload struct {
	Sub   string
	Email string
	Role  string
}

func AuthMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenStr, err := extractToken(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing authentication token")
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			sub, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			role, _ := claims["role"].(string)
			c.Set("user", JWTTokenPayload{
				Sub:   sub,
				Email: email,
				Role:  role,
			})

			return next(c)
		}
	}
}

func RoleGuard(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(JWTTokenPayload)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}
			for _, r := range roles {
				if user.Role == r {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}
