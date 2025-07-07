package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/goto/compass/core/user"
	"github.com/goto/compass/internal/store/postgres"
	"github.com/goto/compass/internal/testutils"
	"github.com/goto/salt/log"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *postgres.Client
	repository *postgres.UserRepository
}

func (r *UserRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewNoop()
	r.client, err = newTestClient(r.T(), logger)
	if err != nil {
		r.T().Fatal(err)
	}

	r.ctx = context.TODO()
	r.repository, err = postgres.NewUserRepository(r.client)
	if err != nil {
		r.T().Fatal(err)
	}
}

func (r *UserRepositoryTestSuite) insertEmail(email string) error {
	query := fmt.Sprintf("insert into users (email) values ('%s')", email)
	if err := r.client.ExecQueries(context.Background(), []string{
		query,
	}); err != nil {
		return err
	}
	return nil
}

func (r *UserRepositoryTestSuite) TestCreate() {
	r.Run("return no error if successfully create user", func() {
		user := getUser("user@gotocompany.com")
		id, err := r.repository.Create(r.ctx, user)
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.Create(r.ctx, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		ud := getUser("user@gotocompany.com")
		id, err := r.repository.Create(r.ctx, ud)
		r.NoError(err)
		r.NotEmpty(id)

		id, err = r.repository.Create(r.ctx, ud)
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestCreateWithTx() {
	validUser := &user.User{
		Email:    "userWithTx@gotocompany.com",
		Provider: "compass",
	}
	r.Run("return no error if successfully create user with email", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUser)
			return err
		})
		r.NotEmpty(id)
		r.NoError(err)
	})

	r.Run("return ErrNilUser if user is nil", func() {
		var id string
		err := r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, nil)
			return err
		})
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if user is already exist", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		id, err := r.repository.Create(r.ctx, validUser)
		r.NoError(err)
		r.NotEmpty(id)

		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			var err error
			id, err = r.repository.CreateWithTx(r.ctx, tx, validUser)
			return err
		})
		r.ErrorAs(err, new(user.DuplicateRecordError))
		r.Empty(id)
	})
}

func (r *UserRepositoryTestSuite) TestGetBy() {
	r.Run("by email", func() {
		r.Run("return empty string and ErrNotFound if email not found in DB", func() {
			usr, err := r.repository.GetByEmail(r.ctx, "random")
			r.ErrorAs(err, new(user.NotFoundError))
			r.Empty(usr)
		})

		r.Run("return non empty user if email found in DB", func() {
			err := testutils.RunMigrationsWithClient(r.T(), r.client)
			r.NoError(err)

			user := getUser("use-getbyemail@gotocompany.com")
			id, err := r.repository.Create(r.ctx, user)
			r.NoError(err)
			r.NotEmpty(id)

			usr, err := r.repository.GetByEmail(r.ctx, user.Email)
			r.NoError(err)
			r.NotEmpty(usr)
		})
	})

	r.Run("by email with tx, return the user created in the tx", func() {
		err := testutils.RunMigrationsWithClient(r.T(), r.client)
		r.NoError(err)

		u := getUser("use-getbyemail@gotocompany.com")
		err = r.client.RunWithinTx(r.ctx, func(tx *sqlx.Tx) error {
			id, err := r.repository.CreateWithTx(r.ctx, tx, u)
			r.NoError(err)

			_, err = r.repository.GetByEmail(r.ctx, u.Email)
			r.ErrorAs(err, new(user.NotFoundError))

			u, err := r.repository.GetByEmailWithTx(r.ctx, tx, u.Email)
			r.NoError(err)
			r.Equal(id, u.ID)

			return nil
		})
		r.NoError(err)
	})
}

func (r *UserRepositoryTestSuite) TestGetOrInsertByEmail() {
	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.GetOrInsertByEmail(r.ctx, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return user ID if record already exist", func() {
		usr := &user.User{Email: "dummy@gotocompany.com"}

		err := r.insertEmail(usr.Email)
		r.NoError(err)

		id, err := r.repository.GetOrInsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)
	})

	r.Run("new row is inserted with email if user not exist", func() {
		usr := &user.User{Email: "user-upsert-1@gotocompany.com"}
		id, err := r.repository.GetOrInsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByEmail(r.ctx, usr.Email)
		r.NoError(err)
		r.Equal(gotUser.Email, usr.Email)
	})
}

func (r *UserRepositoryTestSuite) TestInsertByEmail() {
	r.Run("return ErrNoUserInformation if user is nil", func() {
		id, err := r.repository.InsertByEmail(r.ctx, nil)
		r.ErrorIs(err, user.ErrNoUserInformation)
		r.Empty(id)
	})

	r.Run("return ErrDuplicateRecord if record already exist", func() {
		usr := &user.User{Email: "dummy-insert@gotocompany.com"}

		err := r.insertEmail(usr.Email)
		r.NoError(err)

		id, err := r.repository.InsertByEmail(r.ctx, usr)
		r.ErrorIs(err, user.DuplicateRecordError{Email: usr.Email})
		r.Empty(id)
	})

	r.Run("new row is inserted with email if user not exist", func() {
		usr := &user.User{Email: "user-insert-1@gotocompany.com"}
		id, err := r.repository.InsertByEmail(r.ctx, usr)
		r.NoError(err)
		r.NotEmpty(id)

		gotUser, err := r.repository.GetByEmail(r.ctx, usr.Email)
		r.NoError(err)
		r.Equal(gotUser.Email, usr.Email)
	})
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, &UserRepositoryTestSuite{})
}
