package service

import (
	"context"

	"github.com/google/uuid"

	repo "pipo-edu-project/internal/repository/sqlc"
)

//go:generate mockgen -destination=mock_store_test.go -package=service pipo-edu-project/internal/service ServiceStore
type ServiceStore interface {
	CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error)
	GetUserByEmail(ctx context.Context, email string) (repo.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (repo.User, error)
	GetUserByIDAny(ctx context.Context, id uuid.UUID) (repo.User, error)
	ListUsers(ctx context.Context, arg repo.ListUsersParams) ([]repo.User, error)
	UpdateUser(ctx context.Context, arg repo.UpdateUserParams) (repo.User, error)
	UpdateUserPassword(ctx context.Context, arg repo.UpdateUserPasswordParams) (repo.User, error)
	RestoreUser(ctx context.Context, arg repo.RestoreUserParams) error
	BlockUser(ctx context.Context, arg repo.BlockUserParams) error
	UnblockUser(ctx context.Context, arg repo.UnblockUserParams) error
	SoftDeleteUser(ctx context.Context, arg repo.SoftDeleteUserParams) error

	CreatePass(ctx context.Context, arg repo.CreatePassParams) (repo.Pass, error)
	GetPassByID(ctx context.Context, id uuid.UUID) (repo.Pass, error)
	GetPassByIDAny(ctx context.Context, id uuid.UUID) (repo.Pass, error)
	ListPasses(ctx context.Context, arg repo.ListPassesParams) ([]repo.Pass, error)
	ListPassesByOwner(ctx context.Context, arg repo.ListPassesByOwnerParams) ([]repo.Pass, error)
	SearchPassesByPlate(ctx context.Context, arg repo.SearchPassesByPlateParams) ([]repo.Pass, error)
	UpdatePass(ctx context.Context, arg repo.UpdatePassParams) (repo.Pass, error)
	RestorePass(ctx context.Context, arg repo.RestorePassParams) error
	SoftDeletePass(ctx context.Context, arg repo.SoftDeletePassParams) error

	CreateGuestRequest(ctx context.Context, arg repo.CreateGuestRequestParams) (repo.GuestRequest, error)
	GetGuestRequestByID(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error)
	GetGuestRequestByIDAny(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error)
	ListGuestRequests(ctx context.Context, arg repo.ListGuestRequestsParams) ([]repo.GuestRequest, error)
	ListGuestRequestsByResident(ctx context.Context, arg repo.ListGuestRequestsByResidentParams) ([]repo.GuestRequest, error)
	UpdateGuestRequest(ctx context.Context, arg repo.UpdateGuestRequestParams) (repo.GuestRequest, error)
	RestoreGuestRequest(ctx context.Context, arg repo.RestoreGuestRequestParams) error
	SoftDeleteGuestRequest(ctx context.Context, arg repo.SoftDeleteGuestRequestParams) error

	CreateEntryLog(ctx context.Context, arg repo.CreateEntryLogParams) (repo.EntryLog, error)
	ListEntryLogsByPass(ctx context.Context, arg repo.ListEntryLogsByPassParams) ([]repo.EntryLog, error)
}
