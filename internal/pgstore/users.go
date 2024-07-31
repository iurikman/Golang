package pgstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	rowCount = 2
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

	p.db.QueryRow(ctx, query, id)
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
	userToBeChanged, err := p.GetUserByID(ctx, id)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, models.ErrUserNotFound
	case err != nil:
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	value := reflect.ValueOf(user)
	typ := value.Type()
	queryPart1 := `UPDATE users` + " "
	queryPart2 := `SET` + " "
	queryPart3 := `WHERE id = $1 RETURNING id, company, role, name, surname, phone, email, user_type`
	row := rowCount

	var query string

	queryRowBody := []interface{}{id}

	for i := range make([]struct{}, typ.NumField()) {
		fieldName := typ.Field(i).Name

		fieldValue := value.Field(i)
		if !fieldValue.IsZero() && fieldValue.IsValid() {
			if fieldName == "UserType" {
				fieldName = "user_type"
			} else {
				fieldName = strings.ToLower(fieldName)
			}

			queryPart2 += fieldName + " = $" + strconv.Itoa(row) + ", "
			row++

			queryRowBody = append(queryRowBody, fieldValue.Interface())
		}
	}

	if row > rowCount {
		queryPart1 = strings.TrimSuffix(queryPart1, ", ")
		queryPart2 = strings.TrimSuffix(queryPart2, ", ")
		queryPart3 = strings.TrimSuffix(queryPart3, ", ")
		query = queryPart1 + " " + queryPart2 + " " + queryPart3

		err = p.db.QueryRow(ctx, query, queryRowBody...).Scan(
			&userToBeChanged.ID,
			&userToBeChanged.Company,
			&userToBeChanged.Role,
			&userToBeChanged.Name,
			&userToBeChanged.Surname,
			&userToBeChanged.Phone,
			&userToBeChanged.Email,
			&userToBeChanged.UserType,
		)

		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, models.ErrEmptyRequest
		case err != nil:
			return nil, fmt.Errorf("error updating user: %w", err)
		}

		return userToBeChanged, nil
	}

	return nil, models.ErrEmptyRequest
}

func (p *Postgres) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
				UPDATE users 
				SET deleted = true			
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
