package internal

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type SubscriptionsService interface {
	GetByID(ctx context.Context, id uuid.UUID) (Subscription, error)
	GetByUserIDAndServiceName(ctx context.Context, userID uuid.UUID, serviceName string) ([]Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	GetPricesSum(ctx context.Context, userID uuid.UUID, serviceName string, from time.Time, to time.Time) (int, error)
	Create(ctx context.Context, subscription SubscriptionCreateDTO) (Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, sub SubscriptionUpdateDTO) (Subscription, error)
}

func NewSubscriptionsService(repository SubscriptionsRepository, timeout time.Duration, logger *slog.Logger) SubscriptionsService {
	return &subscriptionsService{
		repository: repository,
		timeout:    timeout,
		logger:     logger,
	}
}

var _ SubscriptionsService = (*subscriptionsService)(nil)

type subscriptionsService struct {
	repository SubscriptionsRepository
	timeout    time.Duration
	logger     *slog.Logger
}

func (s *subscriptionsService) GetByID(ctx context.Context, id uuid.UUID) (Subscription, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	subscription, err := s.repository.GetByID(c, id)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return Subscription{}, err
		}
		s.logger.Error(
			"get by id",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)
		return Subscription{}, ErrInternal
	}
	return subscription, nil
}

func (s *subscriptionsService) GetByUserIDAndServiceName(ctx context.Context, userID uuid.UUID, serviceName string) ([]Subscription, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	subscriptions, err := s.repository.GetByUserIDAndServiceName(c, userID, serviceName)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return nil, err
		}
		s.logger.Error(
			"get by user id and service name",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
			slog.String("service_name", serviceName),
		)
		return nil, ErrInternal
	}
	return subscriptions, nil
}

func (s *subscriptionsService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	subscriptions, err := s.repository.GetByUserID(c, userID)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return nil, err
		}
		s.logger.Error(
			"get by user id",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
		return nil, ErrInternal
	}
	return subscriptions, nil
}

func (s *subscriptionsService) GetPricesSum(ctx context.Context, userID uuid.UUID, serviceName string, from time.Time, to time.Time) (int, error) {
	if from.After(to) {
		return 0, ErrValidation{Msg: "invalid date range, `from` shouldn't be after `to`"}
	}
	if serviceName == "" {
		return 0, ErrValidation{Msg: "service name should be provided"}
	}
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	subscriptions, err := s.repository.GetPricesSum(c, userID, serviceName, from, to)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return 0, err
		}
		s.logger.Error(
			"get prices sum",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
		return 0, ErrInternal
	}
	return subscriptions, nil
}

func (s *subscriptionsService) Create(ctx context.Context, subscription SubscriptionCreateDTO) (Subscription, error) {
	var endDate *time.Time
	if subscription.EndDate.IsSet() && !subscription.EndDate.IsNull() {
		endDate = new(time.Time(subscription.EndDate.GetValue()))
	}
	newSubscription := Subscription{
		UserID:      subscription.UserID,
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		StartDate:   time.Time(subscription.StartDate),
		EndDate:     endDate,
	}
	if err := newSubscription.Validate(); err != nil {
		return Subscription{}, err
	}
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	dbSubscription, err := s.repository.Create(c, newSubscription)
	if err != nil {
		s.logger.Error(
			"create",
			slog.String("error", err.Error()),
			slog.Any("subscription", subscription),
		)
		return Subscription{}, ErrInternal
	}
	return dbSubscription, nil
}

func (s *subscriptionsService) Delete(ctx context.Context, id uuid.UUID) error {
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	err := s.repository.Delete(c, id)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return err
		}
		s.logger.Error(
			"delete",
			slog.String("error", err.Error()),
		)
		return ErrInternal
	}
	return nil
}

func (s *subscriptionsService) Validate() {

}

func (s *subscriptionsService) Update(ctx context.Context, id uuid.UUID, sub SubscriptionUpdateDTO) (Subscription, error) {
	if err := sub.Validate(); err != nil {
		return Subscription{}, err
	}
	c, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	dbSub, err := s.repository.GetByID(c, id)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return Subscription{}, err
		}
		s.logger.Error("getting user from repository", slog.String("error", err.Error()))
		return Subscription{}, ErrInternal
	}
	if sub.StartDate.IsSet() && dbSub.EndDate != nil && dbSub.EndDate.After(time.Time(sub.StartDate.GetValue())) {
		return Subscription{}, ErrValidation{Msg: "start date should be after end date"}
	}
	if sub.EndDate.IsSet() && !sub.EndDate.IsNull() && dbSub.StartDate.After(time.Time(sub.EndDate.GetValue())) {
		return Subscription{}, ErrValidation{Msg: "end date should be before start date"}
	}
	c, cancel = context.WithTimeout(ctx, s.timeout)
	defer cancel()
	updatedSub, err := s.repository.Update(c, id, sub)
	if err != nil {
		if _, ok := errors.AsType[ErrNotFound](err); ok {
			return Subscription{}, err
		}
		s.logger.Error(
			"update",
			slog.String("error", err.Error()),
		)
		return Subscription{}, ErrInternal
	}
	return updatedSub, nil
}
