//go:build integration
// +build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"pipo-edu-project/internal/repository"
	repo "pipo-edu-project/internal/repository/sqlc"
)

type TestDB struct {
	DSN       string
	DB        *sql.DB
	Queries   *repo.Queries
	container testcontainers.Container
}

func StartPostgres(t *testing.T) *TestDB {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "pipo",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		).WithStartupTimeout(90 * time.Second),
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
	require.NoError(t, repository.RunMigrations(dsn, migrationPath()))

	db, err := repository.Open(dsn)
	require.NoError(t, err)

	tdb := &TestDB{
		DSN:       dsn,
		DB:        db,
		Queries:   repo.New(db),
		container: container,
	}

	t.Cleanup(func() {
		_ = tdb.DB.Close()
		_ = tdb.container.Terminate(ctx)
	})

	return tdb
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

func migrationPath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "db", "migrations")
}
