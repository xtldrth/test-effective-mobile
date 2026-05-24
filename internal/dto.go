package internal

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Nullable[T any] struct {
	value  T
	isSet  bool
	isNull bool
}

type Date time.Time

func ParseDate(date string) (time.Time, error) {
	monthYear := strings.Split(date, "-")
	if len(monthYear) != 2 {
		return time.Time{}, ErrValidation{Msg: "invalid date format, valid date format: `12-2001"}
	}
	month := monthYear[0]
	year := monthYear[1]
	return time.Parse(time.DateOnly, year+"-"+month+"-01")
}

func NewDate(t time.Time) Date {
	return Date(t)
}

func (d Date) Month() time.Month {
	return time.Time(d).Month()
}

func (d Date) Year() int {
	return time.Time(d).Year()
}
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var rawDate string
	if err := json.Unmarshal(data, &rawDate); err != nil {
		return err
	}
	t, err := ParseDate(rawDate)
	if err != nil {
		return err
	}
	*d = NewDate(t)
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	return json.Marshal(fmt.Sprintf("%.2d-%d", t.Month(), t.Year()))
}

func NewNullable[T any](v T) Nullable[T] {
	return Nullable[T]{
		value:  v,
		isSet:  true,
		isNull: false,
	}
}

func NewNull[T any]() Nullable[T] {
	return Nullable[T]{
		isSet:  true,
		isNull: true,
	}
}

func (n *Nullable[T]) Value() T {
	return n.value
}

func (n *Nullable[T]) IsSet() bool {
	return n.isSet
}

func (n *Nullable[T]) IsNull() bool {
	return n.isNull
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.isNull = true
		n.isSet = true
		return nil
	}
	if err := json.Unmarshal(data, &n.value); err != nil {
		return err
	}
	n.isSet = true
	return nil
}

type SubscriptionResponseDTO struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	StartDate   Date      `json:"start_date"`
	EndDate     *Date     `json:"end_date,omitempty"`
}

func SubscriptionToResponse(s Subscription) SubscriptionResponseDTO {
	var endDate *Date
	if s.EndDate != nil {
		endDate = new(NewDate(*s.EndDate))
	}
	return SubscriptionResponseDTO{
		ID:          s.ID,
		UserID:      s.UserID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		StartDate:   NewDate(s.StartDate),
		EndDate:     endDate,
	}
}

type SubscriptionCreateDTO struct {
	UserID      uuid.UUID      `json:"user_id"`
	ServiceName string         `json:"service_name"`
	Price       int            `json:"price"`
	StartDate   Date           `json:"start_date"`
	EndDate     Nullable[Date] `json:"end_date,omitempty"`
}

type SubscriptionUpdateDTO struct {
	UserID      Nullable[uuid.UUID] `json:"user_id"`
	ServiceName Nullable[string]    `json:"service_name"`
	Price       Nullable[int]       `json:"price"`
	StartDate   Nullable[time.Time] `json:"start_date"`
	EndDate     Nullable[time.Time] `json:"end_date"`
}
