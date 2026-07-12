package infrastructure

import (
	"context"
	"log/slog"

	authdomain "go-starter/internal/auth/domain"
	shareddomain "go-starter/internal/shared/domain"
	usersdomain "go-starter/internal/users/domain"
)

type AdminUserRepo interface {
	FindByEmail(ctx context.Context, email string) (*usersdomain.User, error)
	Create(ctx context.Context, user *usersdomain.User) (*usersdomain.User, error)
}

func SeedAdmin(ctx context.Context, userRepo AdminUserRepo, passwordHasher authdomain.IPasswordAdapter, adminEmail, adminPassword string) {
	email, err := shareddomain.NewEmail(adminEmail)
	if err != nil {
		slog.Error("failed to create admin email", "error", err)
		return
	}

	existing, err := userRepo.FindByEmail(ctx, email.String())
	if err == nil && existing != nil {
		slog.Info("default admin already exists, skipping seed")
		return
	}

	hashedPw, err := passwordHasher.Hash(adminPassword)
	if err != nil {
		slog.Error("failed to hash admin password", "error", err)
		return
	}

	admin := &usersdomain.User{
		ID:           shareddomain.NewId(),
		Name:         "admin",
		Email:        email,
		PasswordHash: hashedPw,
		Role:         "admin",
	}

	_, err = userRepo.Create(ctx, admin)
	if err != nil {
		slog.Error("failed to create default admin", "error", err)
		return
	}

	slog.Info("default admin account created (" + adminEmail + ")")
}
