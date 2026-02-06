package http

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	repo "pipo-edu-project/internal/repository/sqlc"
	"pipo-edu-project/internal/service"
)

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (repo.User, error)
}

type UserService interface {
	CreateUser(ctx context.Context, input service.UserCreateInput, passwordHash string) (repo.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (repo.User, error)
	GetUserAny(ctx context.Context, id uuid.UUID) (repo.User, error)
	ListUsers(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.User, error)
	UpdateUser(ctx context.Context, input service.UserUpdateInput) (repo.User, error)
	UpdateUserPassword(ctx context.Context, input service.UserPasswordInput, passwordHash string) (repo.User, error)
	SoftDeleteUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error
	RestoreUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error
}

type PassService interface {
	CreatePass(ctx context.Context, input service.PassCreateInput) (repo.Pass, error)
	GetPass(ctx context.Context, id uuid.UUID) (repo.Pass, error)
	GetPassAny(ctx context.Context, id uuid.UUID) (repo.Pass, error)
	ListPasses(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.Pass, error)
	ListPassesByOwner(ctx context.Context, owner uuid.UUID, includeDeleted bool, limit, offset int32) ([]repo.Pass, error)
	SearchPasses(ctx context.Context, plate string, limit, offset int32) ([]repo.Pass, error)
	UpdatePass(ctx context.Context, input service.PassUpdateInput) (repo.Pass, error)
	SoftDeletePass(ctx context.Context, id uuid.UUID, actor uuid.UUID) error
	RestorePass(ctx context.Context, id uuid.UUID, actor uuid.UUID) error
}

type GuestService interface {
	CreateGuestRequest(ctx context.Context, input service.GuestCreateInput) (repo.GuestRequest, error)
	GetGuestRequest(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error)
	GetGuestRequestAny(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error)
	ListGuestRequests(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.GuestRequest, error)
	ListGuestRequestsByResident(ctx context.Context, resident uuid.UUID, includeDeleted bool, limit, offset int32) ([]repo.GuestRequest, error)
	UpdateGuestRequest(ctx context.Context, input service.GuestUpdateInput) (repo.GuestRequest, error)
	SoftDeleteGuestRequest(ctx context.Context, id uuid.UUID, actor uuid.UUID) error
	RestoreGuestRequest(ctx context.Context, id uuid.UUID, actor uuid.UUID) error
}

type EntryService interface {
	CreateEntryLog(ctx context.Context, passID, guardID uuid.UUID, action string, comment sql.NullString) (repo.EntryLog, error)
	ListEntryLogs(ctx context.Context, passID uuid.UUID, limit, offset int32) ([]repo.EntryLog, error)
}
