package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pipo-edu-project/internal/auth"
	repo "pipo-edu-project/internal/repository/sqlc"
)

var errMockUnimplemented = errors.New("mock method not implemented")

type mockStore struct {
	createUserFn                  func(context.Context, repo.CreateUserParams) (repo.User, error)
	getUserByEmailFn              func(context.Context, string) (repo.User, error)
	getUserByIDFn                 func(context.Context, uuid.UUID) (repo.User, error)
	getUserByIDAnyFn              func(context.Context, uuid.UUID) (repo.User, error)
	listUsersFn                   func(context.Context, repo.ListUsersParams) ([]repo.User, error)
	updateUserFn                  func(context.Context, repo.UpdateUserParams) (repo.User, error)
	updateUserPasswordFn          func(context.Context, repo.UpdateUserPasswordParams) (repo.User, error)
	restoreUserFn                 func(context.Context, repo.RestoreUserParams) error
	blockUserFn                   func(context.Context, repo.BlockUserParams) error
	unblockUserFn                 func(context.Context, repo.UnblockUserParams) error
	softDeleteUserFn              func(context.Context, repo.SoftDeleteUserParams) error
	createPassFn                  func(context.Context, repo.CreatePassParams) (repo.Pass, error)
	getPassByIDFn                 func(context.Context, uuid.UUID) (repo.Pass, error)
	getPassByIDAnyFn              func(context.Context, uuid.UUID) (repo.Pass, error)
	listPassesFn                  func(context.Context, repo.ListPassesParams) ([]repo.Pass, error)
	listPassesByOwnerFn           func(context.Context, repo.ListPassesByOwnerParams) ([]repo.Pass, error)
	searchPassesByPlateFn         func(context.Context, repo.SearchPassesByPlateParams) ([]repo.Pass, error)
	updatePassFn                  func(context.Context, repo.UpdatePassParams) (repo.Pass, error)
	restorePassFn                 func(context.Context, repo.RestorePassParams) error
	softDeletePassFn              func(context.Context, repo.SoftDeletePassParams) error
	createGuestRequestFn          func(context.Context, repo.CreateGuestRequestParams) (repo.GuestRequest, error)
	getGuestRequestByIDFn         func(context.Context, uuid.UUID) (repo.GuestRequest, error)
	getGuestRequestByIDAnyFn      func(context.Context, uuid.UUID) (repo.GuestRequest, error)
	listGuestRequestsFn           func(context.Context, repo.ListGuestRequestsParams) ([]repo.GuestRequest, error)
	listGuestRequestsByResidentFn func(context.Context, repo.ListGuestRequestsByResidentParams) ([]repo.GuestRequest, error)
	updateGuestRequestFn          func(context.Context, repo.UpdateGuestRequestParams) (repo.GuestRequest, error)
	restoreGuestRequestFn         func(context.Context, repo.RestoreGuestRequestParams) error
	softDeleteGuestRequestFn      func(context.Context, repo.SoftDeleteGuestRequestParams) error
	createEntryLogFn              func(context.Context, repo.CreateEntryLogParams) (repo.EntryLog, error)
	listEntryLogsByPassFn         func(context.Context, repo.ListEntryLogsByPassParams) ([]repo.EntryLog, error)
}

func (m *mockStore) CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error) {
	if m.createUserFn == nil {
		return repo.User{}, errMockUnimplemented
	}
	return m.createUserFn(ctx, arg)
}
func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	if m.getUserByEmailFn == nil {
		return repo.User{}, errMockUnimplemented
	}
	return m.getUserByEmailFn(ctx, email)
}
func (m *mockStore) GetUserByID(ctx context.Context, id uuid.UUID) (repo.User, error) {
	if m.getUserByIDFn == nil {
		return repo.User{}, errMockUnimplemented
	}
	return m.getUserByIDFn(ctx, id)
}
func (m *mockStore) GetUserByIDAny(ctx context.Context, id uuid.UUID) (repo.User, error) {
	if m.getUserByIDAnyFn == nil {
		return repo.User{}, errMockUnimplemented
	}
	return m.getUserByIDAnyFn(ctx, id)
}
func (m *mockStore) ListUsers(ctx context.Context, arg repo.ListUsersParams) ([]repo.User, error) {
	if m.listUsersFn == nil {
		return nil, errMockUnimplemented
	}
	return m.listUsersFn(ctx, arg)
}
func (m *mockStore) UpdateUser(ctx context.Context, arg repo.UpdateUserParams) (repo.User, error) {
	if m.updateUserFn == nil {
		return repo.User{}, errMockUnimplemented
	}
	return m.updateUserFn(ctx, arg)
}
func (m *mockStore) UpdateUserPassword(ctx context.Context, arg repo.UpdateUserPasswordParams) (repo.User, error) {
	if m.updateUserPasswordFn == nil {
		return repo.User{}, errMockUnimplemented
	}
	return m.updateUserPasswordFn(ctx, arg)
}
func (m *mockStore) RestoreUser(ctx context.Context, arg repo.RestoreUserParams) error {
	if m.restoreUserFn == nil {
		return errMockUnimplemented
	}
	return m.restoreUserFn(ctx, arg)
}
func (m *mockStore) BlockUser(ctx context.Context, arg repo.BlockUserParams) error {
	if m.blockUserFn == nil {
		return errMockUnimplemented
	}
	return m.blockUserFn(ctx, arg)
}
func (m *mockStore) UnblockUser(ctx context.Context, arg repo.UnblockUserParams) error {
	if m.unblockUserFn == nil {
		return errMockUnimplemented
	}
	return m.unblockUserFn(ctx, arg)
}
func (m *mockStore) SoftDeleteUser(ctx context.Context, arg repo.SoftDeleteUserParams) error {
	if m.softDeleteUserFn == nil {
		return errMockUnimplemented
	}
	return m.softDeleteUserFn(ctx, arg)
}
func (m *mockStore) CreatePass(ctx context.Context, arg repo.CreatePassParams) (repo.Pass, error) {
	if m.createPassFn == nil {
		return repo.Pass{}, errMockUnimplemented
	}
	return m.createPassFn(ctx, arg)
}
func (m *mockStore) GetPassByID(ctx context.Context, id uuid.UUID) (repo.Pass, error) {
	if m.getPassByIDFn == nil {
		return repo.Pass{}, errMockUnimplemented
	}
	return m.getPassByIDFn(ctx, id)
}
func (m *mockStore) GetPassByIDAny(ctx context.Context, id uuid.UUID) (repo.Pass, error) {
	if m.getPassByIDAnyFn == nil {
		return repo.Pass{}, errMockUnimplemented
	}
	return m.getPassByIDAnyFn(ctx, id)
}
func (m *mockStore) ListPasses(ctx context.Context, arg repo.ListPassesParams) ([]repo.Pass, error) {
	if m.listPassesFn == nil {
		return nil, errMockUnimplemented
	}
	return m.listPassesFn(ctx, arg)
}
func (m *mockStore) ListPassesByOwner(ctx context.Context, arg repo.ListPassesByOwnerParams) ([]repo.Pass, error) {
	if m.listPassesByOwnerFn == nil {
		return nil, errMockUnimplemented
	}
	return m.listPassesByOwnerFn(ctx, arg)
}
func (m *mockStore) SearchPassesByPlate(ctx context.Context, arg repo.SearchPassesByPlateParams) ([]repo.Pass, error) {
	if m.searchPassesByPlateFn == nil {
		return nil, errMockUnimplemented
	}
	return m.searchPassesByPlateFn(ctx, arg)
}
func (m *mockStore) UpdatePass(ctx context.Context, arg repo.UpdatePassParams) (repo.Pass, error) {
	if m.updatePassFn == nil {
		return repo.Pass{}, errMockUnimplemented
	}
	return m.updatePassFn(ctx, arg)
}
func (m *mockStore) RestorePass(ctx context.Context, arg repo.RestorePassParams) error {
	if m.restorePassFn == nil {
		return errMockUnimplemented
	}
	return m.restorePassFn(ctx, arg)
}
func (m *mockStore) SoftDeletePass(ctx context.Context, arg repo.SoftDeletePassParams) error {
	if m.softDeletePassFn == nil {
		return errMockUnimplemented
	}
	return m.softDeletePassFn(ctx, arg)
}
func (m *mockStore) CreateGuestRequest(ctx context.Context, arg repo.CreateGuestRequestParams) (repo.GuestRequest, error) {
	if m.createGuestRequestFn == nil {
		return repo.GuestRequest{}, errMockUnimplemented
	}
	return m.createGuestRequestFn(ctx, arg)
}
func (m *mockStore) GetGuestRequestByID(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error) {
	if m.getGuestRequestByIDFn == nil {
		return repo.GuestRequest{}, errMockUnimplemented
	}
	return m.getGuestRequestByIDFn(ctx, id)
}
func (m *mockStore) GetGuestRequestByIDAny(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error) {
	if m.getGuestRequestByIDAnyFn == nil {
		return repo.GuestRequest{}, errMockUnimplemented
	}
	return m.getGuestRequestByIDAnyFn(ctx, id)
}
func (m *mockStore) ListGuestRequests(ctx context.Context, arg repo.ListGuestRequestsParams) ([]repo.GuestRequest, error) {
	if m.listGuestRequestsFn == nil {
		return nil, errMockUnimplemented
	}
	return m.listGuestRequestsFn(ctx, arg)
}
func (m *mockStore) ListGuestRequestsByResident(ctx context.Context, arg repo.ListGuestRequestsByResidentParams) ([]repo.GuestRequest, error) {
	if m.listGuestRequestsByResidentFn == nil {
		return nil, errMockUnimplemented
	}
	return m.listGuestRequestsByResidentFn(ctx, arg)
}
func (m *mockStore) UpdateGuestRequest(ctx context.Context, arg repo.UpdateGuestRequestParams) (repo.GuestRequest, error) {
	if m.updateGuestRequestFn == nil {
		return repo.GuestRequest{}, errMockUnimplemented
	}
	return m.updateGuestRequestFn(ctx, arg)
}
func (m *mockStore) RestoreGuestRequest(ctx context.Context, arg repo.RestoreGuestRequestParams) error {
	if m.restoreGuestRequestFn == nil {
		return errMockUnimplemented
	}
	return m.restoreGuestRequestFn(ctx, arg)
}
func (m *mockStore) SoftDeleteGuestRequest(ctx context.Context, arg repo.SoftDeleteGuestRequestParams) error {
	if m.softDeleteGuestRequestFn == nil {
		return errMockUnimplemented
	}
	return m.softDeleteGuestRequestFn(ctx, arg)
}
func (m *mockStore) CreateEntryLog(ctx context.Context, arg repo.CreateEntryLogParams) (repo.EntryLog, error) {
	if m.createEntryLogFn == nil {
		return repo.EntryLog{}, errMockUnimplemented
	}
	return m.createEntryLogFn(ctx, arg)
}
func (m *mockStore) ListEntryLogsByPass(ctx context.Context, arg repo.ListEntryLogsByPassParams) ([]repo.EntryLog, error) {
	if m.listEntryLogsByPassFn == nil {
		return nil, errMockUnimplemented
	}
	return m.listEntryLogsByPassFn(ctx, arg)
}

func TestServiceUnit_Authenticate(t *testing.T) {
	ctx := context.Background()
	passwordHash, err := auth.HashPassword("secret123")
	require.NoError(t, err)
	user := repo.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		PasswordHash: passwordHash,
		Role:         "admin",
	}

	tests := []struct {
		name    string
		store   *mockStore
		pass    string
		wantErr error
	}{
		{
			name: "not found",
			store: &mockStore{
				getUserByEmailFn: func(context.Context, string) (repo.User, error) { return repo.User{}, sql.ErrNoRows },
			},
			pass:    "secret123",
			wantErr: ErrNotFound,
		},
		{
			name: "wrong password",
			store: &mockStore{
				getUserByEmailFn: func(context.Context, string) (repo.User, error) { return user, nil },
			},
			pass:    "bad-password",
			wantErr: ErrForbidden,
		},
		{
			name: "blocked",
			store: &mockStore{
				getUserByEmailFn: func(context.Context, string) (repo.User, error) {
					blocked := user
					blocked.BlockedAt = sql.NullTime{Time: time.Now(), Valid: true}
					return blocked, nil
				},
			},
			pass:    "secret123",
			wantErr: ErrBlocked,
		},
		{
			name: "success",
			store: &mockStore{
				getUserByEmailFn: func(context.Context, string) (repo.User, error) { return user, nil },
			},
			pass: "secret123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := New(tc.store)
			got, err := svc.Authenticate(ctx, user.Email, tc.pass)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, user.ID, got.ID)
		})
	}
}

func TestServiceUnit_CreateUser(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()

	t.Run("validation failure", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.CreateUser(ctx, UserCreateInput{}, "hash")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("resident requires plot", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.CreateUser(ctx, UserCreateInput{
			Email:    "resident@example.com",
			Password: "secret",
			Role:     "resident",
			FullName: "Resident",
		}, "hash")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("success sets audit and plot", func(t *testing.T) {
		var arg repo.CreateUserParams
		svc := New(&mockStore{
			createUserFn: func(_ context.Context, in repo.CreateUserParams) (repo.User, error) {
				arg = in
				return repo.User{ID: uuid.New(), Email: in.Email, Role: in.Role, FullName: in.FullName, PlotNumber: in.PlotNumber}, nil
			},
		})
		got, err := svc.CreateUser(ctx, UserCreateInput{
			Email:      "resident@example.com",
			Password:   "secret",
			Role:       "resident",
			FullName:   "Resident",
			PlotNumber: "  12A  ",
			ActorID:    actorID,
		}, "hash")
		require.NoError(t, err)
		require.Equal(t, "12A", arg.PlotNumber.String)
		require.True(t, arg.CreatedBy.Valid)
		require.Equal(t, actorID, arg.CreatedBy.UUID)
		require.Equal(t, got.Email, "resident@example.com")
	})
}

func TestServiceUnit_GetUserAndAny(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	svc := New(&mockStore{
		getUserByIDFn: func(context.Context, uuid.UUID) (repo.User, error) { return repo.User{}, sql.ErrNoRows },
		getUserByIDAnyFn: func(context.Context, uuid.UUID) (repo.User, error) {
			return repo.User{ID: userID, Email: "a@b.c"}, nil
		},
	})
	_, err := svc.GetUser(ctx, userID)
	require.ErrorIs(t, err, ErrNotFound)
	user, err := svc.GetUserAny(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, userID, user.ID)
}

func TestServiceUnit_ListUsers(t *testing.T) {
	ctx := context.Background()
	svc := New(&mockStore{
		listUsersFn: func(_ context.Context, arg repo.ListUsersParams) ([]repo.User, error) {
			require.True(t, arg.Column1)
			require.Equal(t, int32(25), arg.Limit)
			require.Equal(t, int32(2), arg.Offset)
			return []repo.User{{Email: "user@example.com"}}, nil
		},
	})
	users, err := svc.ListUsers(ctx, true, 25, 2)
	require.NoError(t, err)
	require.Len(t, users, 1)
}

func TestServiceUnit_UpdateUserAndPassword(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()

	t.Run("update user invalid", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.UpdateUser(ctx, UserUpdateInput{})
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("update user success", func(t *testing.T) {
		svc := New(&mockStore{
			updateUserFn: func(_ context.Context, arg repo.UpdateUserParams) (repo.User, error) {
				require.Equal(t, userID, arg.ID)
				require.Equal(t, "resident", arg.Role)
				require.Equal(t, actorID, arg.UpdatedBy.UUID)
				return repo.User{ID: arg.ID, Email: arg.Email, Role: arg.Role, FullName: arg.FullName}, nil
			},
		})
		_, err := svc.UpdateUser(ctx, UserUpdateInput{
			ID:         userID,
			Email:      "resident@example.com",
			Role:       "resident",
			FullName:   "Resident",
			PlotNumber: "15B",
			ActorID:    actorID,
		})
		require.NoError(t, err)
	})

	t.Run("update password invalid", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.UpdateUserPassword(ctx, UserPasswordInput{ID: userID}, "hash")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("update password not found", func(t *testing.T) {
		svc := New(&mockStore{
			updateUserPasswordFn: func(context.Context, repo.UpdateUserPasswordParams) (repo.User, error) {
				return repo.User{}, sql.ErrNoRows
			},
		})
		_, err := svc.UpdateUserPassword(ctx, UserPasswordInput{ID: userID, Password: "x", ActorID: actorID}, "hash")
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestServiceUnit_UserStateMethods(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	actorID := uuid.New()

	svc := New(&mockStore{
		softDeleteUserFn: func(_ context.Context, arg repo.SoftDeleteUserParams) error {
			require.Equal(t, userID, arg.ID)
			require.Equal(t, actorID, arg.UpdatedBy.UUID)
			return nil
		},
		restoreUserFn: func(_ context.Context, arg repo.RestoreUserParams) error {
			require.Equal(t, userID, arg.ID)
			return nil
		},
		blockUserFn: func(_ context.Context, arg repo.BlockUserParams) error {
			require.Equal(t, userID, arg.ID)
			return nil
		},
		unblockUserFn: func(_ context.Context, arg repo.UnblockUserParams) error {
			require.Equal(t, userID, arg.ID)
			return nil
		},
	})
	require.NoError(t, svc.SoftDeleteUser(ctx, userID, actorID))
	require.NoError(t, svc.RestoreUser(ctx, userID, actorID))
	require.NoError(t, svc.BlockUser(ctx, userID, actorID))
	require.NoError(t, svc.UnblockUser(ctx, userID, actorID))
}

func TestServiceUnit_PassMethods(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	passID := uuid.New()
	ownerID := uuid.New()

	t.Run("create pass invalid plate", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.CreatePass(ctx, PassCreateInput{PlateNumber: "bad"})
		require.ErrorIs(t, err, ErrInvalidPlate)
	})

	t.Run("create pass success", func(t *testing.T) {
		svc := New(&mockStore{
			createPassFn: func(_ context.Context, arg repo.CreatePassParams) (repo.Pass, error) {
				require.Equal(t, "A123BC77", arg.PlateNumber)
				require.Equal(t, actorID, arg.CreatedBy.UUID)
				return repo.Pass{ID: passID, PlateNumber: arg.PlateNumber, OwnerUserID: ownerID, Status: "active"}, nil
			},
		})
		_, err := svc.CreatePass(ctx, PassCreateInput{
			OwnerID:     ownerID,
			PlateNumber: "a123bc77",
			Status:      "active",
			ActorID:     actorID,
		})
		require.NoError(t, err)
	})

	t.Run("get pass and any", func(t *testing.T) {
		svc := New(&mockStore{
			getPassByIDFn: func(context.Context, uuid.UUID) (repo.Pass, error) { return repo.Pass{}, sql.ErrNoRows },
			getPassByIDAnyFn: func(context.Context, uuid.UUID) (repo.Pass, error) {
				return repo.Pass{ID: passID}, nil
			},
		})
		_, err := svc.GetPass(ctx, passID)
		require.ErrorIs(t, err, ErrNotFound)
		_, err = svc.GetPassAny(ctx, passID)
		require.NoError(t, err)
	})

	t.Run("list/search/update/delete/restore", func(t *testing.T) {
		svc := New(&mockStore{
			listPassesFn: func(context.Context, repo.ListPassesParams) ([]repo.Pass, error) {
				return []repo.Pass{{ID: passID}}, nil
			},
			listPassesByOwnerFn: func(_ context.Context, arg repo.ListPassesByOwnerParams) ([]repo.Pass, error) {
				require.Equal(t, ownerID, arg.OwnerUserID)
				return []repo.Pass{{ID: passID}}, nil
			},
			searchPassesByPlateFn: func(_ context.Context, arg repo.SearchPassesByPlateParams) ([]repo.Pass, error) {
				require.Equal(t, "%A123BC77%", arg.PlateNumber)
				return []repo.Pass{{ID: passID}}, nil
			},
			updatePassFn: func(_ context.Context, arg repo.UpdatePassParams) (repo.Pass, error) {
				require.Equal(t, actorID, arg.UpdatedBy.UUID)
				return repo.Pass{ID: arg.ID, PlateNumber: arg.PlateNumber}, nil
			},
			softDeletePassFn: func(context.Context, repo.SoftDeletePassParams) error { return nil },
			restorePassFn:    func(context.Context, repo.RestorePassParams) error { return nil },
		})
		_, err := svc.ListPasses(ctx, true, 10, 0)
		require.NoError(t, err)
		_, err = svc.ListPassesByOwner(ctx, ownerID, false, 10, 0)
		require.NoError(t, err)
		_, err = svc.SearchPasses(ctx, "A123BC77", 10, 0)
		require.NoError(t, err)
		_, err = svc.UpdatePass(ctx, PassUpdateInput{
			ID:          passID,
			PlateNumber: "A123BC77",
			Status:      "active",
			ActorID:     actorID,
		})
		require.NoError(t, err)
		require.NoError(t, svc.SoftDeletePass(ctx, passID, actorID))
		require.NoError(t, svc.RestorePass(ctx, passID, actorID))
	})
}

func TestServiceUnit_GuestMethods(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	guestID := uuid.New()
	residentID := uuid.New()
	now := time.Now().UTC()

	t.Run("create guest invalid plate", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.CreateGuestRequest(ctx, GuestCreateInput{
			ResidentID:  residentID,
			GuestName:   "Guest",
			PlateNumber: "bad",
			ValidFrom:   now,
			ValidTo:     now.Add(time.Hour),
		})
		require.ErrorIs(t, err, ErrInvalidPlate)
	})

	t.Run("create guest invalid range", func(t *testing.T) {
		svc := New(&mockStore{})
		_, err := svc.CreateGuestRequest(ctx, GuestCreateInput{
			ResidentID:  residentID,
			GuestName:   "Guest",
			PlateNumber: "A123BC77",
			ValidFrom:   now.Add(time.Hour),
			ValidTo:     now,
		})
		require.ErrorIs(t, err, ErrInvalidRange)
	})

	t.Run("guest full lifecycle methods", func(t *testing.T) {
		svc := New(&mockStore{
			createGuestRequestFn: func(_ context.Context, arg repo.CreateGuestRequestParams) (repo.GuestRequest, error) {
				require.Equal(t, "A123BC77", arg.PlateNumber)
				require.Equal(t, actorID, arg.CreatedBy.UUID)
				return repo.GuestRequest{ID: guestID, ResidentUserID: residentID, GuestFullName: arg.GuestFullName, PlateNumber: arg.PlateNumber}, nil
			},
			getGuestRequestByIDFn: func(_ context.Context, id uuid.UUID) (repo.GuestRequest, error) {
				if id == uuid.Nil {
					return repo.GuestRequest{}, sql.ErrNoRows
				}
				return repo.GuestRequest{ID: id, ResidentUserID: residentID, PlateNumber: "A123BC77", GuestFullName: "Guest", ValidFrom: now, ValidTo: now.Add(time.Hour), Status: "pending"}, nil
			},
			getGuestRequestByIDAnyFn: func(context.Context, uuid.UUID) (repo.GuestRequest, error) {
				return repo.GuestRequest{ID: guestID}, nil
			},
			listGuestRequestsFn: func(context.Context, repo.ListGuestRequestsParams) ([]repo.GuestRequest, error) {
				return []repo.GuestRequest{{ID: guestID}}, nil
			},
			listGuestRequestsByResidentFn: func(_ context.Context, arg repo.ListGuestRequestsByResidentParams) ([]repo.GuestRequest, error) {
				require.Equal(t, residentID, arg.ResidentUserID)
				return []repo.GuestRequest{{ID: guestID}}, nil
			},
			updateGuestRequestFn: func(_ context.Context, arg repo.UpdateGuestRequestParams) (repo.GuestRequest, error) {
				require.Equal(t, actorID, arg.UpdatedBy.UUID)
				return repo.GuestRequest{ID: arg.ID, GuestFullName: arg.GuestFullName, PlateNumber: arg.PlateNumber}, nil
			},
			softDeleteGuestRequestFn: func(context.Context, repo.SoftDeleteGuestRequestParams) error { return nil },
			restoreGuestRequestFn:    func(context.Context, repo.RestoreGuestRequestParams) error { return nil },
		})

		_, err := svc.CreateGuestRequest(ctx, GuestCreateInput{
			ResidentID:  residentID,
			GuestName:   "Guest",
			PlateNumber: "A123BC77",
			ValidFrom:   now,
			ValidTo:     now.Add(time.Hour),
			Status:      "pending",
			ActorID:     actorID,
		})
		require.NoError(t, err)

		_, err = svc.GetGuestRequest(ctx, uuid.Nil)
		require.ErrorIs(t, err, ErrNotFound)
		_, err = svc.GetGuestRequest(ctx, guestID)
		require.NoError(t, err)
		_, err = svc.GetGuestRequestAny(ctx, guestID)
		require.NoError(t, err)
		_, err = svc.ListGuestRequests(ctx, false, 10, 0)
		require.NoError(t, err)
		_, err = svc.ListGuestRequestsByResident(ctx, residentID, false, 10, 0)
		require.NoError(t, err)
		_, err = svc.UpdateGuestRequest(ctx, GuestUpdateInput{
			ID:          guestID,
			GuestName:   "Guest Updated",
			PlateNumber: "A123BC77",
			ValidFrom:   now,
			ValidTo:     now.Add(time.Hour),
			Status:      "approved",
			ActorID:     actorID,
		})
		require.NoError(t, err)
		require.NoError(t, svc.SoftDeleteGuestRequest(ctx, guestID, actorID))
		require.NoError(t, svc.RestoreGuestRequest(ctx, guestID, actorID))
	})
}

func TestServiceUnit_EntryLogMethods(t *testing.T) {
	ctx := context.Background()
	passID := uuid.New()
	guardID := uuid.New()
	entryID := uuid.New()

	svc := New(&mockStore{
		createEntryLogFn: func(_ context.Context, arg repo.CreateEntryLogParams) (repo.EntryLog, error) {
			require.Equal(t, passID, arg.PassID)
			require.Equal(t, "entry", arg.Action)
			return repo.EntryLog{ID: entryID, PassID: arg.PassID, GuardUserID: arg.GuardUserID, Action: arg.Action}, nil
		},
		listEntryLogsByPassFn: func(_ context.Context, arg repo.ListEntryLogsByPassParams) ([]repo.EntryLog, error) {
			require.Equal(t, passID, arg.PassID)
			return []repo.EntryLog{{ID: entryID, PassID: passID, GuardUserID: guardID, Action: "entry"}}, nil
		},
	})
	_, err := svc.CreateEntryLog(ctx, passID, guardID, "entry", sql.NullString{Valid: false})
	require.NoError(t, err)
	_, err = svc.ListEntryLogs(ctx, passID, 10, 0)
	require.NoError(t, err)
}

func TestServiceUnit_ErrorBranches(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	passID := uuid.New()
	guestID := uuid.New()
	repoErr := errors.New("repo failed")

	t.Run("authenticate repository error", func(t *testing.T) {
		svc := New(&mockStore{
			getUserByEmailFn: func(context.Context, string) (repo.User, error) {
				return repo.User{}, repoErr
			},
		})
		_, err := svc.Authenticate(ctx, "a@b.c", "pwd")
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("authenticate bad hash", func(t *testing.T) {
		svc := New(&mockStore{
			getUserByEmailFn: func(context.Context, string) (repo.User, error) {
				return repo.User{PasswordHash: "not-a-valid-hash"}, nil
			},
		})
		_, err := svc.Authenticate(ctx, "a@b.c", "pwd")
		require.Error(t, err)
	})

	t.Run("create user repository error", func(t *testing.T) {
		svc := New(&mockStore{
			createUserFn: func(context.Context, repo.CreateUserParams) (repo.User, error) {
				return repo.User{}, repoErr
			},
		})
		_, err := svc.CreateUser(ctx, UserCreateInput{
			Email:      "u@example.com",
			Password:   "pwd",
			Role:       "admin",
			FullName:   "User",
			PlotNumber: "",
		}, "hash")
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("get user generic error", func(t *testing.T) {
		svc := New(&mockStore{
			getUserByIDFn: func(context.Context, uuid.UUID) (repo.User, error) { return repo.User{}, repoErr },
		})
		_, err := svc.GetUser(ctx, userID)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("get user any not found", func(t *testing.T) {
		svc := New(&mockStore{
			getUserByIDAnyFn: func(context.Context, uuid.UUID) (repo.User, error) { return repo.User{}, sql.ErrNoRows },
		})
		_, err := svc.GetUserAny(ctx, userID)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("update user not found", func(t *testing.T) {
		svc := New(&mockStore{
			updateUserFn: func(context.Context, repo.UpdateUserParams) (repo.User, error) { return repo.User{}, sql.ErrNoRows },
		})
		_, err := svc.UpdateUser(ctx, UserUpdateInput{
			ID:       userID,
			Email:    "u@example.com",
			Role:     "admin",
			FullName: "User",
		})
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("update user password success", func(t *testing.T) {
		svc := New(&mockStore{
			updateUserPasswordFn: func(_ context.Context, arg repo.UpdateUserPasswordParams) (repo.User, error) {
				require.Equal(t, "hash", arg.PasswordHash)
				return repo.User{ID: arg.ID}, nil
			},
		})
		_, err := svc.UpdateUserPassword(ctx, UserPasswordInput{
			ID:       userID,
			Password: "newpass",
		}, "hash")
		require.NoError(t, err)
	})

	t.Run("create pass repository error", func(t *testing.T) {
		svc := New(&mockStore{
			createPassFn: func(context.Context, repo.CreatePassParams) (repo.Pass, error) { return repo.Pass{}, repoErr },
		})
		_, err := svc.CreatePass(ctx, PassCreateInput{
			OwnerID:     userID,
			PlateNumber: "A123BC77",
			Status:      "active",
		})
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("get pass generic error", func(t *testing.T) {
		svc := New(&mockStore{
			getPassByIDFn: func(context.Context, uuid.UUID) (repo.Pass, error) { return repo.Pass{}, repoErr },
		})
		_, err := svc.GetPass(ctx, passID)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("get pass any not found", func(t *testing.T) {
		svc := New(&mockStore{
			getPassByIDAnyFn: func(context.Context, uuid.UUID) (repo.Pass, error) { return repo.Pass{}, sql.ErrNoRows },
		})
		_, err := svc.GetPassAny(ctx, passID)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("search passes repository error", func(t *testing.T) {
		svc := New(&mockStore{
			searchPassesByPlateFn: func(context.Context, repo.SearchPassesByPlateParams) ([]repo.Pass, error) { return nil, repoErr },
		})
		_, err := svc.SearchPasses(ctx, "A123BC77", 10, 0)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("update pass not found", func(t *testing.T) {
		svc := New(&mockStore{
			updatePassFn: func(context.Context, repo.UpdatePassParams) (repo.Pass, error) { return repo.Pass{}, sql.ErrNoRows },
		})
		_, err := svc.UpdatePass(ctx, PassUpdateInput{
			ID:          passID,
			PlateNumber: "A123BC77",
			Status:      "active",
		})
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("create guest repository error", func(t *testing.T) {
		now := time.Now()
		svc := New(&mockStore{
			createGuestRequestFn: func(context.Context, repo.CreateGuestRequestParams) (repo.GuestRequest, error) {
				return repo.GuestRequest{}, repoErr
			},
		})
		_, err := svc.CreateGuestRequest(ctx, GuestCreateInput{
			ResidentID:  userID,
			GuestName:   "Guest",
			PlateNumber: "A123BC77",
			ValidFrom:   now,
			ValidTo:     now.Add(time.Hour),
			Status:      "pending",
		})
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("get guest any not found", func(t *testing.T) {
		svc := New(&mockStore{
			getGuestRequestByIDAnyFn: func(context.Context, uuid.UUID) (repo.GuestRequest, error) {
				return repo.GuestRequest{}, sql.ErrNoRows
			},
		})
		_, err := svc.GetGuestRequestAny(ctx, guestID)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("update guest not found", func(t *testing.T) {
		now := time.Now()
		svc := New(&mockStore{
			updateGuestRequestFn: func(context.Context, repo.UpdateGuestRequestParams) (repo.GuestRequest, error) {
				return repo.GuestRequest{}, sql.ErrNoRows
			},
		})
		_, err := svc.UpdateGuestRequest(ctx, GuestUpdateInput{
			ID:          guestID,
			GuestName:   "Guest",
			PlateNumber: "A123BC77",
			ValidFrom:   now,
			ValidTo:     now.Add(time.Hour),
			Status:      "pending",
		})
		require.ErrorIs(t, err, ErrNotFound)
	})
}
