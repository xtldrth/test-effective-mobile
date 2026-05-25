package internal

import (
	"database/sql/driver"
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

func (n *Nullable[T]) Value() (driver.Value, error) {
	if n.IsNull() {
		return (*T)(nil), nil
	}
	return n.value, nil
}

func (n *Nullable[T]) GetValue() T {
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

type Date time.Time

func ParseDate(date string) (time.Time, error) {
	monthYear := strings.Split(date, "-")
	if len(monthYear) != 2 {
		return time.Time{}, ErrValidation{Msg: "invalid date format, valid date format: `MM-YYYY"}
	}
	month := monthYear[0]
	year := monthYear[1]
	return time.Parse(time.DateOnly, year+"-"+month+"-01")
}

func NewDate(t time.Time) Date {
	return Date(t)
}

func (d Date) String() string {
	return fmt.Sprintf("%.2d-%d", d.Month(), d.Year())
}

func (d Date) Month() time.Month {
	return time.Time(d).Month()
}

func (d Date) After(u Date) bool {
	return time.Time(d).After(time.Time(u))
}

func (d Date) Before(u Date) bool {
	return time.Time(d).Before(time.Time(u))
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
	return json.Marshal(d.String())
}

func (d Date) Value() (driver.Value, error) {
	if time.Time(d).IsZero() {
		return (*time.Time)(nil), nil
	}
	return time.Time(d), nil
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
	StartDate   Nullable[Date]      `json:"start_date"`
	EndDate     Nullable[Date]      `json:"end_date"`
}

func (s SubscriptionUpdateDTO) Validate() error {
	if s.Price.IsSet() && s.Price.GetValue() < 0 {
		return ErrValidation{Msg: "price should be non negative integer"}
	}
	if s.ServiceName.IsSet() && len(s.ServiceName.GetValue()) == 0 {
		return ErrValidation{Msg: "service name were not provided"}
	}
	if s.StartDate.IsSet() && s.EndDate.IsSet() && !s.EndDate.IsNull() && s.StartDate.GetValue().After(s.EndDate.GetValue()) {
		return ErrValidation{Msg: "start date should be before end date"}
	}
	return nil
}
