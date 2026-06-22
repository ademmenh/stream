package auth

import (
	"github.com/labstack/echo/v4"

	"go-starter/internal/auth/application"
	"go-starter/internal/auth/presentation"
	config "go-starter/internal/config/domain"
)

type Module struct {
	handlers *presentation.AuthHandlers
}

type Dependencies struct {
	JwtAdapter      application.JwtAdapter
	PasswordAdapter application.PasswordAdapter
	IDGenerator     application.IDGenerator
	UserRepo        application.AuthUserRepo
	Config          config.IConfig
}

func NewModule(deps Dependencies) *Module {
	registerUseCase := application.NewRegister(deps.UserRepo, deps.PasswordAdapter, deps.IDGenerator, deps.JwtAdapter)
	loginUseCase := application.NewLogin(deps.UserRepo, deps.PasswordAdapter, deps.JwtAdapter)
	refreshUseCase := application.NewRefreshToken(deps.UserRepo, deps.JwtAdapter)

	handlers := presentation.NewAuthHandlers(loginUseCase, registerUseCase, refreshUseCase, deps.Config)

	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(group *echo.Group) {
	group.POST("/login", m.handlers.Login)
	group.POST("/register", m.handlers.Register)
	group.POST("/refresh", m.handlers.Refresh)
}
