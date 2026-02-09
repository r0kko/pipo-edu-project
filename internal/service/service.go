package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"pipo-edu-project/internal/auth"
	repo "pipo-edu-project/internal/repository/sqlc"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrBlocked      = errors.New("blocked")
	ErrInvalidInput = errors.New("invalid input")
	ErrInvalidRange = errors.New("invalid date range")
	ErrInvalidPlate = errors.New("invalid plate number")
)

type Service struct {
	q ServiceStore
}

func New(q ServiceStore) *Service {
	return &Service{q: q}
}

type UserCreateInput struct {
	Email      string
	Password   string
	Role       string
	FullName   string
	PlotNumber string
	ActorID    uuid.UUID
}

type UserUpdateInput struct {
	ID         uuid.UUID
	Email      string
	Role       string
	FullName   string
	PlotNumber string
	ActorID    uuid.UUID
}

type UserPasswordInput struct {
	ID       uuid.UUID
	Password string
	ActorID  uuid.UUID
}

type PassCreateInput struct {
	OwnerID      uuid.UUID
	PlateNumber  string
	VehicleBrand sql.NullString
	VehicleColor sql.NullString
	Status       string
	ActorID      uuid.UUID
}

type PassUpdateInput struct {
	ID           uuid.UUID
	PlateNumber  string
	VehicleBrand sql.NullString
	VehicleColor sql.NullString
	Status       string
	ActorID      uuid.UUID
}

type GuestCreateInput struct {
	ResidentID  uuid.UUID
	GuestName   string
	PlateNumber string
	ValidFrom   time.Time
	ValidTo     time.Time
	Status      string
	ActorID     uuid.UUID
}

type GuestUpdateInput struct {
	ID          uuid.UUID
	GuestName   string
	PlateNumber string
	ValidFrom   time.Time
	ValidTo     time.Time
	Status      string
	ActorID     uuid.UUID
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (repo.User, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		return repo.User{}, err
	}
	ok, err := auth.VerifyPassword(user.PasswordHash, password)
	if err != nil {
		return repo.User{}, err
	}
	if !ok {
		return repo.User{}, ErrForbidden
	}
	if user.BlockedAt.Valid {
		return repo.User{}, ErrBlocked
	}
	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, input UserCreateInput, passwordHash string) (repo.User, error) {
	if input.Email == "" || input.Password == "" || input.Role == "" || input.FullName == "" {
		return repo.User{}, ErrInvalidInput
	}
	if err := ValidateRole(input.Role); err != nil {
		return repo.User{}, err
	}
	plot := strings.TrimSpace(input.PlotNumber)
	if input.Role == "resident" && plot == "" {
		return repo.User{}, ErrInvalidInput
	}
	user, err := s.q.CreateUser(ctx, repo.CreateUserParams{
		Email:        input.Email,
		PasswordHash: passwordHash,
		Role:         input.Role,
		FullName:     input.FullName,
		PlotNumber:   sql.NullString{String: plot, Valid: plot != ""},
		CreatedBy:    uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
		UpdatedBy:    uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		return repo.User{}, err
	}
	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (repo.User, error) {
	user, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		return repo.User{}, err
	}
	return user, nil
}

func (s *Service) GetUserAny(ctx context.Context, id uuid.UUID) (repo.User, error) {
	user, err := s.q.GetUserByIDAny(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		return repo.User{}, err
	}
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.User, error) {
	return s.q.ListUsers(ctx, repo.ListUsersParams{
		Column1: includeDeleted,
		Limit:   limit,
		Offset:  offset,
	})
}

func (s *Service) UpdateUser(ctx context.Context, input UserUpdateInput) (repo.User, error) {
	if input.Email == "" || input.Role == "" || input.FullName == "" {
		return repo.User{}, ErrInvalidInput
	}
	if err := ValidateRole(input.Role); err != nil {
		return repo.User{}, err
	}
	plot := strings.TrimSpace(input.PlotNumber)
	if input.Role == "resident" && plot == "" {
		return repo.User{}, ErrInvalidInput
	}
	user, err := s.q.UpdateUser(ctx, repo.UpdateUserParams{
		ID:         input.ID,
		Email:      input.Email,
		Role:       input.Role,
		FullName:   input.FullName,
		PlotNumber: sql.NullString{String: plot, Valid: plot != ""},
		UpdatedBy: uuid.NullUUID{UUID: input.ActorID,
			Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		return repo.User{}, err
	}
	return user, nil
}

func (s *Service) UpdateUserPassword(ctx context.Context, input UserPasswordInput, passwordHash string) (repo.User, error) {
	if input.Password == "" {
		return repo.User{}, ErrInvalidInput
	}
	user, err := s.q.UpdateUserPassword(ctx, repo.UpdateUserPasswordParams{
		ID:           input.ID,
		PasswordHash: passwordHash,
		UpdatedBy:    uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.User{}, ErrNotFound
		}
		return repo.User{}, err
	}
	return user, nil
}

func (s *Service) SoftDeleteUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.SoftDeleteUser(ctx, repo.SoftDeleteUserParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) RestoreUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.RestoreUser(ctx, repo.RestoreUserParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) BlockUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.BlockUser(ctx, repo.BlockUserParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) UnblockUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.UnblockUser(ctx, repo.UnblockUserParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) CreatePass(ctx context.Context, input PassCreateInput) (repo.Pass, error) {
	if err := ValidatePlate(input.PlateNumber); err != nil {
		return repo.Pass{}, err
	}
	pass, err := s.q.CreatePass(ctx, repo.CreatePassParams{
		OwnerUserID:  input.OwnerID,
		PlateNumber:  NormalizePlate(input.PlateNumber),
		VehicleBrand: input.VehicleBrand,
		VehicleColor: input.VehicleColor,
		Status:       input.Status,
		CreatedBy:    uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
		UpdatedBy:    uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		return repo.Pass{}, err
	}
	return pass, nil
}

func (s *Service) GetPass(ctx context.Context, id uuid.UUID) (repo.Pass, error) {
	pass, err := s.q.GetPassByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.Pass{}, ErrNotFound
		}
		return repo.Pass{}, err
	}
	return pass, nil
}

func (s *Service) GetPassAny(ctx context.Context, id uuid.UUID) (repo.Pass, error) {
	pass, err := s.q.GetPassByIDAny(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.Pass{}, ErrNotFound
		}
		return repo.Pass{}, err
	}
	return pass, nil
}

func (s *Service) ListPassesByOwner(ctx context.Context, owner uuid.UUID, includeDeleted bool, limit, offset int32) ([]repo.Pass, error) {
	return s.q.ListPassesByOwner(ctx, repo.ListPassesByOwnerParams{
		OwnerUserID: owner,
		Column2:     includeDeleted,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *Service) ListPasses(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.Pass, error) {
	return s.q.ListPasses(ctx, repo.ListPassesParams{Column1: includeDeleted, Limit: limit, Offset: offset})
}

func (s *Service) SearchPasses(ctx context.Context, plate string, limit, offset int32) ([]repo.Pass, error) {
	if err := ValidatePlate(plate); err != nil {
		return nil, err
	}
	pattern := "%" + NormalizePlate(plate) + "%"
	return s.q.SearchPassesByPlate(ctx, repo.SearchPassesByPlateParams{
		PlateNumber: pattern,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *Service) UpdatePass(ctx context.Context, input PassUpdateInput) (repo.Pass, error) {
	if err := ValidatePlate(input.PlateNumber); err != nil {
		return repo.Pass{}, err
	}
	pass, err := s.q.UpdatePass(ctx, repo.UpdatePassParams{
		ID:           input.ID,
		PlateNumber:  NormalizePlate(input.PlateNumber),
		VehicleBrand: input.VehicleBrand,
		VehicleColor: input.VehicleColor,
		Status:       input.Status,
		UpdatedBy:    uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.Pass{}, ErrNotFound
		}
		return repo.Pass{}, err
	}
	return pass, nil
}

func (s *Service) SoftDeletePass(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.SoftDeletePass(ctx, repo.SoftDeletePassParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) RestorePass(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.RestorePass(ctx, repo.RestorePassParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) CreateGuestRequest(ctx context.Context, input GuestCreateInput) (repo.GuestRequest, error) {
	if err := ValidatePlate(input.PlateNumber); err != nil {
		return repo.GuestRequest{}, err
	}
	if input.ValidFrom.After(input.ValidTo) {
		return repo.GuestRequest{}, ErrInvalidRange
	}
	guest, err := s.q.CreateGuestRequest(ctx, repo.CreateGuestRequestParams{
		ResidentUserID: input.ResidentID,
		GuestFullName:  input.GuestName,
		PlateNumber:    NormalizePlate(input.PlateNumber),
		ValidFrom:      input.ValidFrom,
		ValidTo:        input.ValidTo,
		Status:         input.Status,
		CreatedBy:      uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
		UpdatedBy:      uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		return repo.GuestRequest{}, err
	}
	return guest, nil
}

func (s *Service) GetGuestRequest(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error) {
	guest, err := s.q.GetGuestRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.GuestRequest{}, ErrNotFound
		}
		return repo.GuestRequest{}, err
	}
	return guest, nil
}

func (s *Service) GetGuestRequestAny(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error) {
	guest, err := s.q.GetGuestRequestByIDAny(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.GuestRequest{}, ErrNotFound
		}
		return repo.GuestRequest{}, err
	}
	return guest, nil
}

func (s *Service) ListGuestRequestsByResident(ctx context.Context, resident uuid.UUID, includeDeleted bool, limit, offset int32) ([]repo.GuestRequest, error) {
	return s.q.ListGuestRequestsByResident(ctx, repo.ListGuestRequestsByResidentParams{
		ResidentUserID: resident,
		Column2:        includeDeleted,
		Limit:          limit,
		Offset:         offset,
	})
}

func (s *Service) ListGuestRequests(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.GuestRequest, error) {
	return s.q.ListGuestRequests(ctx, repo.ListGuestRequestsParams{Column1: includeDeleted, Limit: limit, Offset: offset})
}

func (s *Service) UpdateGuestRequest(ctx context.Context, input GuestUpdateInput) (repo.GuestRequest, error) {
	if err := ValidatePlate(input.PlateNumber); err != nil {
		return repo.GuestRequest{}, err
	}
	if input.ValidFrom.After(input.ValidTo) {
		return repo.GuestRequest{}, ErrInvalidRange
	}
	guest, err := s.q.UpdateGuestRequest(ctx, repo.UpdateGuestRequestParams{
		ID:            input.ID,
		GuestFullName: input.GuestName,
		PlateNumber:   NormalizePlate(input.PlateNumber),
		ValidFrom:     input.ValidFrom,
		ValidTo:       input.ValidTo,
		Status:        input.Status,
		UpdatedBy:     uuid.NullUUID{UUID: input.ActorID, Valid: input.ActorID != uuid.Nil},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.GuestRequest{}, ErrNotFound
		}
		return repo.GuestRequest{}, err
	}
	return guest, nil
}

func (s *Service) SoftDeleteGuestRequest(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.SoftDeleteGuestRequest(ctx, repo.SoftDeleteGuestRequestParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) RestoreGuestRequest(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return s.q.RestoreGuestRequest(ctx, repo.RestoreGuestRequestParams{ID: id, UpdatedBy: uuid.NullUUID{UUID: actor, Valid: actor != uuid.Nil}})
}

func (s *Service) CreateEntryLog(ctx context.Context, passID, guardID uuid.UUID, action string, comment sql.NullString) (repo.EntryLog, error) {
	return s.q.CreateEntryLog(ctx, repo.CreateEntryLogParams{PassID: passID, GuardUserID: guardID, Action: action, Comment: comment})
}

func (s *Service) ListEntryLogs(ctx context.Context, passID uuid.UUID, limit, offset int32) ([]repo.EntryLog, error) {
	return s.q.ListEntryLogsByPass(ctx, repo.ListEntryLogsByPassParams{PassID: passID, Limit: limit, Offset: offset})
}
