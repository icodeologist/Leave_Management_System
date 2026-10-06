package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
	return mux
}

func (api *API) register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
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
		Role:         model.RoleEmployee,
		LeaveBalance: 20,
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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
