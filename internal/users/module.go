package users

import (
	"github.com/labstack/echo/v4"

	sharedpres "go-starter/internal/shared/presentation"
	"go-starter/internal/users/application"
	"go-starter/internal/users/domain"
	"go-starter/internal/users/presentation"
)

type Module struct {
	handlers *presentation.UsersHandlers
}

type Dependencies struct {
	UserRepo        domain.IUserRepository
	PasswordAdapter application.PasswordAdapter
}

func NewModule(deps Dependencies) *Module {
	getUserUseCase := application.NewGetUser(deps.UserRepo)
	listUsersUseCase := application.NewListUsers(deps.UserRepo)
	updateUserUseCase := application.NewUpdateUser(deps.UserRepo, deps.PasswordAdapter)
	deleteUserUseCase := application.NewDeleteUser(deps.UserRepo)
	banUserUseCase := application.NewBanUser(deps.UserRepo)
	unbanUserUseCase := application.NewUnbanUser(deps.UserRepo)

	handlers := presentation.NewUsersHandlers(
		getUserUseCase,
		listUsersUseCase,
		updateUserUseCase,
		deleteUserUseCase,
		banUserUseCase,
		unbanUserUseCase,
	)

	return &Module{handlers: handlers}
}

func (m *Module) RegisterRoutes(group *echo.Group, jwtSecret string) {
	group.Use(sharedpres.AuthMiddleware(jwtSecret))

	group.GET("/@me", m.handlers.GetCurrentUser)
	group.PUT("/@me", m.handlers.UpdateCurrentUser)
	group.DELETE("/@me", m.handlers.DeleteCurrentUser)

	admin := group.Group("", sharedpres.RoleGuard("admin"))
	admin.GET("", m.handlers.ListUsers)
	admin.GET("/:id", m.handlers.GetUser)
	admin.PUT("/:id", m.handlers.UpdateUser)
	admin.DELETE("/:id", m.handlers.DeleteUser)
	admin.POST("/:id/ban", m.handlers.BanUser)
	admin.POST("/:id/unban", m.handlers.UnbanUser)
}
