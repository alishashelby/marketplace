package user

import (
	"errors"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepoPostgres_Save(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewUserRepoPostgres(mock)
	testUser := &entity.User{
		ID:       uuid.New(),
		Username: "testUser",
		Password: "test",
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO users").
			WithArgs(testUser.ID, testUser.Username, testUser.Password).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Save(testUser)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure", func(t *testing.T) {
		testErr := errors.New("test error")
		mock.ExpectExec("INSERT INTO users").
			WithArgs(testUser.ID, testUser.Username, testUser.Password).
			WillReturnError(testErr)

		err = repo.Save(testUser)

		assert.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// nolint:dupl
func TestUserRepoPostgres_GetByUsername(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewUserRepoPostgres(mock)
	testUser := &entity.User{
		ID:       uuid.New(),
		Username: "testUser",
		Password: "test",
	}

	t.Run("Success", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "username", "password"}).
			AddRow(testUser.ID, testUser.Username, testUser.Password)

		mock.ExpectQuery("SELECT id, username, password FROM users WHERE username = $1").
			WithArgs(testUser.Username).
			WillReturnRows(rows)

		user, err := repo.GetByUsername(testUser.Username)

		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure - not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, password FROM users WHERE username = $1").
			WithArgs(testUser.Username).
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.GetByUsername(testUser.Username)

		assert.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure - database error", func(t *testing.T) {
		testErr := errors.New("test error")
		mock.ExpectQuery("SELECT id, username, password FROM users WHERE username = $1").
			WithArgs(testUser.Username).
			WillReturnError(testErr)

		user, err := repo.GetByUsername(testUser.Username)

		assert.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// nolint:dupl
func TestUserRepoPostgres_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewUserRepoPostgres(mock)
	testUser := &entity.User{
		ID:       uuid.New(),
		Username: "testUser",
		Password: "test",
	}

	t.Run("Success", func(t *testing.T) {
		rows := mock.NewRows([]string{"id", "username", "password"}).
			AddRow(testUser.ID, testUser.Username, testUser.Password)
		mock.ExpectQuery("SELECT id, username, password FROM users WHERE id = $1").
			WithArgs(testUser.ID).
			WillReturnRows(rows)

		user, err := repo.GetByID(testUser.ID)

		assert.NoError(t, err)
		assert.Equal(t, testUser, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure - not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, password FROM users WHERE id = $1").
			WithArgs(testUser.ID).
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.GetByID(testUser.ID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Failure - database error", func(t *testing.T) {
		testErr := errors.New("test error")
		mock.ExpectQuery("SELECT id, username, password FROM users WHERE id = $1").
			WithArgs(testUser.ID).
			WillReturnError(testErr)

		user, err := repo.GetByID(testUser.ID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
