package user

import (
	"context"

	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxPool interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Close()
	Ping(context.Context) error
}

type UserRepoPostgres struct {
	db PgxPool
}

func NewUserRepoPostgres(db PgxPool) *UserRepoPostgres {
	return &UserRepoPostgres{
		db: db,
	}
}

func (r *UserRepoPostgres) Save(user *entity.User) error {
	_, err := r.db.Exec(
		context.Background(),
		"INSERT INTO users (id, username, password) VALUES ($1, $2, $3)",
		user.ID, user.Username, user.Password)

	return err
}

func (r *UserRepoPostgres) GetByUsername(username string) (*entity.User, error) {
	var user entity.User

	err := r.db.QueryRow(
		context.Background(),
		"SELECT id, username, password FROM users WHERE username = $1",
		username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepoPostgres) GetByID(id uuid.UUID) (*entity.User, error) {
	var user entity.User

	err := r.db.QueryRow(
		context.Background(),
		"SELECT id, username, password FROM users WHERE id = $1",
		id).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
