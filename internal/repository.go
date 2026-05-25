package internal

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionsRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Subscription, error)
	GetByUserIDAndServiceName(ctx context.Context, userID uuid.UUID, serviceName string) ([]Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	GetPricesSum(ctx context.Context, userID uuid.UUID, serviceName string, from time.Time, to time.Time) (int, error)
	Create(ctx context.Context, subscription Subscription) (Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, fieldValue map[string]any) (Subscription, error)
}

var _ SubscriptionsRepository = (*postgresSubscriptionRepository)(nil)

func NewPSQLSubscriptionsRepository(pool *pgxpool.Pool) SubscriptionsRepository {
	return &postgresSubscriptionRepository{pool: pool}
}

type postgresSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func (r *postgresSubscriptionRepository) scan(row pgx.Row) (sub Subscription, err error) {
	err = r.checkErr(row.Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price, &sub.StartDate, &sub.EndDate))
	return
}

func (r *postgresSubscriptionRepository) scanRows(rows pgx.Rows) (subscriptions []Subscription, err error) {
	for rows.Next() {
		sub, err := r.scan(rows)
		if err != nil {
			return nil, r.checkErr(err)
		}
		subscriptions = append(subscriptions, sub)
	}
	return
}

func (r *postgresSubscriptionRepository) checkErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound{Resource: "subscription"}
	}
	return err
}

func (r *postgresSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (Subscription, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date FROM public.subscriptions WHERE id = $1`
	return r.scan(r.pool.QueryRow(ctx, query, id))
}

func (r *postgresSubscriptionRepository) GetByUserIDAndServiceName(ctx context.Context, userID uuid.UUID, serviceName string) ([]Subscription, error) {
	query := `
SELECT id, user_id, service_name, price, start_date, end_date 
FROM public.subscriptions 
WHERE user_id = $1 AND service_name = $2`
	rows, err := r.pool.Query(ctx, query, userID, serviceName)
	if err := r.checkErr(err); err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *postgresSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	query := `
SELECT id, user_id, service_name, price, start_date, end_date 
FROM public.subscriptions 
WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, r.checkErr(err)
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *postgresSubscriptionRepository) calculateSum(price int, startDate time.Time, endDate *time.Time, from, to time.Time) int {
	if startDate.After(from) {
		from = startDate
	}
	if endDate != nil && endDate.Before(to) {
		to = *endDate
	}
	month := int(to.Month()) - int(from.Month()) + 12*(to.Year()-from.Year())
	return month * price
}

func (r *postgresSubscriptionRepository) GetPricesSum(ctx context.Context, userID uuid.UUID, serviceName string, from, to time.Time) (int, error) {
	query := `
SELECT price, start_date, end_date
FROM public.subscriptions
WHERE user_id = @userID AND service_name = @serviceName 
	AND (
		start_date BETWEEN @fromDate AND @toDate 
		OR 
		end_date BETWEEN @fromDate AND @toDate
	)
`
	args := pgx.NamedArgs{"userID": userID, "serviceName": serviceName, "fromDate": from, "toDate": to}
	rows, err := r.pool.Query(
		ctx,
		query,
		args,
	)
	if err != nil {
		return 0, r.checkErr(err)
	}
	defer rows.Close()
	totalSum := 0
	for rows.Next() {
		var (
			price     int
			startDate time.Time
			endDate   *time.Time
		)
		if err := rows.Scan(&price, &startDate, &endDate); err != nil {
			return 0, fmt.Errorf("row scan: %w", err)
		}
		totalSum += r.calculateSum(price, startDate, endDate, from, to)
	}
	return totalSum, nil
}

func (r *postgresSubscriptionRepository) Create(ctx context.Context, subscription Subscription) (Subscription, error) {
	query := `
INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date)
VALUES (@userID, @serviceName, @price, @startDate, @endDate)
RETURNING id, user_id, service_name, price, start_date, end_date`
	args := pgx.NamedArgs{
		"userID":      subscription.UserID,
		"serviceName": subscription.ServiceName,
		"price":       subscription.Price,
		"startDate":   subscription.StartDate,
		"endDate":     subscription.EndDate,
	}
	dbSubscription, err := r.scan(r.pool.QueryRow(ctx, query, args))
	if err != nil {
		return Subscription{}, fmt.Errorf("scan row: %w", err)
	}
	return dbSubscription, nil
}

func (r *postgresSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM subscriptions WHERE id = $1", id)
	return r.checkErr(err)
}

func BuildUpdateQuery(id uuid.UUID, fieldValue map[string]any) (query string, args []any) {
	if len(fieldValue) == 0 {
		return "", nil
	}
	fields := make([]string, 0, len(fieldValue))
	for field := range fieldValue {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	sb := strings.Builder{}
	sb.WriteString("UPDATE subscriptions SET ")
	for i, field := range fields {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(field)
		sb.WriteString(" = $")
		sb.WriteString(strconv.Itoa(i + 1))
		args = append(args, fieldValue[field])
	}
	sb.WriteString(" WHERE id = $")
	sb.WriteString(strconv.Itoa(len(fields) + 1))
	args = append(args, id)
	sb.WriteString(" RETURNING id, user_id, service_name, price, start_date, end_date")
	return sb.String(), args
}

func (r *postgresSubscriptionRepository) Update(ctx context.Context, id uuid.UUID, fieldValue map[string]any) (Subscription, error) {
	query, args := BuildUpdateQuery(id, fieldValue)
	if query == "" {
		return Subscription{}, ErrValidation{Msg: "no field to update"}
	}
	dbSubscription, err := r.scan(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		return Subscription{}, fmt.Errorf("scan row: %w", err)
	}
	return dbSubscription, nil
}
