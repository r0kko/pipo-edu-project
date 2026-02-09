package http

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pipo-edu-project/internal/identity"
	repo "pipo-edu-project/internal/repository/sqlc"
)

func TestParsePagination(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/passes", nil)
	limit, offset := parsePagination(req)
	require.Equal(t, int32(20), limit)
	require.Equal(t, int32(0), offset)

	req = httptest.NewRequest(http.MethodGet, "/passes?limit=120&offset=10", nil)
	limit, offset = parsePagination(req)
	require.Equal(t, int32(100), limit)
	require.Equal(t, int32(10), offset)

	req = httptest.NewRequest(http.MethodGet, "/passes?limit=bad&offset=-1", nil)
	limit, offset = parsePagination(req)
	require.Equal(t, int32(20), limit)
	require.Equal(t, int32(0), offset)
}

func TestCORSMiddleware(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware([]string{"http://allowed.local"})(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://allowed.local")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	require.True(t, called)
	require.Equal(t, "http://allowed.local", resp.Header().Get("Access-Control-Allow-Origin"))

	called = false
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://allowed.local")
	resp = httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	require.False(t, called)
	require.Equal(t, http.StatusNoContent, resp.Code)
}

func TestCORSMiddlewareAllowAll(t *testing.T) {
	handler := CORSMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://any.local")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	require.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
}

func TestContextHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.Equal(t, uuid.Nil, actorFromContext(req))
	require.Equal(t, "", roleFromContext(req))

	ctx := identity.WithUser(context.Background(), "not-uuid", "admin")
	req = req.WithContext(ctx)
	require.Equal(t, uuid.Nil, actorFromContext(req))
	require.Equal(t, "admin", roleFromContext(req))

	userID := uuid.New()
	ctx = identity.WithUser(context.Background(), userID.String(), "resident")
	req = req.WithContext(ctx)
	require.Equal(t, userID, actorFromContext(req))
}

func TestMapFunctions(t *testing.T) {
	now := time.Now().UTC()
	actor := uuid.New()

	user := repo.User{
		ID:         uuid.New(),
		Email:      "u@example.com",
		Role:       "resident",
		FullName:   "User",
		PlotNumber: sqlNullString("12A"),
		BlockedAt:  sqlNullTime(now),
		CreatedBy:  uuid.NullUUID{UUID: actor, Valid: true},
		UpdatedBy:  uuid.NullUUID{UUID: actor, Valid: true},
		DeletedAt:  sqlNullTime(now),
	}
	mappedUser := mapUser(user)
	require.NotNil(t, mappedUser.PlotNumber)
	require.NotNil(t, mappedUser.BlockedAt)
	require.NotNil(t, mappedUser.DeletedAt)

	pass := repo.Pass{
		ID:           uuid.New(),
		OwnerUserID:  user.ID,
		PlateNumber:  "A123BC77",
		VehicleBrand: sqlNullString("Toyota"),
		VehicleColor: sqlNullString("Black"),
		Status:       "active",
		CreatedBy:    uuid.NullUUID{UUID: actor, Valid: true},
		UpdatedBy:    uuid.NullUUID{UUID: actor, Valid: true},
		DeletedAt:    sqlNullTime(now),
	}
	mappedPass := mapPass(pass)
	require.NotNil(t, mappedPass.VehicleBrand)
	require.NotNil(t, mappedPass.VehicleColor)
	require.NotNil(t, mappedPass.DeletedAt)
	withOwner := mapPassWithOwner(pass, &user)
	require.NotNil(t, withOwner.OwnerFullName)
	require.NotNil(t, withOwner.OwnerPlotNumber)

	guest := repo.GuestRequest{
		ID:             uuid.New(),
		ResidentUserID: user.ID,
		GuestFullName:  "Guest",
		PlateNumber:    "A123BC77",
		ValidFrom:      now,
		ValidTo:        now.Add(time.Hour),
		Status:         "pending",
		CreatedBy:      uuid.NullUUID{UUID: actor, Valid: true},
		UpdatedBy:      uuid.NullUUID{UUID: actor, Valid: true},
		DeletedAt:      sqlNullTime(now),
	}
	mappedGuest := mapGuest(guest)
	require.NotNil(t, mappedGuest.DeletedAt)
	require.Len(t, mapPasses([]repo.Pass{pass}), 1)
	require.Len(t, mapGuests([]repo.GuestRequest{guest}), 1)
}

func sqlNullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: true}
}

func sqlNullTime(v time.Time) sql.NullTime {
	return sql.NullTime{Time: v, Valid: true}
}
