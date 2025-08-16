package service

import (
	"errors"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorUserExists                    = errors.New("user with this username already exists")
	ErrorUserWithUsernameDoesNotExists = errors.New("user with this username does not exist")
	ErrorUserWithIDDoesNotExists       = errors.New("user with this id does not exist")
	ErrorInvalidPassword               = errors.New("invalid password")
)

//go:generate mockgen -source=user_service.go -destination=user_repo_mock.go -package=service UserRepository
type UserRepository interface {
	Save(user *entity.User) error
	GetByUsername(username string) (*entity.User, error)
	GetByID(id uuid.UUID) (*entity.User, error)
}

type UserService struct {
	repo       UserRepository
	jwtService *JWTService
}

func NewUserService(repo UserRepository, service *JWTService) *UserService {
	return &UserService{
		repo:       repo,
		jwtService: service,
	}
}

func (s *UserService) Register(username, password string) (*string, error) {
	if user, err := s.repo.GetByUsername(username); err == nil && user != nil {
		return nil, ErrorUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:       uuid.New(),
		Username: username,
		Password: string(hash),
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	return s.jwtService.GenerateToken(user)
}

func (s *UserService) Login(username, password string) (*string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, ErrorUserWithUsernameDoesNotExists
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrorInvalidPassword
	}

	return s.jwtService.GenerateToken(user)
}

func (s *UserService) GetByID(id uuid.UUID) (*entity.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrorUserWithIDDoesNotExists
	}

	return user, nil
}
