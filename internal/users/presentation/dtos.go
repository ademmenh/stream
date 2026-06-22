package presentation

type GetUserDto struct{}

type UpdateUserDto struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

type ListUsersQuery struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}
