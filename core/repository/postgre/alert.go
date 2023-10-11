package postgre

import (
	"context"
	"fmt"
	"github.com/rhmdnrhuda/unified/core/entity"
	"github.com/rhmdnrhuda/unified/pkg/postgres"
	"strings"
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
	var values []string
	for _, alert := range data {
		values = append(values, fmt.Sprintf("('%s', '%d', '%s', '%s')", alert.UserID, alert.Date, alert.Message, alert.University))
	}

	sql := fmt.Sprintf("INSERT INTO public.alert (user_id, date, message, university) VALUES %s", strings.Join(values, ", "))

	_, err := t.Pool.Query(ctx, sql)
	if err != nil {
		return fmt.Errorf("AlertRepository - Create - t.Pool.Query: %w", err)
	}

	return nil
}

func (t *AlertRepository) FindAlert(ctx context.Context, day int64) ([]entity.AlertDBResponse, error) {
	now := time.Now().Unix()

	// Calculate the end date of the range.
	endDate := now + (day * 24 * 60 * 60)

	sql := fmt.Sprintf(`SELECT user_id, university, MIN(date) AS date, ARRAY_AGG(message) AS messages
			FROM alert
			WHERE date > %d AND date < %d
			GROUP BY user_id, university
			ORDER BY date ASC;`,
		now, endDate)

	// Execute the SQL statement and get the results.
	rows, err := t.Pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("AlertRepository - FindAlert - t.Pool.Query: %w", err)
	}

	// Scan the results into a slice of alerts.
	alerts := []entity.AlertDBResponse{}
	for rows.Next() {
		alert := entity.AlertDBResponse{}
		err := rows.Scan(&alert.UserID, &alert.University, &alert.Date, &alert.Messages)
		if err != nil {
			return nil, fmt.Errorf("AlertRepository - FindAlert - rows.Scan: %w", err)
		}

		alerts = append(alerts, alert)
	}

	// Close the rows.
	defer rows.Close()

	return alerts, nil
}
