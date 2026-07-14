package presentation

import (
	"net/http"

	sharedpres "go-starter/internal/shared/presentation"
	"go-starter/internal/users/application"

	"github.com/labstack/echo/v4"
)

type UsersHandlers struct {
	getUserUseCase                  *application.GetUser
	listUsersUseCase                *application.ListUsers
	updateCurrentUserUseCase        *application.UpdateCurrentUser
	updateUserUseCase               *application.UpdateUser
	deleteUserUseCase               *application.DeleteUser
	banUserUseCase                  *application.BanUser
	unbanUserUseCase                *application.UnbanUser
	getProfileImageUploadUrlUseCase *application.GetProfileImageUploadUrl
	updateProfileImageUseCase       *application.UpdateProfileImage
	resetPasswordUseCase            *application.ResetPassword
}

func NewUsersHandlers(
	getUser *application.GetUser,
	listUsers *application.ListUsers,
	updateCurrentUser *application.UpdateCurrentUser,
	updateUser *application.UpdateUser,
	deleteUser *application.DeleteUser,
	banUser *application.BanUser,
	unbanUser *application.UnbanUser,
	getProfileImageUploadUrl *application.GetProfileImageUploadUrl,
	updateProfileImage *application.UpdateProfileImage,
	resetPassword *application.ResetPassword,
) *UsersHandlers {
	return &UsersHandlers{
		getUserUseCase:                  getUser,
		listUsersUseCase:                listUsers,
		updateCurrentUserUseCase:        updateCurrentUser,
		updateUserUseCase:               updateUser,
		deleteUserUseCase:               deleteUser,
		banUserUseCase:                  banUser,
		unbanUserUseCase:                unbanUser,
		getProfileImageUploadUrlUseCase: getProfileImageUploadUrl,
		updateProfileImageUseCase:       updateProfileImage,
		resetPasswordUseCase:            resetPassword,
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

	var dto UpdateCurrentUserDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.updateCurrentUserUseCase.Execute(c.Request().Context(), application.UpdateCurrentUserInput{
		ID:    user.Sub,
		Name:  dto.Name,
		Email: dto.Email,
		Phone: dto.Phone,
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
		Role:   query.Role,
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
		ID:      id,
		Name:    dto.Name,
		Email:   dto.Email,
		Phone:   dto.Phone,
		NewRole: dto.Role,
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

func (h *UsersHandlers) GetProfileImageUploadUrl(c echo.Context) error {
	user := c.Get("user").(sharedpres.JWTTokenPayload)

	result, err := h.getProfileImageUploadUrlUseCase.Execute(c.Request().Context(), application.GetProfileImageUploadUrlInput{
		UserID: user.Sub,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[*application.GetProfileImageUploadUrlOutput]{
		Message:    "Upload URL generated",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) UpdateProfileImage(c echo.Context) error {
	user := c.Get("user").(sharedpres.JWTTokenPayload)

	var dto UpdateProfileImageDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	result, err := h.updateProfileImageUseCase.Execute(c.Request().Context(), application.UpdateProfileImageInput{
		UserID: user.Sub,
		Key:    dto.Key,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[*application.UserOutput]{
		Message:    "Profile image updated",
		StatusCode: 200,
		Data:       result,
	})
}

func (h *UsersHandlers) ResetPassword(c echo.Context) error {
	user := c.Get("user").(sharedpres.JWTTokenPayload)

	var dto ResetPasswordDto
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	err := h.resetPasswordUseCase.Execute(c.Request().Context(), application.ResetPasswordInput{
		UserID:          user.Sub,
		OldPassword:     dto.OldPassword,
		NewPassword:     dto.NewPassword,
		ConfirmPassword: dto.ConfirmPassword,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sharedpres.Response[any]{
		Message:    "Password reset successfully",
		StatusCode: 200,
		Data:       nil,
	})
}
