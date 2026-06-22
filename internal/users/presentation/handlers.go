package presentation

import (
	"net/http"

	sharedpres "go-starter/internal/shared/presentation"
	"go-starter/internal/users/application"

	"github.com/labstack/echo/v4"
)

type UsersHandlers struct {
	getUserUseCase    *application.GetUser
	listUsersUseCase  *application.ListUsers
	updateUserUseCase *application.UpdateUser
	deleteUserUseCase *application.DeleteUser
	banUserUseCase    *application.BanUser
	unbanUserUseCase  *application.UnbanUser
}

func NewUsersHandlers(
	getUser *application.GetUser,
	listUsers *application.ListUsers,
	updateUser *application.UpdateUser,
	deleteUser *application.DeleteUser,
	banUser *application.BanUser,
	unbanUser *application.UnbanUser,
) *UsersHandlers {
	return &UsersHandlers{
		getUserUseCase:    getUser,
		listUsersUseCase:  listUsers,
		updateUserUseCase: updateUser,
		deleteUserUseCase: deleteUser,
		banUserUseCase:    banUser,
		unbanUserUseCase:  unbanUser,
	}
}

// ── User (owner) endpoints ─────────────────────────────────────────────────

func (h *UsersHandlers) GetCurrentUser(c echo.Context) error {
	user := c.Get("user").(sharedpres.JWTTokenPayload)
	result, err := h.getUserUseCase.Execute(c.Request().Context(), user.Sub)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "Current user retrieved",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) UpdateCurrentUser(c echo.Context) error {
	user := c.Get("user").(sharedpres.JWTTokenPayload)

	var dto UpdateUserDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	dto.Role = nil

	result, err := h.updateUserUseCase.Execute(c.Request().Context(), application.UpdateUserInput{
		ID:       user.Sub,
		Name:     dto.Name,
		Email:    dto.Email,
		Phone:    dto.Phone,
		Password: dto.Password,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "Current user updated",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) DeleteCurrentUser(c echo.Context) error {
	user := c.Get("user").(sharedpres.JWTTokenPayload)
	err := h.deleteUserUseCase.Execute(c.Request().Context(), user.Sub)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sharedpres.Response[any]{
		Message:    "Current user deleted",
		StatusCode: 200,
		Data:       nil,
	})
}

// ── Admin endpoints ────────────────────────────────────────────────────────

func (h *UsersHandlers) GetUser(c echo.Context) error {
	id := c.Param("id")
	result, err := h.getUserUseCase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "User retrieved",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) ListUsers(c echo.Context) error {
	query := ListUsersQuery{Page: 1, Limit: 20}
	_ = c.Bind(&query)
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	result, err := h.listUsersUseCase.Execute(c.Request().Context(), application.ListUsersInput{
		Search: query.Search,
		Page:   query.Page,
		Limit:  query.Limit,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.PaginatedResponse[application.UserOutput]{
		Message:    "Users retrieved",
		StatusCode: 200,
		Data:       result.Users,
		Pagination: sharedpres.PaginationMeta{
			Total: result.Total,
			Page:  result.Page,
			Limit: result.Limit,
		},
	})
}

func (h *UsersHandlers) UpdateUser(c echo.Context) error {
	id := c.Param("id")

	var dto UpdateUserDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.updateUserUseCase.Execute(c.Request().Context(), application.UpdateUserInput{
		ID:       id,
		Name:     dto.Name,
		Email:    dto.Email,
		Phone:    dto.Phone,
		Password: dto.Password,
		NewRole:  dto.Role,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "User updated",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	err := h.deleteUserUseCase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sharedpres.Response[any]{
		Message:    "User deleted",
		StatusCode: 200,
		Data:       nil,
	})
}

func (h *UsersHandlers) BanUser(c echo.Context) error {
	id := c.Param("id")
	result, err := h.banUserUseCase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "User banned",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) UnbanUser(c echo.Context) error {
	id := c.Param("id")
	result, err := h.unbanUserUseCase.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "User unbanned",
		StatusCode: 200,
		Data:       result,
	})
}
