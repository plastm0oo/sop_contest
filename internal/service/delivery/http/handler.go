package deliveryhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/plastm0oo/sop_contest/internal/service"
)

type handler struct {
	uc        service.UseCase
	jwtSecret string
}

func New(uc service.UseCase, jwtSecret string) service.Handler {
	return &handler{
		uc:        uc,
		jwtSecret: jwtSecret,
	}
}
func (h *handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/teachers", h.ListTeachers)
	mux.HandleFunc("/api/auth/register", h.Register)
	mux.HandleFunc("/api/auth/login", h.Login)
	mux.Handle("/api/feedbacks", h.authMiddleware(http.HandlerFunc(h.CreateFeedback)))
	mux.Handle("/api/feedbacks/me", h.authMiddleware(http.HandlerFunc(h.ListMyFeedbacks)))
}

func (h *handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resp := h.uc.Health(r.Context())
	writeJSON(w, http.StatusOK, resp)
}

func (h *handler) ListTeachers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query()

	limit, ok := parseIntQuery(query.Get("limit"), 20)
	if !ok || limit < 1 || limit > 100 {
		writeValidationError(w, map[string]string{
			"limit": "must be an integer from 1 to 100",
		})
		return
	}

	offset, ok := parseIntQuery(query.Get("offset"), 0)
	if !ok || offset < 0 {
		writeValidationError(w, map[string]string{
			"offset": "must be an integer greater than or equal to 0",
		})
		return
	}

	params := service.TeacherListParams{
		Q:       query.Get("q"),
		Faculty: query.Get("faculty"),
		Limit:   limit,
		Offset:  offset,
	}

	resp, err := h.uc.ListTeachers(r.Context(), params)
	if err != nil {
		log.Printf("list teachers failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func parseIntQuery(raw string, fallback int) (int, bool) {
	if raw == "" {
		return fallback, true
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}

	return value, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write json response failed: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeValidationError(w http.ResponseWriter, details map[string]string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"error":   "validation failed",
		"details": details,
	})
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.uc.Register(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.uc.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleServiceError(w http.ResponseWriter, err error) {
	var validationErr service.ValidationError

	switch {
	case errors.As(err, &validationErr):
		writeValidationError(w, validationErr.Details)

	case errors.Is(err, service.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "пользователь с таким email уже существует")

	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "неверный email или пароль")

	case errors.Is(err, service.ErrAccountBlocked):
		writeError(w, http.StatusForbidden, "аккаунт заблокирован")

	case errors.Is(err, service.ErrFeedbackAlreadyExists):
		writeError(w, http.StatusConflict, "вы уже оставляли отзыв на этого преподавателя")

	case errors.Is(err, service.ErrTeacherNotFound):
		writeError(w, http.StatusNotFound, "преподаватель не найден")

	default:
		log.Printf("internal service error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *handler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "пользователь не авторизован")
		return
	}

	var req service.FeedbackCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	resp, err := h.uc.CreateFeedback(r.Context(), userID, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *handler) ListMyFeedbacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "пользователь не авторизован")
		return
	}

	resp, err := h.uc.ListMyFeedbacks(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
