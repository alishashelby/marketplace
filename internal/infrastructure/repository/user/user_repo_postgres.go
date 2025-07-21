package user

import (
	"database/sql"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
)

type UserRepoPostgres struct {
	db *sql.DB
}

func NewUserRepoPostgres(db *sql.DB) *UserRepoPostgres {
	return &UserRepoPostgres{
		db: db,
	}
}

func (r *UserRepoPostgres) Save(user *entity.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (id, username, password) VALUES ($1, $2, $3)",
		user.ID, user.Username, user.Password)

	return err
}

func (r *UserRepoPostgres) GetByUsername(username string) (*entity.User, error) {
	var user entity.User

	err := r.db.QueryRow(
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
		"SELECT id, username, password FROM users WHERE id = $1",
		id).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
