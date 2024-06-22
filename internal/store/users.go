package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *Postgres) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	query := `
			INSERT INTO users (id, company, role, name, surname, phone, email, user_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id, company, role, name, surname, phone, email, user_type
			`

	err := p.db.QueryRow(
		ctx,
		query,
		uuid.New(),
		user.Company,
		user.Role,
		user.Name,
		user.Surname,
		user.Phone,
		user.Email,
		user.UserType,
	).Scan(
		&user.ID,
		&user.Company,
		&user.Role,
		&user.Name,
		&user.Surname,
		&user.Phone,
		&user.Email,
		&user.UserType,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, models.ErrDuplicateUser
		}

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return nil, models.ErrCompanyNotFound
		}

		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}

func (p *Postgres) GetUsers(ctx context.Context, params models.GetParams) ([]*models.User, error) {
	users := make([]*models.User, 0, 1)

	query := `
			SELECT id, company, role, name, surname, phone, email, user_type
			FROM users
			`

	if params.Filter != "" {
		query += fmt.Sprintf(" WHERE name LIKE '%%%s%%'", params.Filter)
	}

	if params.Sorting != "" {
		query += " ORDER BY " + params.Sorting
		if params.Descending {
			query += " DESC"
		}
	}

	query += fmt.Sprintf(" OFFSET %d LIMIT %d", params.Offset, params.Limit)

	rows, err := p.db.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := new(models.User)

		err := rows.Scan(
			&user.ID,
			&user.Company,
			&user.Role,
			&user.Name,
			&user.Surname,
			&user.Phone,
			&user.Email,
			&user.UserType,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting users: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := new(models.User)

	query := `
		SELECT id, company, role, name, surname, phone, email, user_type
		FROM users
		WHERE id = $1
		`

	err := p.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Company,
		&user.Role,
		&user.Name,
		&user.Surname,
		&user.Phone,
		&user.Email,
		&user.UserType,
	)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, models.ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

func (p *Postgres) UpdateUser(ctx context.Context, id uuid.UUID, user models.UpdateUserRequest) (*models.User, error) {
	changedUser, err := p.GetUserByID(ctx, id)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, models.ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	editedFlag := false

	if user.Company != uuid.Nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET company = $2 WHERE id = $1 RETURNING company`,
			id,
			user.Company,
		).Scan(
			&changedUser.Company,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userCompany: %w", err)
		}
	}

	if user.Role != nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET role = $2 WHERE id = $1 RETURNING role`,
			id,
			user.Role,
		).Scan(
			&changedUser.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userRole: %w", err)
		}
	}

	if user.Name != nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET name = $2 WHERE id = $1 RETURNING name`,
			id,
			user.Name,
		).Scan(
			&changedUser.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userName: %w", err)
		}
	}

	if user.Surname != nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET surname = $2 WHERE id = $1 RETURNING surname`,
			id,
			user.Surname,
		).Scan(
			&changedUser.Surname,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userSurname: %w", err)
		}
	}

	if user.Phone != nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET phone = $2 WHERE id = $1 RETURNING phone`,
			id,
			user.Phone,
		).Scan(
			&changedUser.Phone,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userPhone: %w", err)
		}
	}

	if user.Email != nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET email = $2 WHERE id = $1 RETURNING email`,
			id,
			user.Email,
		).Scan(
			&changedUser.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userEmail: %w", err)
		}
	}

	if user.UserType != nil {
		editedFlag = true

		err := p.db.QueryRow(
			ctx,
			`UPDATE users SET user_type = $2 WHERE id = $1 RETURNING user_type`,
			id,
			user.UserType,
		).Scan(
			&changedUser.UserType,
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userUserType: %w", err)
		}
	}

	if editedFlag {
		_, err = p.db.Exec(
			ctx,
			`UPDATE users SET updated_at = $2 WHERE id = $1 RETURNING updated_at`,
			id,
			time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("error updating userUserUpdatedAt: %w", err)
		}

		return changedUser, nil
	}

	return nil, models.ErrEmptyRequest
}

func (p *Postgres) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1
		`

	result, err := p.db.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrUserNotFound
	}

	return nil
}
