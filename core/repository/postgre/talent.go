package postgre

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/temukan-co/monolith/core/entity"
	"time"

	"github.com/temukan-co/monolith/pkg/postgres"
)

func NewTalentRepository(pg *postgres.Postgres) *TalentRepository {
	return &TalentRepository{pg}
}

const (
	talentRepositoryName = "talent"
)

type TalentRepository struct {
	*postgres.Postgres
}

func (t *TalentRepository) Create(ctx context.Context, data *entity.Talent) error {
	now := time.Now().Unix()
	sql, args, err := t.Builder.
		Insert(talentRepositoryName).
		Columns(`name, calendar_url, university, status, major, created_at, updated_at, is_deleted`).
		Values(data.Name, data.CalendarURL, data.University, data.Status, data.Major, now, now, 0).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return fmt.Errorf("TalentRepository - Create - t.Builder: %w", err)
	}

	var lastInsertID int64
	err = t.Pool.QueryRow(ctx, sql, args...).Scan(&lastInsertID)
	if err != nil {
		return fmt.Errorf("TalentRepository - Create - t.Pool.QueryRow: %w", err)
	}

	data.ID = lastInsertID

	return nil
}

func (t *TalentRepository) Update(ctx context.Context, data *entity.Talent) error {
	now := time.Now().Unix()
	sql, args, err := t.Builder.
		Update(talentRepositoryName).
		Set("name", data.Name).
		Set("status", data.Status).
		Set("university", data.University).
		Set("major", data.Major).
		Set("calendar_url", data.CalendarURL).
		Set("updated_at", now).
		Where(squirrel.Eq{"id": data.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("TalentRepository - Update - t.Builder: %w", err)
	}

	_, err = t.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TalentRepository - Update - t.Pool.Exec: %w", err)
	}

	return nil
}

func (t *TalentRepository) FindTalentByUniversityAndMajor(ctx context.Context, university, major string) (entity.Talent, error) {
	talent := entity.Talent{}

	sql := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE university like $1 AND major like $1
		LIMIT 1;
	`, talentRepositoryName)

	row, err := t.Pool.Query(ctx, sql, university, major)
	if err != nil {
		return talent, fmt.Errorf("TalentRepository - FindTalentByUniversityAndMajor - t.Query: %w", err)
	}

	err = pgxscan.ScanOne(&talent, row)
	if err != nil {
		return talent, fmt.Errorf("TalentRepository - FindTalentByUniversityAndMajor - t.ScanOne: %w", err)
	}

	return talent, nil
}
