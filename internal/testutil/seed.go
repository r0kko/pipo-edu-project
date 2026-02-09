//go:build integration
// +build integration

package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pipo-edu-project/internal/auth"
	repo "pipo-edu-project/internal/repository/sqlc"
)

type SeedUsers struct {
	Admin    repo.User
	Guard    repo.User
	Resident repo.User
}

func SeedDefaultUsers(t *testing.T, q *repo.Queries) SeedUsers {
	t.Helper()
	ctx := context.Background()

	adminHash, err := auth.HashPassword("admin123")
	require.NoError(t, err)
	admin, err := q.CreateUser(ctx, repo.CreateUserParams{
		Email:        "admin@example.com",
		PasswordHash: adminHash,
		Role:         "admin",
		FullName:     "Admin",
		PlotNumber:   sql.NullString{Valid: false},
		CreatedBy:    uuid.NullUUID{Valid: false},
		UpdatedBy:    uuid.NullUUID{Valid: false},
	})
	require.NoError(t, err)

	guardHash, err := auth.HashPassword("guard123")
	require.NoError(t, err)
	guard, err := q.CreateUser(ctx, repo.CreateUserParams{
		Email:        "guard@example.com",
		PasswordHash: guardHash,
		Role:         "guard",
		FullName:     "Guard",
		PlotNumber:   sql.NullString{Valid: false},
		CreatedBy:    uuid.NullUUID{UUID: admin.ID, Valid: true},
		UpdatedBy:    uuid.NullUUID{UUID: admin.ID, Valid: true},
	})
	require.NoError(t, err)

	residentHash, err := auth.HashPassword("resident123")
	require.NoError(t, err)
	resident, err := q.CreateUser(ctx, repo.CreateUserParams{
		Email:        "resident@example.com",
		PasswordHash: residentHash,
		Role:         "resident",
		FullName:     "Resident",
		PlotNumber:   sql.NullString{String: "12A", Valid: true},
		CreatedBy:    uuid.NullUUID{UUID: admin.ID, Valid: true},
		UpdatedBy:    uuid.NullUUID{UUID: admin.ID, Valid: true},
	})
	require.NoError(t, err)

	return SeedUsers{Admin: admin, Guard: guard, Resident: resident}
}

func SeedPass(t *testing.T, q *repo.Queries, ownerID, actorID uuid.UUID, plate string) repo.Pass {
	t.Helper()
	ctx := context.Background()
	pass, err := q.CreatePass(ctx, repo.CreatePassParams{
		OwnerUserID:  ownerID,
		PlateNumber:  plate,
		VehicleBrand: sql.NullString{String: "Toyota", Valid: true},
		VehicleColor: sql.NullString{String: "Black", Valid: true},
		Status:       "active",
		CreatedBy:    uuid.NullUUID{UUID: actorID, Valid: true},
		UpdatedBy:    uuid.NullUUID{UUID: actorID, Valid: true},
	})
	require.NoError(t, err)
	return pass
}

func SeedGuestRequest(t *testing.T, q *repo.Queries, residentID, actorID uuid.UUID, plate string) repo.GuestRequest {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UTC()
	guest, err := q.CreateGuestRequest(ctx, repo.CreateGuestRequestParams{
		ResidentUserID: residentID,
		GuestFullName:  "Seed Guest",
		PlateNumber:    plate,
		ValidFrom:      now.Add(1 * time.Hour),
		ValidTo:        now.Add(2 * time.Hour),
		Status:         "pending",
		CreatedBy:      uuid.NullUUID{UUID: actorID, Valid: true},
		UpdatedBy:      uuid.NullUUID{UUID: actorID, Valid: true},
	})
	require.NoError(t, err)
	return guest
}
