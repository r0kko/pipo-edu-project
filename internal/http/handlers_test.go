package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"pipo-edu-project/internal/auth"
	repo "pipo-edu-project/internal/repository/sqlc"
	"pipo-edu-project/internal/service"
)

type stubService struct{}

func (s stubService) Authenticate(ctx context.Context, email, password string) (repo.User, error) {
	return repo.User{}, nil
}

func (s stubService) CreateUser(ctx context.Context, input service.UserCreateInput, passwordHash string) (repo.User, error) {
	return repo.User{ID: uuid.New(), Email: input.Email, Role: input.Role, FullName: input.FullName, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) GetUser(ctx context.Context, id uuid.UUID) (repo.User, error) {
	return repo.User{ID: id, Email: "user@example.com", Role: string(auth.RoleAdmin), FullName: "Admin", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) GetUserAny(ctx context.Context, id uuid.UUID) (repo.User, error) {
	return s.GetUser(ctx, id)
}

func (s stubService) ListUsers(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.User, error) {
	return []repo.User{}, nil
}

func (s stubService) UpdateUser(ctx context.Context, input service.UserUpdateInput) (repo.User, error) {
	return repo.User{ID: input.ID, Email: input.Email, Role: input.Role, FullName: input.FullName, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) UpdateUserPassword(ctx context.Context, input service.UserPasswordInput, passwordHash string) (repo.User, error) {
	return repo.User{ID: input.ID}, nil
}

func (s stubService) SoftDeleteUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) RestoreUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) BlockUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) UnblockUser(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) CreatePass(ctx context.Context, input service.PassCreateInput) (repo.Pass, error) {
	return repo.Pass{ID: uuid.New(), OwnerUserID: input.OwnerID, PlateNumber: input.PlateNumber, Status: input.Status, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) GetPass(ctx context.Context, id uuid.UUID) (repo.Pass, error) {
	return repo.Pass{ID: id, OwnerUserID: uuid.New(), PlateNumber: "A123BC77", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) GetPassAny(ctx context.Context, id uuid.UUID) (repo.Pass, error) {
	return s.GetPass(ctx, id)
}

func (s stubService) ListPasses(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.Pass, error) {
	return []repo.Pass{}, nil
}

func (s stubService) ListPassesByOwner(ctx context.Context, owner uuid.UUID, includeDeleted bool, limit, offset int32) ([]repo.Pass, error) {
	return []repo.Pass{}, nil
}

func (s stubService) SearchPasses(ctx context.Context, plate string, limit, offset int32) ([]repo.Pass, error) {
	return []repo.Pass{}, nil
}

func (s stubService) UpdatePass(ctx context.Context, input service.PassUpdateInput) (repo.Pass, error) {
	return repo.Pass{ID: input.ID, PlateNumber: input.PlateNumber, Status: input.Status, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) SoftDeletePass(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) RestorePass(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) CreateGuestRequest(ctx context.Context, input service.GuestCreateInput) (repo.GuestRequest, error) {
	return repo.GuestRequest{ID: uuid.New(), ResidentUserID: input.ResidentID, GuestFullName: input.GuestName, PlateNumber: input.PlateNumber, Status: input.Status, ValidFrom: input.ValidFrom, ValidTo: input.ValidTo, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) GetGuestRequest(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error) {
	return repo.GuestRequest{ID: id, ResidentUserID: uuid.New(), GuestFullName: "Guest", PlateNumber: "A123BC77", Status: "pending", ValidFrom: time.Now(), ValidTo: time.Now().Add(2 * time.Hour), CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) GetGuestRequestAny(ctx context.Context, id uuid.UUID) (repo.GuestRequest, error) {
	return s.GetGuestRequest(ctx, id)
}

func (s stubService) ListGuestRequests(ctx context.Context, includeDeleted bool, limit, offset int32) ([]repo.GuestRequest, error) {
	return []repo.GuestRequest{}, nil
}

func (s stubService) ListGuestRequestsByResident(ctx context.Context, resident uuid.UUID, includeDeleted bool, limit, offset int32) ([]repo.GuestRequest, error) {
	return []repo.GuestRequest{}, nil
}

func (s stubService) UpdateGuestRequest(ctx context.Context, input service.GuestUpdateInput) (repo.GuestRequest, error) {
	return repo.GuestRequest{ID: input.ID, GuestFullName: input.GuestName, PlateNumber: input.PlateNumber, Status: input.Status, ValidFrom: input.ValidFrom, ValidTo: input.ValidTo, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubService) SoftDeleteGuestRequest(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) RestoreGuestRequest(ctx context.Context, id uuid.UUID, actor uuid.UUID) error {
	return nil
}

func (s stubService) CreateEntryLog(ctx context.Context, passID, guardID uuid.UUID, action string, comment sql.NullString) (repo.EntryLog, error) {
	return repo.EntryLog{ID: uuid.New(), PassID: passID, GuardUserID: guardID, Action: action, ActionAt: time.Now()}, nil
}

func (s stubService) ListEntryLogs(ctx context.Context, passID uuid.UUID, limit, offset int32) ([]repo.EntryLog, error) {
	return []repo.EntryLog{}, nil
}

func newAuthToken(role auth.Role) string {
	manager := auth.NewTokenManager("test-access", "test-refresh", time.Hour, time.Hour)
	access, _, _ := manager.GenerateTokens(uuid.New(), role)
	return access
}

func setupRouter() http.Handler {
	manager := auth.NewTokenManager("test-access", "test-refresh", time.Hour, time.Hour)
	return NewRouter(&Handler{Auth: manager, Service: stubService{}})
}

func TestGuardCannotCreatePass(t *testing.T) {
	router := setupRouter()
	payload := map[string]string{"plate_number": "A123BC77"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/passes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+newAuthToken(auth.RoleGuard))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestAdminCanCreateUser(t *testing.T) {
	router := setupRouter()
	payload := map[string]string{"email": "user@example.com", "password": "secret", "role": "resident", "full_name": "User", "plot_number": "12A"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+newAuthToken(auth.RoleAdmin))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestResidentCannotAccessUsers(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+newAuthToken(auth.RoleResident))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.Code)
	}
}

func TestAdminCanBlockUser(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest(http.MethodPost, "/users/"+uuid.New().String()+"/block", nil)
	req.Header.Set("Authorization", "Bearer "+newAuthToken(auth.RoleAdmin))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestResidentCanCreatePass(t *testing.T) {
	router := setupRouter()
	payload := map[string]string{"plate_number": "A123BC77"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/passes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+newAuthToken(auth.RoleResident))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}
}

func TestGuardCanSearchPasses(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/passes/search?plate=A123BC77", nil)
	req.Header.Set("Authorization", "Bearer "+newAuthToken(auth.RoleGuard))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}
