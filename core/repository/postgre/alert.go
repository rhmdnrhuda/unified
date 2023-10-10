package postgre

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/rhmdnrhuda/unified/core/entity"
	"github.com/rhmdnrhuda/unified/pkg/postgres"
	"time"
)

func NewAlertRepository(pg *postgres.Postgres) *AlertRepository {
	return &AlertRepository{pg}
}

const (
	alertRepositoryName = "public.alert"
)

type AlertRepository struct {
	*postgres.Postgres
}

func (t *AlertRepository) Create(ctx context.Context, data []entity.Alert) error {
	values := []interface{}{}
	for _, alert := range data {
		values = append(values, alert.UserID, alert.Date, alert.Message)
	}

	sql, args, err := t.Builder.
		Insert(alertRepositoryName).
		Columns(`user_id, date, message`).
		Values(values).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return fmt.Errorf("AlertRepository - Create - t.Builder: %w", err)
	}

	_, err = t.Pool.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("AlertRepository - Create - t.Pool.Query: %w", err)
	}

	return nil
}

func (t *AlertRepository) FindAlert(ctx context.Context, day int64) ([]entity.Alert, error) {
	now := time.Now().Unix()

	// Calculate the end date of the range.
	endDate := now + (day * 24 * 60 * 60)

	// Build the SQL statement.
	sql, args, err := t.Builder.
		Select("*").
		From(alertRepositoryName).
		Where(squirrel.And{
			squirrel.GtOrEq{"date": now},
			squirrel.Lt{"date": endDate},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AlertRepository - FindAlert - t.Builder: %w", err)
	}

	// Execute the SQL statement and get the results.
	rows, err := t.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("AlertRepository - FindAlert - t.Pool.Query: %w", err)
	}

	// Scan the results into a slice of alerts.
	alerts := []entity.Alert{}
	for rows.Next() {
		alert := entity.Alert{}
		err := rows.Scan(&alert.ID, &alert.UserID, &alert.Date, &alert.Message)
		if err != nil {
			return nil, fmt.Errorf("AlertRepository - FindAlert - rows.Scan: %w", err)
		}

		alerts = append(alerts, alert)
	}

	// Close the rows.
	defer rows.Close()

	return alerts, nil
}
