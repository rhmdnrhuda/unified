package postgre

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/rhmdnrhuda/unified/core/entity"
	"strings"
	"time"

	"github.com/rhmdnrhuda/unified/pkg/postgres"
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

func (t *TalentRepository) FindTalentByUniversityAndMajor(ctx context.Context, universities, majors []string) (entity.Talent, error) {
	talents := []entity.Talent{}

	sql := fmt.Sprintf(`
       SELECT name, major, status, university
       FROM %s
    `, talentRepositoryName)

	rows, err := t.Pool.Query(ctx, sql)
	if err != nil {
		return entity.Talent{}, fmt.Errorf("TalentRepository - FindTalentByUniversityAndMajor - t.Query: %w", err)
	}

	for rows.Next() {
		talent := entity.Talent{}
		err := rows.Scan(&talent.Name, &talent.Major, &talent.Status, &talent.University)
		if err != nil {
			return entity.Talent{}, fmt.Errorf("AlertRepository - FindAlert - rows.Scan: %w", err)
		}

		talents = append(talents, talent)
	}

	for _, talent := range talents {
		for _, university := range universities {
			if strings.EqualFold(talent.University, university) || strings.Contains(university, talent.University) || strings.Contains(talent.University, university) {
				return talent, nil
			}
		}

		for _, major := range majors {
			if strings.EqualFold(talent.Major, major) || strings.Contains(major, talent.Major) || strings.Contains(talent.Major, major) {
				return talent, nil
			}
		}
	}

	return entity.Talent{}, errors.New("empty result")
}
