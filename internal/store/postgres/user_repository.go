package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/goto/compass/core/user"
	"github.com/jmoiron/sqlx"
)

// UserRepository is a type that manages user operation to the primary database
type UserRepository struct {
	client *Client
}

// InsertByEmail insert a new row if email not found
func (r *UserRepository) InsertByEmail(ctx context.Context, ud *user.User) (string, error) {
	var userID string

	if err := ud.Validate(); err != nil {
		return "", err
	}

	um := newUserModel(ud)

	if err := r.client.db.QueryRowxContext(ctx, `
				INSERT INTO users (email, provider) VALUES ($1, $2)
				RETURNING id
		`, um.Email, um.Provider).Scan(&userID); err != nil {
		err := checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", user.DuplicateRecordError{Email: ud.Email}
		}
		return "", err
	}
	if userID == "" {
		return "", fmt.Errorf("error User ID is empty from DB")
	}
	return userID, nil
}

// Create insert a user to the database
// returns error if email is empty
func (r *UserRepository) Create(ctx context.Context, ud *user.User) (string, error) {
	return r.create(ctx, r.client.db, ud)
}

// Create insert a user to the database using given transaction as client
func (r *UserRepository) CreateWithTx(ctx context.Context, tx *sqlx.Tx, ud *user.User) (string, error) {
	return r.create(ctx, tx, ud)
}

func (r *UserRepository) create(ctx context.Context, querier sqlx.QueryerContext, ud *user.User) (string, error) {
	var userID string

	if ud == nil {
		return "", user.ErrNoUserInformation
	}

	if ud.Email == "" {
		return "", user.ErrNoUserInformation
	}

	um := newUserModel(ud)

	if err := querier.QueryRowxContext(ctx, `
					INSERT INTO
					users
						(email, provider)
					VALUES
						($1, $2)
					RETURNING id
					`, um.Email, um.Provider).Scan(&userID); err != nil {
		err := checkPostgresError(err)
		if errors.Is(err, errDuplicateKey) {
			return "", user.DuplicateRecordError{Email: ud.Email}
		}
		return "", err
	}
	if userID == "" {
		return "", fmt.Errorf("error User ID is empty from DB")
	}
	return userID, nil
}

// GetByEmail retrieves user by given the email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	return r.GetByEmailWithTx(ctx, r.client.db, email)
}

func (r *UserRepository) GetByEmailWithTx(ctx context.Context, querier sqlx.QueryerContext, email string) (user.User, error) {
	u, err := getUserByPredicate(ctx, querier, sq.Eq{"email": email})
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.NotFoundError{Email: email}
	}
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepository) GetOrInsertByEmail(ctx context.Context, ud *user.User) (string, error) {
	var userID string
	if err := ud.Validate(); err != nil {
		return "", err
	}
	um := newUserModel(ud)

	email := um.Email.String
	err := r.client.RunWithinTx(ctx, func(*sqlx.Tx) error {
		usr, err := r.GetByEmail(ctx, email)
		if err == nil {
			userID = usr.ID
			return nil
		}

		userID, err = r.InsertByEmail(ctx, &user.User{Email: email})
		return err
	})
	return userID, err
}

func getUserByPredicate(ctx context.Context, querier sqlx.QueryerContext, pred sq.Eq) (user.User, error) {
	qry, args, err := sq.Select("id", "email", "provider", "created_at", "updated_at").
		From("users").
		Where(pred).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return user.User{}, fmt.Errorf("build query to get user by predicate: %w", err)
	}
	var um UserModel
	if err := sqlx.GetContext(ctx, querier, &um, qry, args...); err != nil {
		return user.User{}, fmt.Errorf("get user by predicate: %w", err)
	}

	return um.toUser(), nil
}

// NewUserRepository initializes user repository clients
func NewUserRepository(c *Client) (*UserRepository, error) {
	if c == nil {
		return nil, errNilPostgresClient
	}
	return &UserRepository{
		client: c,
	}, nil
}
