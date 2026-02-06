package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"pipo-edu-project/internal/auth"
	"pipo-edu-project/internal/repository"
	repo "pipo-edu-project/internal/repository/sqlc"
)

func setupService(t *testing.T) (*Service, func()) {
	t.Helper()
	ctx := context.Background()

	if os.Getenv("TESTCONTAINERS_DISABLED") == "true" {
		t.Skip("testcontainers disabled")
	}
	if _, err := os.Stat("/var/run/docker.sock"); err != nil && os.Getenv("DOCKER_HOST") == "" {
		t.Skip("docker socket not available")
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "pipo",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/pipo?sslmode=disable", host, port.Port())
	require.NoError(t, waitForDB(dsn, 60*time.Second))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	migrationsPath := filepath.Join(cwd, "..", "..", "db", "migrations")
	require.NoError(t, runMigrationsWithRetry(dsn, migrationsPath, 60*time.Second))

	db, err := repository.Open(dsn)
	require.NoError(t, err)

	queries := repo.New(db)
	svc := New(queries)

	cleanup := func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}

	return svc, cleanup
}

func waitForDB(dsn string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := repository.Open(dsn)
		if err == nil {
			_ = db.Close()
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("database did not become ready in %s", timeout)
	}
	return fmt.Errorf("wait for db: %w", lastErr)
}

func runMigrationsWithRetry(dsn, migrationsPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		err := repository.RunMigrations(dsn, migrationsPath)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("migrations did not complete in %s", timeout)
	}
	return fmt.Errorf("run migrations with retry: %w", lastErr)
}

func TestServiceCRUDIntegration(t *testing.T) {
	svc, cleanup := setupService(t)
	defer cleanup()
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
