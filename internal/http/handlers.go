package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"pipo-edu-project/internal/auth"
	"pipo-edu-project/internal/identity"
	repo "pipo-edu-project/internal/repository/sqlc"
	"pipo-edu-project/internal/service"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	FullName   string     `json:"full_name"`
	PlotNumber *string    `json:"plot_number,omitempty"`
	BlockedAt  *time.Time `json:"blocked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy  *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateUserRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	FullName   string `json:"full_name"`
	PlotNumber string `json:"plot_number"`
}

type UpdateUserRequest struct {
	Email      *string `json:"email"`
	Role       *string `json:"role"`
	FullName   *string `json:"full_name"`
	Password   *string `json:"password"`
	PlotNumber *string `json:"plot_number"`
}

type PassRequest struct {
	OwnerUserID  *uuid.UUID `json:"owner_user_id,omitempty"`
	PlateNumber  string     `json:"plate_number"`
	VehicleBrand *string    `json:"vehicle_brand,omitempty"`
	VehicleColor *string    `json:"vehicle_color,omitempty"`
	Status       *string    `json:"status,omitempty"`
}

type PassResponse struct {
	ID              uuid.UUID  `json:"id"`
	OwnerUserID     uuid.UUID  `json:"owner_user_id"`
	OwnerFullName   *string    `json:"owner_full_name,omitempty"`
	OwnerPlotNumber *string    `json:"owner_plot_number,omitempty"`
	PlateNumber     string     `json:"plate_number"`
	VehicleBrand    *string    `json:"vehicle_brand,omitempty"`
	VehicleColor    *string    `json:"vehicle_color,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type GuestRequest struct {
	ResidentUserID *uuid.UUID `json:"resident_user_id,omitempty"`
	GuestFullName  string     `json:"guest_full_name"`
	PlateNumber    string     `json:"plate_number"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidTo        time.Time  `json:"valid_to"`
	Status         *string    `json:"status,omitempty"`
}

type GuestResponse struct {
	ID             uuid.UUID  `json:"id"`
	ResidentUserID uuid.UUID  `json:"resident_user_id"`
	GuestFullName  string     `json:"guest_full_name"`
	PlateNumber    string     `json:"plate_number"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidTo        time.Time  `json:"valid_to"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy      *uuid.UUID `json:"updated_by,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type EntryRequest struct {
	Comment *string `json:"comment"`
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	user, err := h.Service.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	access, refresh, err := h.Auth.GenerateTokens(user.ID, auth.Role(user.Role))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "token error")
		return
	}
	WriteJSON(w, http.StatusOK, TokenResponse{AccessToken: access, RefreshToken: refresh, User: mapUser(user)})
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	claims, err := h.Auth.ParseRefresh(req.RefreshToken)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	user, err := h.Service.GetUserAny(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	if user.DeletedAt.Valid || user.BlockedAt.Valid {
		WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	access, refresh, err := h.Auth.GenerateTokens(userID, claims.Role)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "token error")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"access_token": access, "refresh_token": refresh})
}

func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	actorID := actorFromContext(r)
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "hash error")
		return
	}
	user, err := h.Service.CreateUser(r.Context(), service.UserCreateInput{
		Email:      req.Email,
		Password:   req.Password,
		Role:       req.Role,
		FullName:   req.FullName,
		PlotNumber: req.PlotNumber,
		ActorID:    actorID,
	}, passwordHash)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Users.WithLabelValues("created").Inc()
	}
	WriteJSON(w, http.StatusCreated, mapUser(user))
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user, err := h.Service.GetUser(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	WriteJSON(w, http.StatusOK, mapUser(user))
}

func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"
	limit, offset := parsePagination(r)
	users, err := h.Service.ListUsers(r.Context(), includeDeleted, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "error")
		return
	}
	resp := make([]UserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, mapUser(user))
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	actorID := actorFromContext(r)
	user, err := h.Service.GetUser(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}

	if req.Password != nil {
		passwordHash, err := auth.HashPassword(*req.Password)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "hash error")
			return
		}
		_, err = h.Service.UpdateUserPassword(r.Context(), service.UserPasswordInput{
			ID:       id,
			Password: *req.Password,
			ActorID:  actorID,
		}, passwordHash)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	email := user.Email
	role := user.Role
	fullName := user.FullName
	plotNumber := ""
	if user.PlotNumber.Valid {
		plotNumber = user.PlotNumber.String
	}
	if req.Email != nil {
		email = *req.Email
	}
	if req.Role != nil {
		role = *req.Role
	}
	if req.FullName != nil {
		fullName = *req.FullName
	}
	if req.PlotNumber != nil {
		plotNumber = *req.PlotNumber
	}
	updated, err := h.Service.UpdateUser(r.Context(), service.UserUpdateInput{
		ID:         id,
		Email:      email,
		Role:       role,
		FullName:   fullName,
		PlotNumber: plotNumber,
		ActorID:    actorID,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Users.WithLabelValues("updated").Inc()
	}
	WriteJSON(w, http.StatusOK, mapUser(updated))
}

func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := actorFromContext(r)
	if err := h.Service.SoftDeleteUser(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Users.WithLabelValues("deleted").Inc()
	}
	WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) HandleRestoreUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := actorFromContext(r)
	if err := h.Service.RestoreUser(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Users.WithLabelValues("restored").Inc()
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "restored"})
}

func (h *Handler) HandleBlockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := actorFromContext(r)
	if err := h.Service.BlockUser(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Users.WithLabelValues("blocked").Inc()
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "blocked"})
}

func (h *Handler) HandleUnblockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	actorID := actorFromContext(r)
	if err := h.Service.UnblockUser(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Users.WithLabelValues("unblocked").Inc()
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "unblocked"})
}

func (h *Handler) HandleCreatePass(w http.ResponseWriter, r *http.Request) {
	role := roleFromContext(r)
	var req PassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	actorID := actorFromContext(r)
	ownerID := actorID
	if role == string(auth.RoleAdmin) {
		if req.OwnerUserID != nil {
			ownerID = *req.OwnerUserID
		}
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	status := "active"
	if req.Status != nil && role == string(auth.RoleAdmin) {
		status = *req.Status
	}
	pass, err := h.Service.CreatePass(r.Context(), service.PassCreateInput{
		OwnerID:      ownerID,
		PlateNumber:  req.PlateNumber,
		VehicleBrand: toNullString(req.VehicleBrand),
		VehicleColor: toNullString(req.VehicleColor),
		Status:       status,
		ActorID:      actorID,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Passes.WithLabelValues("created").Inc()
	}
	WriteJSON(w, http.StatusCreated, mapPass(pass))
}

func (h *Handler) HandleGetPass(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	pass, err := h.Service.GetPass(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && pass.OwnerUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	owner, err := h.Service.GetUserAny(r.Context(), pass.OwnerUserID)
	if err != nil {
		WriteJSON(w, http.StatusOK, mapPass(pass))
		return
	}
	WriteJSON(w, http.StatusOK, mapPassWithOwner(pass, &owner))
}

func (h *Handler) HandleListPasses(w http.ResponseWriter, r *http.Request) {
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	limit, offset := parsePagination(r)
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	if role == string(auth.RoleAdmin) {
		passes, err := h.Service.ListPasses(r.Context(), includeDeleted, limit, offset)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "error")
			return
		}
		WriteJSON(w, http.StatusOK, mapPasses(passes))
		return
	}

	if role == string(auth.RoleResident) {
		passes, err := h.Service.ListPassesByOwner(r.Context(), actorID, false, limit, offset)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "error")
			return
		}
		WriteJSON(w, http.StatusOK, mapPasses(passes))
		return
	}

	WriteError(w, http.StatusForbidden, "forbidden")
}

func (h *Handler) HandleSearchPasses(w http.ResponseWriter, r *http.Request) {
	role := roleFromContext(r)
	if role != string(auth.RoleGuard) && role != string(auth.RoleAdmin) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	plate := r.URL.Query().Get("plate")
	if plate == "" {
		WriteError(w, http.StatusBadRequest, "missing plate")
		return
	}
	limit, offset := parsePagination(r)
	passes, err := h.Service.SearchPasses(r.Context(), plate, limit, offset)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := make([]PassResponse, 0, len(passes))
	for _, pass := range passes {
		owner, err := h.Service.GetUserAny(r.Context(), pass.OwnerUserID)
		if err != nil {
			resp = append(resp, mapPass(pass))
			continue
		}
		resp = append(resp, mapPassWithOwner(pass, &owner))
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleUpdatePass(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	var req PassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	pass, err := h.Service.GetPass(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && pass.OwnerUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	plate := pass.PlateNumber
	if req.PlateNumber != "" {
		plate = req.PlateNumber
	}
	brand := pass.VehicleBrand
	color := pass.VehicleColor
	if req.VehicleBrand != nil {
		brand = toNullString(req.VehicleBrand)
	}
	if req.VehicleColor != nil {
		color = toNullString(req.VehicleColor)
	}
	status := pass.Status
	if req.Status != nil && role == string(auth.RoleAdmin) {
		status = *req.Status
	}
	updated, err := h.Service.UpdatePass(r.Context(), service.PassUpdateInput{
		ID:           id,
		PlateNumber:  plate,
		VehicleBrand: brand,
		VehicleColor: color,
		Status:       status,
		ActorID:      actorID,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Passes.WithLabelValues("updated").Inc()
	}
	WriteJSON(w, http.StatusOK, mapPass(updated))
}

func (h *Handler) HandleDeletePass(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	pass, err := h.Service.GetPass(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && pass.OwnerUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.Service.SoftDeletePass(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Passes.WithLabelValues("deleted").Inc()
	}
	WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) HandleRestorePass(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	pass, err := h.Service.GetPassAny(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && pass.OwnerUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.Service.RestorePass(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Passes.WithLabelValues("restored").Inc()
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "restored"})
}

func (h *Handler) HandleEntry(w http.ResponseWriter, r *http.Request) {
	h.handleEntryExit(w, r, "entry")
}

func (h *Handler) HandleExit(w http.ResponseWriter, r *http.Request) {
	h.handleEntryExit(w, r, "exit")
}

func (h *Handler) handleEntryExit(w http.ResponseWriter, r *http.Request, action string) {
	role := roleFromContext(r)
	if role != string(auth.RoleGuard) && role != string(auth.RoleAdmin) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	passID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req EntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	_, err = h.Service.GetPass(r.Context(), passID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "pass not found")
		return
	}
	comment := toNullString(req.Comment)
	logEntry, err := h.Service.CreateEntryLog(r.Context(), passID, actorFromContext(r), action, comment)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "log error")
		return
	}
	WriteJSON(w, http.StatusCreated, logEntry)
}

func (h *Handler) HandleCreateGuest(w http.ResponseWriter, r *http.Request) {
	role := roleFromContext(r)
	var req GuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	actorID := actorFromContext(r)
	residentID := actorID
	if role == string(auth.RoleAdmin) && req.ResidentUserID != nil {
		residentID = *req.ResidentUserID
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	status := "pending"
	if req.Status != nil && role == string(auth.RoleAdmin) {
		status = *req.Status
	}
	guest, err := h.Service.CreateGuestRequest(r.Context(), service.GuestCreateInput{
		ResidentID:  residentID,
		GuestName:   req.GuestFullName,
		PlateNumber: req.PlateNumber,
		ValidFrom:   req.ValidFrom,
		ValidTo:     req.ValidTo,
		Status:      status,
		ActorID:     actorID,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Guest.WithLabelValues("created").Inc()
	}
	WriteJSON(w, http.StatusCreated, mapGuest(guest))
}

func (h *Handler) HandleListGuest(w http.ResponseWriter, r *http.Request) {
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	limit, offset := parsePagination(r)
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"

	if role == string(auth.RoleAdmin) {
		guests, err := h.Service.ListGuestRequests(r.Context(), includeDeleted, limit, offset)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "error")
			return
		}
		WriteJSON(w, http.StatusOK, mapGuests(guests))
		return
	}

	if role == string(auth.RoleResident) {
		guests, err := h.Service.ListGuestRequestsByResident(r.Context(), actorID, false, limit, offset)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "error")
			return
		}
		WriteJSON(w, http.StatusOK, mapGuests(guests))
		return
	}

	WriteError(w, http.StatusForbidden, "forbidden")
}

func (h *Handler) HandleGetGuest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	guest, err := h.Service.GetGuestRequest(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && guest.ResidentUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	WriteJSON(w, http.StatusOK, mapGuest(guest))
}

func (h *Handler) HandleUpdateGuest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	var req GuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	guest, err := h.Service.GetGuestRequest(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && guest.ResidentUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	guestName := guest.GuestFullName
	plate := guest.PlateNumber
	validFrom := guest.ValidFrom
	validTo := guest.ValidTo
	status := guest.Status
	if req.GuestFullName != "" {
		guestName = req.GuestFullName
	}
	if req.PlateNumber != "" {
		plate = req.PlateNumber
	}
	if !req.ValidFrom.IsZero() {
		validFrom = req.ValidFrom
	}
	if !req.ValidTo.IsZero() {
		validTo = req.ValidTo
	}
	if req.Status != nil && role == string(auth.RoleAdmin) {
		status = *req.Status
	}
	updated, err := h.Service.UpdateGuestRequest(r.Context(), service.GuestUpdateInput{
		ID:          id,
		GuestName:   guestName,
		PlateNumber: plate,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		Status:      status,
		ActorID:     actorID,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Guest.WithLabelValues("updated").Inc()
	}
	WriteJSON(w, http.StatusOK, mapGuest(updated))
}

func (h *Handler) HandleDeleteGuest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	guest, err := h.Service.GetGuestRequest(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && guest.ResidentUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.Service.SoftDeleteGuestRequest(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Guest.WithLabelValues("deleted").Inc()
	}
	WriteJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) HandleRestoreGuest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	role := roleFromContext(r)
	actorID := actorFromContext(r)
	guest, err := h.Service.GetGuestRequestAny(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if role == string(auth.RoleResident) && guest.ResidentUserID != actorID {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if role == string(auth.RoleGuard) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}
	if err := h.Service.RestoreGuestRequest(r.Context(), id, actorID); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.Metrics != nil {
		h.Metrics.Guest.WithLabelValues("restored").Inc()
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "restored"})
}

func mapUser(user repo.User) UserResponse {
	resp := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if user.PlotNumber.Valid {
		resp.PlotNumber = &user.PlotNumber.String
	}
	if user.BlockedAt.Valid {
		resp.BlockedAt = &user.BlockedAt.Time
	}
	if user.CreatedBy.Valid {
		resp.CreatedBy = &user.CreatedBy.UUID
	}
	if user.UpdatedBy.Valid {
		resp.UpdatedBy = &user.UpdatedBy.UUID
	}
	if user.DeletedAt.Valid {
		resp.DeletedAt = &user.DeletedAt.Time
	}
	return resp
}

func mapPass(pass repo.Pass) PassResponse {
	resp := PassResponse{
		ID:          pass.ID,
		OwnerUserID: pass.OwnerUserID,
		PlateNumber: pass.PlateNumber,
		Status:      pass.Status,
		CreatedAt:   pass.CreatedAt,
		UpdatedAt:   pass.UpdatedAt,
	}
	if pass.VehicleBrand.Valid {
		resp.VehicleBrand = &pass.VehicleBrand.String
	}
	if pass.VehicleColor.Valid {
		resp.VehicleColor = &pass.VehicleColor.String
	}
	if pass.CreatedBy.Valid {
		resp.CreatedBy = &pass.CreatedBy.UUID
	}
	if pass.UpdatedBy.Valid {
		resp.UpdatedBy = &pass.UpdatedBy.UUID
	}
	if pass.DeletedAt.Valid {
		resp.DeletedAt = &pass.DeletedAt.Time
	}
	return resp
}

func mapPassWithOwner(pass repo.Pass, owner *repo.User) PassResponse {
	resp := mapPass(pass)
	if owner == nil {
		return resp
	}
	resp.OwnerFullName = &owner.FullName
	if owner.PlotNumber.Valid {
		resp.OwnerPlotNumber = &owner.PlotNumber.String
	}
	return resp
}

func mapPasses(passes []repo.Pass) []PassResponse {
	resp := make([]PassResponse, 0, len(passes))
	for _, pass := range passes {
		resp = append(resp, mapPass(pass))
	}
	return resp
}

func mapGuest(guest repo.GuestRequest) GuestResponse {
	resp := GuestResponse{
		ID:             guest.ID,
		ResidentUserID: guest.ResidentUserID,
		GuestFullName:  guest.GuestFullName,
		PlateNumber:    guest.PlateNumber,
		ValidFrom:      guest.ValidFrom,
		ValidTo:        guest.ValidTo,
		Status:         guest.Status,
		CreatedAt:      guest.CreatedAt,
		UpdatedAt:      guest.UpdatedAt,
	}
	if guest.CreatedBy.Valid {
		resp.CreatedBy = &guest.CreatedBy.UUID
	}
	if guest.UpdatedBy.Valid {
		resp.UpdatedBy = &guest.UpdatedBy.UUID
	}
	if guest.DeletedAt.Valid {
		resp.DeletedAt = &guest.DeletedAt.Time
	}
	return resp
}

func mapGuests(guests []repo.GuestRequest) []GuestResponse {
	resp := make([]GuestResponse, 0, len(guests))
	for _, guest := range guests {
		resp = append(resp, mapGuest(guest))
	}
	return resp
}

func actorFromContext(r *http.Request) uuid.UUID {
	id, ok := identity.UserIDFrom(r.Context())
	if !ok {
		return uuid.Nil
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return uid
}

func roleFromContext(r *http.Request) string {
	role, _ := identity.RoleFrom(r.Context())
	return role
}

func parsePagination(r *http.Request) (int32, int32) {
	limit := int32(20)
	offset := int32(0)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			if value > 100 {
				value = 100
			}
			limit = int32(value)
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value >= 0 {
			offset = int32(value)
		}
	}
	return limit, offset
}

func toNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *value, Valid: true}
}
