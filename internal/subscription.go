package internal

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

func (s Subscription) Validate() error {
	if s.UserID == uuid.Nil {
		return ErrValidation{Msg: "user_id must be provided"}
	}
	if s.ServiceName == "" {
		return ErrValidation{Msg: "service_name must be provided"}
	}
	if s.Price < 0 {
		return ErrValidation{Msg: "price must be not negative integer"}
	}
	if s.StartDate.IsZero() {
		return ErrValidation{Msg: "start_date must be provided"}
	}
	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return ErrValidation{Msg: "end_date must be after start_date"}
	}
	return nil
}
