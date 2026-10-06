package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"leave-management/internal/auth"
	"leave-management/internal/model"
	"leave-management/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type API struct {
	store     *store.Store
	jwtSecret string
	jwtExpiry time.Duration
}

func New(dataStore *store.Store, jwtSecret string, jwtExpiry time.Duration) http.Handler {
	api := &API{store: dataStore, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", api.register)
	mux.HandleFunc("POST /api/auth/login", api.login)
	mux.Handle("POST /api/leaves", api.requireAuth(http.HandlerFunc(api.createLeave)))
	mux.Handle("GET /api/leaves/my", api.requireAuth(http.HandlerFunc(api.myLeaves)))
	mux.Handle("GET /api/admin/leaves", api.requireAuth(http.HandlerFunc(api.adminLeaves)))
	mux.Handle("PATCH /api/admin/leaves/{id}/approve", api.requireAuth(http.HandlerFunc(api.approveLeave)))
	mux.Handle("PATCH /api/admin/leaves/{id}/reject", api.requireAuth(http.HandlerFunc(api.rejectLeave)))
	mux.Handle("GET /api/admin/dashboard", api.requireAuth(http.HandlerFunc(api.adminDashboard)))
	return mux
}

func (api *API) register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string     `json:"name"`
		Email    string     `json:"email"`
		Password string     `json:"password"`
		Role     model.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || input.Email == "" || !strings.Contains(input.Email, "@") || len(input.Password) < 8 {
		writeError(w, http.StatusBadRequest, "name, valid email, and password with at least 8 characters are required")
		return
	}
	if input.Role == "" {
		input.Role = model.RoleEmployee
	}
	if input.Role != model.RoleEmployee && input.Role != model.RoleAdmin {
		writeError(w, http.StatusBadRequest, "role must be EMPLOYEE or ADMIN")
		return
	}

	leaveBalance := 0.0
	if input.Role == model.RoleEmployee {
		leaveBalance = 20
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	user := model.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		Role:         input.Role,
		LeaveBalance: leaveBalance,
	}
	if err := api.store.CreateUser(r.Context(), &user); err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			writeError(w, http.StatusConflict, "email is already registered")
			return
		}
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	token, err := auth.GenerateToken(user, api.jwtSecret, api.jwtExpiry)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

func (api *API) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, err := api.store.UserByEmail(r.Context(), input.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateToken(user, api.jwtSecret, api.jwtExpiry)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func (api *API) createLeave(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(roleKey).(model.Role)
	if !ok || role != model.RoleEmployee {
		writeError(w, http.StatusForbidden, "only employees can create leave requests")
		return
	}

	var input struct {
		LeaveType model.LeaveType `json:"leave_type"`
		DayType   model.DayType   `json:"day_type"`
		StartDate string          `json:"start_date"`
		EndDate   string          `json:"end_date"`
		Reason    string          `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	//TODO:  lets change it afterwards
	if input.LeaveType != model.LeaveTypeAnnual &&
		input.LeaveType != model.LeaveTypeCasual &&
		input.LeaveType != model.LeaveTypeSick {
		writeError(w, http.StatusBadRequest, "leave_type must be ANNUAL, CASUAL, or SICK")
		return
	}
	if input.DayType != model.DayTypeFull && input.DayType != model.DayTypeHalf {
		writeError(w, http.StatusBadRequest, "day_type must be FULL_DAY or HALF_DAY")
		return
	}

	startDate, err := time.Parse(model.DateFormat, input.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "start_date must use YYYY-MM-DD")
		return
	}
	endDate, err := time.Parse(model.DateFormat, input.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "end_date must use YYYY-MM-DD")
		return
	}
	if endDate.Before(startDate) {
		writeError(w, http.StatusBadRequest, "end_date cannot be before start_date")
		return
	}

	todayText := time.Now().UTC().Format(model.DateFormat)
	today, _ := time.Parse(model.DateFormat, todayText)
	if startDate.Before(today) {
		writeError(w, http.StatusBadRequest, "start_date cannot be in the past")
		return
	}
	if startDate.Year() != today.Year() || endDate.Year() != today.Year() {
		writeError(w, http.StatusBadRequest, "leave dates must be in the current year")
		return
	}
	if input.DayType == model.DayTypeHalf && !startDate.Equal(endDate) {
		writeError(w, http.StatusBadRequest, "half-day leave must start and end on the same date")
		return
	}

	input.Reason = strings.TrimSpace(input.Reason)
	if input.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	numberOfDays := endDate.Sub(startDate).Hours()/24 + 1
	if input.DayType == model.DayTypeHalf {
		numberOfDays = 0.5
	}

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	leave := model.LeaveRequest{
		UserID:       userID,
		LeaveType:    input.LeaveType,
		DayType:      input.DayType,
		StartDate:    startDate.Format(model.DateFormat),
		EndDate:      endDate.Format(model.DateFormat),
		NumberOfDays: numberOfDays,
		Reason:       input.Reason,
		Status:       model.StatusPending,
	}
	if err := api.store.CreateLeave(r.Context(), &leave); err != nil {
		switch {
		case errors.Is(err, store.ErrInsufficientLeave):
			writeError(w, http.StatusConflict, "not enough leave balance")
		case errors.Is(err, store.ErrOverlappingRequest):
			writeError(w, http.StatusConflict, "leave dates overlap an existing request")
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusUnauthorized, "user not found")
		default:
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "could not create leave request")
		}
		return
	}

	writeJSON(w, http.StatusCreated, leave)
}

func (api *API) myLeaves(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(roleKey).(model.Role)
	if !ok || role != model.RoleEmployee {
		writeError(w, http.StatusForbidden, "only employees can view their leave history")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "invalid user")
		return
	}

	leaves, err := api.store.UserLeaves(r.Context(), userID)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not fetch leave history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"leaves": leaves})
}

func (api *API) adminLeaves(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(roleKey).(model.Role)
	if !ok || role != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "only admins can view pending leave requests")
		return
	}

	leaves, err := api.store.PendingLeaves(r.Context())
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not fetch pending leave requests")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"leaves": leaves})
}

func (api *API) approveLeave(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(roleKey).(model.Role)
	if !ok || role != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "only admins can approve leave requests")
		return
	}

	leaveID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || leaveID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid leave request ID")
		return
	}
	adminID, ok := r.Context().Value(userIDKey).(int64)
	if !ok || adminID <= 0 {
		writeError(w, http.StatusUnauthorized, "invalid admin")
		return
	}

	leave, remainingBalance, err := api.store.ApproveLeave(r.Context(), leaveID, adminID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrLeaveNotFound):
			writeError(w, http.StatusNotFound, "leave request not found")
		case errors.Is(err, store.ErrLeaveNotPending):
			writeError(w, http.StatusConflict, "only pending leave requests can be approved")
		case errors.Is(err, store.ErrInsufficientLeave):
			writeError(w, http.StatusConflict, "employee does not have enough leave balance")
		case errors.Is(err, store.ErrNotFound), errors.Is(err, store.ErrInvalidEmployee):
			writeError(w, http.StatusConflict, "leave request does not belong to a valid employee")
		default:
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "could not approve leave request")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"leave":             leave,
		"remaining_balance": remainingBalance,
	})
}

func (api *API) rejectLeave(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(roleKey).(model.Role)
	if !ok || role != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "only admins can reject leave requests")
		return
	}

	leaveID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || leaveID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid leave request ID")
		return
	}
	adminID, ok := r.Context().Value(userIDKey).(int64)
	if !ok || adminID <= 0 {
		writeError(w, http.StatusUnauthorized, "invalid admin")
		return
	}

	leave, err := api.store.RejectLeave(r.Context(), leaveID, adminID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrLeaveNotFound):
			writeError(w, http.StatusNotFound, "leave request not found")
		case errors.Is(err, store.ErrLeaveNotPending):
			writeError(w, http.StatusConflict, "only pending leave requests can be rejected")
		default:
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "could not reject leave request")
		}
		return
	}

	writeJSON(w, http.StatusOK, leave)
}

func (api *API) adminDashboard(w http.ResponseWriter, r *http.Request) {
	role, ok := r.Context().Value(roleKey).(model.Role)
	if !ok || role != model.RoleAdmin {
		writeError(w, http.StatusForbidden, "only admins can view the dashboard")
		return
	}

	dashboard, err := api.store.Dashboard(r.Context())
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "could not fetch dashboard")
		return
	}

	writeJSON(w, http.StatusOK, dashboard)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
