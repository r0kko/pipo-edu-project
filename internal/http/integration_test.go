//go:build integration
// +build integration

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"pipo-edu-project/internal/auth"
	repo "pipo-edu-project/internal/repository/sqlc"
	"pipo-edu-project/internal/service"
	"pipo-edu-project/internal/testutil"
)

type testApp struct {
	server       *httptest.Server
	client       *http.Client
	tokens       *auth.TokenManager
	queries      *repo.Queries
	users        testutil.SeedUsers
	adminAccess  string
	guardAccess  string
	resAccess    string
	adminRefresh string
	resRefresh   string
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()
	setRepoRootWD(t)

	tdb := testutil.StartPostgres(t)
	users := testutil.SeedDefaultUsers(t, tdb.Queries)
	svc := service.New(tdb.Queries)
	tokens := auth.NewTokenManager("test-access", "test-refresh", time.Hour, 24*time.Hour)

	adminAccess, adminRefresh, err := tokens.GenerateTokens(users.Admin.ID, auth.RoleAdmin)
	require.NoError(t, err)
	guardAccess, _, err := tokens.GenerateTokens(users.Guard.ID, auth.RoleGuard)
	require.NoError(t, err)
	resAccess, resRefresh, err := tokens.GenerateTokens(users.Resident.ID, auth.RoleResident)
	require.NoError(t, err)

	router := NewRouter(&Handler{
		Auth:    tokens,
		Service: svc,
		Metrics: nil,
		CORS:    []string{"http://localhost:5173"},
	})
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &testApp{
		server:       server,
		client:       server.Client(),
		tokens:       tokens,
		queries:      tdb.Queries,
		users:        users,
		adminAccess:  adminAccess,
		guardAccess:  guardAccess,
		resAccess:    resAccess,
		adminRefresh: adminRefresh,
		resRefresh:   resRefresh,
	}
}

func (a *testApp) request(t *testing.T, method, path, token string, payload interface{}) (*http.Response, []byte) {
	t.Helper()
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, a.server.URL+path, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := a.client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	var out bytes.Buffer
	_, err = out.ReadFrom(resp.Body)
	require.NoError(t, err)
	return resp, out.Bytes()
}

func (a *testApp) requestRaw(t *testing.T, method, path, token string, raw string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, a.server.URL+path, bytes.NewBufferString(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := a.client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	var out bytes.Buffer
	_, err = out.ReadFrom(resp.Body)
	require.NoError(t, err)
	return resp, out.Bytes()
}

func setRepoRootWD(t *testing.T) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	old, err := filepath.Abs(".")
	require.NoError(t, err)
	err = os.Chdir(root)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestAPIIntegration_AllRoutes(t *testing.T) {
	app := setupTestApp(t)
	ctx := context.Background()

	var createdUserID uuid.UUID
	var createdPassID uuid.UUID
	var createdGuestID uuid.UUID
	var secondResidentToken string

	t.Run("public endpoints", func(t *testing.T) {
		resp, _ := app.request(t, http.MethodGet, "/health", "", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, body := app.request(t, http.MethodGet, "/docs", "", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, string(body), "Swagger UI")

		resp, body = app.request(t, http.MethodGet, "/openapi.yaml", "", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, string(body), "openapi:")
	})

	t.Run("auth login and refresh", func(t *testing.T) {
		resp, _ := app.request(t, http.MethodPost, "/auth/login", "", map[string]string{
			"email":    "admin@example.com",
			"password": "admin123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.requestRaw(t, http.MethodPost, "/auth/login", "", "{bad json")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/auth/login", "", map[string]string{
			"email":    "admin@example.com",
			"password": "wrong",
		})
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/auth/refresh", "", map[string]string{
			"refresh_token": app.adminRefresh,
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.requestRaw(t, http.MethodPost, "/auth/refresh", "", "{bad json")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/auth/refresh", "", map[string]string{
			"refresh_token": "invalid",
		})
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		require.NoError(t, app.queries.BlockUser(ctx, repo.BlockUserParams{
			ID:        app.users.Resident.ID,
			UpdatedBy: uuid.NullUUID{UUID: app.users.Admin.ID, Valid: true},
		}))
		resp, _ = app.request(t, http.MethodPost, "/auth/refresh", "", map[string]string{
			"refresh_token": app.resRefresh,
		})
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		require.NoError(t, app.queries.UnblockUser(ctx, repo.UnblockUserParams{
			ID:        app.users.Resident.ID,
			UpdatedBy: uuid.NullUUID{UUID: app.users.Admin.ID, Valid: true},
		}))
	})

	t.Run("users routes", func(t *testing.T) {
		resp, _ := app.request(t, http.MethodGet, "/users", app.resAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, body := app.request(t, http.MethodPost, "/users", app.adminAccess, map[string]string{
			"email":       "resident2@example.com",
			"password":    "resident123",
			"role":        "resident",
			"full_name":   "Resident 2",
			"plot_number": "15B",
		})
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var created map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &created))
		idRaw, ok := created["id"].(string)
		require.True(t, ok)
		parsed, err := uuid.Parse(idRaw)
		require.NoError(t, err)
		createdUserID = parsed

		resp, _ = app.request(t, http.MethodPost, "/users", app.adminAccess, map[string]string{
			"email":       "badresident@example.com",
			"password":    "resident123",
			"role":        "resident",
			"full_name":   "Bad Resident",
			"plot_number": "",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/users", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/users/"+createdUserID.String(), app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/users/not-a-uuid", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/users/"+uuid.New().String(), app.adminAccess, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/users/not-a-uuid", app.adminAccess, map[string]string{
			"full_name": "x",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.requestRaw(t, http.MethodPatch, "/users/"+createdUserID.String(), app.adminAccess, "{bad json")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/users/"+createdUserID.String(), app.adminAccess, map[string]string{
			"full_name": "Resident Updated",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/users/"+createdUserID.String(), app.adminAccess, map[string]string{
			"role": "unknown",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/users/"+createdUserID.String(), app.adminAccess, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/users/not-a-uuid", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/users/"+createdUserID.String()+"/restore", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/users/not-a-uuid/restore", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/users/"+createdUserID.String()+"/block", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/users/not-a-uuid/block", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/users/"+createdUserID.String()+"/unblock", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/users/not-a-uuid/unblock", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("passes routes", func(t *testing.T) {
		resp, _ := app.request(t, http.MethodPost, "/passes", app.guardAccess, map[string]string{
			"plate_number": "A123BC77",
		})
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, body := app.request(t, http.MethodPost, "/passes", app.resAccess, map[string]string{
			"plate_number": "A123BC77",
		})
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var pass map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &pass))
		idRaw, ok := pass["id"].(string)
		require.True(t, ok)
		parsed, err := uuid.Parse(idRaw)
		require.NoError(t, err)
		createdPassID = parsed

		resp, _ = app.request(t, http.MethodPost, "/passes", app.adminAccess, map[string]interface{}{
			"owner_user_id": app.users.Resident.ID.String(),
			"plate_number":  "M777MM77",
		})
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes", app.resAccess, map[string]string{
			"plate_number": "bad",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes", app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes", app.resAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes?includeDeleted=true", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes/"+createdPassID.String(), app.resAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes/not-a-uuid", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, body = app.request(t, http.MethodPost, "/users", app.adminAccess, map[string]string{
			"email":       "resident3@example.com",
			"password":    "resident123",
			"role":        "resident",
			"full_name":   "Resident 3",
			"plot_number": "33C",
		})
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		createdResidentID := parseUUIDField(t, body, "id")
		_, refresh, err := app.tokens.GenerateTokens(createdResidentID, auth.RoleResident)
		require.NoError(t, err)
		respRefresh, refreshBody := app.request(t, http.MethodPost, "/auth/refresh", "", map[string]string{"refresh_token": refresh})
		require.Equal(t, http.StatusOK, respRefresh.StatusCode)
		secondResidentToken = extractAccessToken(t, refreshBody)

		resp, _ = app.request(t, http.MethodGet, "/passes/"+createdPassID.String(), secondResidentToken, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/passes/"+createdPassID.String(), app.guardAccess, map[string]string{
			"plate_number": "A123BC77",
		})
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/passes/"+createdPassID.String(), app.resAccess, map[string]string{
			"vehicle_color": "White",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/passes/not-a-uuid", app.adminAccess, map[string]string{
			"plate_number": "A123BC77",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.requestRaw(t, http.MethodPatch, "/passes/"+createdPassID.String(), app.adminAccess, "{bad json")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/passes/"+createdPassID.String(), app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/passes/"+createdPassID.String(), app.resAccess, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/passes/not-a-uuid", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+createdPassID.String()+"/restore", app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+createdPassID.String()+"/restore", app.resAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/not-a-uuid/restore", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes/search?plate=A123BC77", app.resAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes/search", app.guardAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/passes/search?plate=A123BC77", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+createdPassID.String()+"/entry", app.resAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+createdPassID.String()+"/entry", app.guardAccess, nil)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/not-a-uuid/entry", app.guardAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.requestRaw(t, http.MethodPost, "/passes/"+createdPassID.String()+"/entry", app.guardAccess, "{bad json")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+uuid.New().String()+"/entry", app.guardAccess, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+createdPassID.String()+"/exit", app.resAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/"+createdPassID.String()+"/exit", app.adminAccess, nil)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/passes/not-a-uuid/exit", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("guest requests routes", func(t *testing.T) {
		now := time.Now().UTC()
		resp, _ := app.request(t, http.MethodPost, "/guest-requests", app.guardAccess, map[string]interface{}{
			"guest_full_name": "Guest",
			"plate_number":    "A123BC77",
			"valid_from":      now.Add(1 * time.Hour),
			"valid_to":        now.Add(2 * time.Hour),
		})
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, body := app.request(t, http.MethodPost, "/guest-requests", app.resAccess, map[string]interface{}{
			"guest_full_name": "Guest",
			"plate_number":    "A123BC77",
			"valid_from":      now.Add(1 * time.Hour),
			"valid_to":        now.Add(2 * time.Hour),
		})
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var guest map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &guest))
		parsed, err := uuid.Parse(guest["id"].(string))
		require.NoError(t, err)
		createdGuestID = parsed

		resp, _ = app.request(t, http.MethodPost, "/guest-requests", app.resAccess, map[string]interface{}{
			"guest_full_name": "Guest",
			"plate_number":    "A123BC77",
			"valid_from":      now.Add(3 * time.Hour),
			"valid_to":        now.Add(2 * time.Hour),
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/guest-requests", app.adminAccess, map[string]interface{}{
			"resident_user_id": app.users.Resident.ID.String(),
			"guest_full_name":  "Guest 2",
			"plate_number":     "M777MM77",
			"valid_from":       now.Add(1 * time.Hour),
			"valid_to":         now.Add(2 * time.Hour),
			"status":           "approved",
		})
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests", app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests", app.resAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests?includeDeleted=true", app.adminAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests/"+createdGuestID.String(), app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests/"+createdGuestID.String(), app.resAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests/not-a-uuid", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodGet, "/guest-requests/"+createdGuestID.String(), secondResidentToken, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/guest-requests/"+createdGuestID.String(), app.guardAccess, map[string]interface{}{
			"status": "approved",
		})
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/guest-requests/"+createdGuestID.String(), app.resAccess, map[string]interface{}{
			"guest_full_name": "Guest Updated",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPatch, "/guest-requests/not-a-uuid", app.adminAccess, map[string]interface{}{
			"guest_full_name": "Guest Updated",
		})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.requestRaw(t, http.MethodPatch, "/guest-requests/"+createdGuestID.String(), app.adminAccess, "{bad json")
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/guest-requests/"+createdGuestID.String(), app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/guest-requests/"+createdGuestID.String(), app.resAccess, nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = app.request(t, http.MethodDelete, "/guest-requests/not-a-uuid", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/guest-requests/"+createdGuestID.String()+"/restore", app.guardAccess, nil)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/guest-requests/"+createdGuestID.String()+"/restore", app.resAccess, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp, _ = app.request(t, http.MethodPost, "/guest-requests/not-a-uuid/restore", app.adminAccess, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func parseUUIDField(t *testing.T, body []byte, key string) uuid.UUID {
	t.Helper()
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &out))
	id, ok := out[key].(string)
	require.True(t, ok)
	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	return parsed
}

func extractAccessToken(t *testing.T, body []byte) string {
	t.Helper()
	var out map[string]string
	require.NoError(t, json.Unmarshal(body, &out))
	token, ok := out["access_token"]
	require.True(t, ok)
	return token
}
