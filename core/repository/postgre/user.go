package postgre

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/rhmdnrhuda/unified/core/entity"
	"github.com/rhmdnrhuda/unified/pkg/postgres"
	"time"
)

func NewUserRepository(pg *postgres.Postgres) *UserRepository {
	return &UserRepository{pg}
}

const (
	userRepositoryName = "public.user"
)

type UserRepository struct {
	*postgres.Postgres
}

func (t *UserRepository) Create(ctx context.Context, data *entity.User) error {
	now := time.Now().Unix()
	sql, args, err := t.Builder.
		Insert(userRepositoryName).
		Columns(`name, number, university_preferences, major_preferences, created_at, updated_at, is_deleted`).
		Values(data.Name, data.Number, data.UniversityPreferences, data.MajorPreferences, now, now, 0).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepository - Create - t.Builder: %w", err)
	}

	var lastInsertID int64
	err = t.Pool.QueryRow(ctx, sql, args...).Scan(&lastInsertID)
	if err != nil {
		return fmt.Errorf("UserRepository - Create - t.Pool.QueryRow: %w", err)
	}

	data.ID = lastInsertID

	return nil
}

func (t *UserRepository) Update(ctx context.Context, data *entity.User) error {
	now := time.Now().Unix()
	sql, args, err := t.Builder.
		Update(userRepositoryName).
		Set("university_preferences", data.UniversityPreferences).
		Set("major_preferences", data.MajorPreferences).
		Set("updated_at", now).
		Where(squirrel.Eq{"number": data.Number}).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepository - Update - t.Builder: %w", err)
	}

	_, err = t.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepository - Update - t.Pool.Exec: %w", err)
	}

	return nil
}

func (t *UserRepository) FindUserByNumber(ctx context.Context, number string) (entity.User, error) {
	user := entity.User{}

	sql := fmt.Sprintf(`
		SELECT name, number, university_preferences, major_preferences
		FROM %s
		WHERE number like $1
		LIMIT 1;
	`, userRepositoryName)

	row, err := t.Pool.Query(ctx, sql, number)
	if err != nil {
		return user, fmt.Errorf("UserRepository - FindUserByNumber - t.Query: %w", err)
	}

	err = pgxscan.ScanOne(&user, row)
	if err != nil {
		return user, fmt.Errorf("UserRepository - FindUserByNumber - t.ScanOne: %w", err)
	}

	return user, nil
}
