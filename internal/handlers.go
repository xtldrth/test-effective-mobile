package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type R = map[string]any

type SubscriptionsHandler struct {
	Service SubscriptionsService
	Logger  *slog.Logger
}

func NewSubscriptionsHandler(service SubscriptionsService, logger *slog.Logger) SubscriptionsHandler {
	return SubscriptionsHandler{
		Service: service,
		Logger:  logger,
	}
}

func getStatusCode(err error) int {
	var errValidation ErrValidation
	var errNotFound ErrNotFound
	switch {
	case errors.As(err, &errValidation):
		return http.StatusBadRequest
	case errors.As(err, &errNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func parseUserID(r *http.Request) (uuid.UUID, error) {
	strID := r.URL.Query().Get("user_id")
	if strID == "" {
		return uuid.UUID{}, errors.New("user id must be provided")
	}
	return uuid.Parse(strID)
}

func (h SubscriptionsHandler) response(w http.ResponseWriter, obj any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		h.Logger.Error("json encode", slog.String("error", err.Error()))
	}
}
func (h SubscriptionsHandler) errorResponse(w http.ResponseWriter, message string, status int) {
	h.response(w, R{"error": message}, status)
}

func (h SubscriptionsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	strID := r.PathValue("id")
	id, err := uuid.Parse(strID)
	if err != nil {
		h.errorResponse(w, fmt.Errorf("invalid user id in path: %w", err).Error(), http.StatusBadRequest)
		return
	}
	sub, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		h.errorResponse(w, err.Error(), getStatusCode(err))
		return
	}
	h.response(w, R{"subscription": SubscriptionToResponse(sub)}, http.StatusOK)
}
func (h SubscriptionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		h.errorResponse(w, fmt.Errorf("invalid user id: %w", err).Error(), http.StatusBadRequest)
		return
	}
	queries := r.URL.Query()
	serviceName := queries.Get("service_name")
	if serviceName == "" {
		subs, err := h.Service.GetByUserID(r.Context(), userID)
		if err != nil {
			h.errorResponse(w, err.Error(), getStatusCode(err))
			return
		}
		subscriptions := make([]SubscriptionResponseDTO, 0, len(subs))
		for _, sub := range subs {
			subscriptions = append(subscriptions, SubscriptionToResponse(sub))
		}
		h.response(w, R{"subscriptions": subscriptions}, http.StatusOK)
		return
	}
	subs, err := h.Service.GetByUserIDAndServiceName(r.Context(), userID, serviceName)
	if err != nil {
		h.errorResponse(w, err.Error(), getStatusCode(err))
		return
	}
	subscriptions := make([]SubscriptionResponseDTO, 0, len(subs))
	for _, sub := range subs {
		subscriptions = append(subscriptions, SubscriptionToResponse(sub))
	}
	h.response(w, R{"subscriptions": subscriptions}, http.StatusOK)
}

func (h SubscriptionsHandler) GetPricesSum(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	userID, err := parseUserID(r)
	if err != nil {
		h.errorResponse(w, fmt.Errorf("invalid user id: %w", err).Error(), http.StatusBadRequest)
		return
	}
	serviceName := queries.Get("service_name")
	if serviceName == "" {
		h.errorResponse(w, "service name must be provided", http.StatusBadRequest)
		return
	}
	strFromDate := queries.Get("from")
	if strFromDate == "" {
		h.errorResponse(w, "from date should be provided", http.StatusBadRequest)
		return
	}
	fromDate, err := ParseDate(strFromDate)
	if err != nil {
		h.errorResponse(w, fmt.Errorf("date parse: %w", err).Error(), http.StatusBadRequest)
		return
	}
	var toDate time.Time
	strToDate := queries.Get("to")
	if strToDate == "" {
		toDate = time.Now()
	} else {
		toDate, err = ParseDate(strToDate)
		if err != nil {
			h.errorResponse(w, fmt.Errorf("date parse: %w", err).Error(), http.StatusBadRequest)
			return
		}
	}
	sum, err := h.Service.GetPricesSum(r.Context(), userID, serviceName, fromDate, toDate)
	if err != nil {
		h.errorResponse(w, err.Error(), getStatusCode(err))
		return
	}
	h.response(w, R{"total_spent": sum}, http.StatusOK)
}
func (h SubscriptionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var subscriptionCreate SubscriptionCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&subscriptionCreate); err != nil {
		h.errorResponse(w, fmt.Errorf("decode body: %w", err).Error(), http.StatusBadRequest)
		return
	}
	subscription, err := h.Service.Create(r.Context(), subscriptionCreate)
	if err != nil {
		h.errorResponse(w, err.Error(), getStatusCode(err))
		return
	}
	h.response(w, SubscriptionToResponse(subscription), http.StatusCreated)
}
func (h SubscriptionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// TODO: implement me
	panic("not implemented")
}
func (h SubscriptionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: implement me
	panic("not implemented")
}
