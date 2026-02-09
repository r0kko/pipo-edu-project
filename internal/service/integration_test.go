//go:build integration
// +build integration

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pipo-edu-project/internal/auth"
	"pipo-edu-project/internal/testutil"
)

func TestServiceCRUDIntegration(t *testing.T) {
	tdb := testutil.StartPostgres(t)
	svc := New(tdb.Queries)
	ctx := context.Background()

	adminHash, err := auth.HashPassword("admin123")
	require.NoError(t, err)
	admin, err := svc.CreateUser(ctx, UserCreateInput{
		Email:    "admin@example.com",
		Password: "admin123",
		Role:     "admin",
		FullName: "Admin",
		ActorID:  uuid.Nil,
	}, adminHash)
	require.NoError(t, err)

	residentHash, err := auth.HashPassword("resident123")
	require.NoError(t, err)
	resident, err := svc.CreateUser(ctx, UserCreateInput{
		Email:      "resident@example.com",
		Password:   "resident123",
		Role:       "resident",
		FullName:   "Resident",
		PlotNumber: "12A",
		ActorID:    admin.ID,
	}, residentHash)
	require.NoError(t, err)

	require.NoError(t, svc.BlockUser(ctx, resident.ID, admin.ID))
	_, err = svc.Authenticate(ctx, resident.Email, "resident123")
	require.ErrorIs(t, err, ErrBlocked)
	require.NoError(t, svc.UnblockUser(ctx, resident.ID, admin.ID))
	_, err = svc.Authenticate(ctx, resident.Email, "resident123")
	require.NoError(t, err)

	users, err := svc.ListUsers(ctx, false, 10, 0)
	require.NoError(t, err)
	require.Len(t, users, 2)

	updatedUser, err := svc.UpdateUser(ctx, UserUpdateInput{
		ID:         resident.ID,
		Email:      "resident2@example.com",
		Role:       "resident",
		FullName:   "Resident Updated",
		PlotNumber: "14B",
		ActorID:    admin.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "resident2@example.com", updatedUser.Email)
	require.True(t, updatedUser.PlotNumber.Valid)
	require.Equal(t, "14B", updatedUser.PlotNumber.String)

	require.NoError(t, svc.SoftDeleteUser(ctx, resident.ID, admin.ID))
	_, err = svc.GetUser(ctx, resident.ID)
	require.Error(t, err)
	require.NoError(t, svc.RestoreUser(ctx, resident.ID, admin.ID))
	_, err = svc.GetUser(ctx, resident.ID)
	require.NoError(t, err)

	pass, err := svc.CreatePass(ctx, PassCreateInput{
		OwnerID:      resident.ID,
		PlateNumber:  "A123BC77",
		VehicleBrand: sql.NullString{String: "Toyota", Valid: true},
		VehicleColor: sql.NullString{String: "Black", Valid: true},
		Status:       "active",
		ActorID:      admin.ID,
	})
	require.NoError(t, err)

	pass, err = svc.UpdatePass(ctx, PassUpdateInput{
		ID:           pass.ID,
		PlateNumber:  "A123BC77",
		VehicleBrand: sql.NullString{String: "Toyota", Valid: true},
		VehicleColor: sql.NullString{String: "White", Valid: true},
		Status:       "active",
		ActorID:      admin.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "White", pass.VehicleColor.String)

	found, err := svc.SearchPasses(ctx, "A123BC77", 10, 0)
	require.NoError(t, err)
	require.Len(t, found, 1)

	require.NoError(t, svc.SoftDeletePass(ctx, pass.ID, admin.ID))
	_, err = svc.GetPass(ctx, pass.ID)
	require.Error(t, err)
	require.NoError(t, svc.RestorePass(ctx, pass.ID, admin.ID))
	_, err = svc.GetPass(ctx, pass.ID)
	require.NoError(t, err)

	guest, err := svc.CreateGuestRequest(ctx, GuestCreateInput{
		ResidentID:  resident.ID,
		GuestName:   "Guest User",
		PlateNumber: "A123BC77",
		ValidFrom:   time.Now().Add(1 * time.Hour),
		ValidTo:     time.Now().Add(2 * time.Hour),
		Status:      "pending",
		ActorID:     resident.ID,
	})
	require.NoError(t, err)

	guest, err = svc.UpdateGuestRequest(ctx, GuestUpdateInput{
		ID:          guest.ID,
		GuestName:   "Guest User 2",
		PlateNumber: "A123BC77",
		ValidFrom:   guest.ValidFrom,
		ValidTo:     guest.ValidTo.Add(1 * time.Hour),
		Status:      "approved",
		ActorID:     admin.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "approved", guest.Status)

	require.NoError(t, svc.SoftDeleteGuestRequest(ctx, guest.ID, admin.ID))
	_, err = svc.GetGuestRequest(ctx, guest.ID)
	require.Error(t, err)
	require.NoError(t, svc.RestoreGuestRequest(ctx, guest.ID, admin.ID))
	_, err = svc.GetGuestRequest(ctx, guest.ID)
	require.NoError(t, err)

	guardHash, err := auth.HashPassword("guard123")
	require.NoError(t, err)
	guard, err := svc.CreateUser(ctx, UserCreateInput{
		Email:    "guard@example.com",
		Password: "guard123",
		Role:     "guard",
		FullName: "Guard",
		ActorID:  admin.ID,
	}, guardHash)
	require.NoError(t, err)

	logEntry, err := svc.CreateEntryLog(ctx, pass.ID, guard.ID, "entry", sql.NullString{Valid: false})
	require.NoError(t, err)
	require.Equal(t, "entry", logEntry.Action)

	logs, err := svc.ListEntryLogs(ctx, pass.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, logs, 1)
}
