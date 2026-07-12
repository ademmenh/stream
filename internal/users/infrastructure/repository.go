package infrastructure

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"

	ent "go-starter/internal/shared/infrastructure/ent/generated"
	"go-starter/internal/shared/infrastructure/ent/generated/userschema"
	shareddomain "go-starter/internal/shared/domain"
	domain "go-starter/internal/users/domain"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	var phoneStr *string
	if u.Phone != nil {
		p := u.Phone.String()
		phoneStr = &p
	}
	created, err := r.client.UserSchema.Create().
		SetID(uuid.MustParse(u.ID.String())).
		SetName(u.Name).
		SetEmail(u.Email.String()).
		SetNillablePhone(phoneStr).
		SetPasswordHash(u.PasswordHash).
		SetRole(u.Role).
		SetBanned(u.Banned).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDomain(created), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := r.client.UserSchema.Query().
		Where(userschema.IDEQ(uuid.MustParse(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return toDomain(u), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.client.UserSchema.Query().
		Where(userschema.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return toDomain(u), nil
}

func (r *UserRepository) List(ctx context.Context, filter domain.UserListFilter) ([]*domain.User, int, error) {

	query := r.client.UserSchema.Query()
	if filter.Search != "" {
		query = query.Where(
			userschema.Or(
				userschema.NameContainsFold(filter.Search),
				userschema.EmailContainsFold(filter.Search),
			),
		)
	}
	if filter.Role != "" {
		query = query.Where(userschema.RoleEQ(filter.Role))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	users, err := query.
		Order(ent.Asc(userschema.FieldName)).
		Limit(filter.Limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	result := make([]*domain.User, len(users))
	for i, u := range users {
		result[i] = toDomain(u)
	}
	return result, total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	var phoneStr *string
	if u.Phone != nil {
		p := u.Phone.String()
		phoneStr = &p
	}
	updated, err := r.client.UserSchema.UpdateOneID(uuid.MustParse(u.ID.String())).
		SetName(u.Name).
		SetEmail(u.Email.String()).
		SetNillablePhone(phoneStr).
		SetPasswordHash(u.PasswordHash).
		SetRole(u.Role).
		SetBanned(u.Banned).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return toDomain(updated), nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	err := r.client.UserSchema.DeleteOneID(uuid.MustParse(id)).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) Ban(ctx context.Context, id string) (*domain.User, error) {
	updated, err := r.client.UserSchema.UpdateOneID(uuid.MustParse(id)).
		SetBanned(true).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("ban user: %w", err)
	}
	return toDomain(updated), nil
}

func (r *UserRepository) Unban(ctx context.Context, id string) (*domain.User, error) {
	updated, err := r.client.UserSchema.UpdateOneID(uuid.MustParse(id)).
		SetBanned(false).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unban user: %w", err)
	}
	return toDomain(updated), nil
}

// ── InMemoryUserRepository ─────────────────────────────────────────────────

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.users {
		if existing.Email.String() == u.Email.String() {
			return nil, fmt.Errorf("email already exists: %s", u.Email.String())
		}
	}

	clone := cloneUser(u)
	if clone.ID.String() == "" {
		clone.ID = shareddomain.IdFromStr(uuid.New().String())
	}
	r.users[clone.ID.String()] = clone
	return cloneUser(clone), nil
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return cloneUser(u), nil
}

func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email.String() == email {
			return cloneUser(u), nil
		}
	}
	return nil, nil
}

func (r *InMemoryUserRepository) List(ctx context.Context, filter domain.UserListFilter) ([]*domain.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*domain.User
	for _, u := range r.users {
		if filter.Search != "" {
			search := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(u.Name), search) &&
				!strings.Contains(strings.ToLower(u.Email.String()), search) {
				continue
			}
		}
		if filter.Role != "" && u.Role != filter.Role {
			continue
		}
		filtered = append(filtered, u)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	total := len(filtered)

	offset := (filter.Page - 1) * filter.Limit
	if offset > len(filtered) {
		return []*domain.User{}, total, nil
	}
	end := offset + filter.Limit
	if end > len(filtered) {
		end = len(filtered)
	}

	result := make([]*domain.User, end-offset)
	for i, u := range filtered[offset:end] {
		result[i] = cloneUser(u)
	}
	return result, total, nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[u.ID.String()]; !ok {
		return nil, nil
	}

	clone := cloneUser(u)
	r.users[clone.ID.String()] = clone
	return cloneUser(clone), nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return nil
	}
	delete(r.users, id)
	return nil
}

func (r *InMemoryUserRepository) Ban(ctx context.Context, id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	u.Banned = true
	return cloneUser(u), nil
}

func (r *InMemoryUserRepository) Unban(ctx context.Context, id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	u.Banned = false
	return cloneUser(u), nil
}

func cloneUser(u *domain.User) *domain.User {
	if u == nil {
		return nil
	}
	c := &domain.User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Banned:       u.Banned,
	}
	if u.Phone != nil {
		p := *u.Phone
		c.Phone = &p
	}
	return c
}
