package presentation

type GetUserDto struct{}

type UpdateCurrentUserDto struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

type UpdateUserDto struct {
	Name    *string `json:"name"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	Role    *string `json:"role"`
}

type ListUsersQuery struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Role   string `query:"role"`
}

type UpdateProfileImageDto struct {
	Key string `json:"key"`
}

type ResetPasswordDto struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}
