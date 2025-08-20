package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alishashelby/marketplace/internal/application/dto"
	"github.com/alishashelby/marketplace/internal/application/service"
	"github.com/alishashelby/marketplace/internal/application/validator"
	"github.com/alishashelby/marketplace/pkg"
	"log"
	"net/http"
)

type UserController struct {
	userService *service.UserService
	validator   *validator.UserValidator
}

func NewUserController(userService *service.UserService,
	validator *validator.UserValidator) *UserController {
	return &UserController{
		userService: userService,
		validator:   validator,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account and returns a JWT token in the Authorization header
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserDTO true "User credentials"
// @Success 201 {object} dto.UserDTO
// @Header 201 {string} Authorization "Bearer token"
// @Failure 400 {object} pkg.ValidationErrorResponse "Validation or parsing error"
// @Failure 409 {object} pkg.ErrorResponse "User already exists"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error"
// @Router /api/register [post]
func (s *UserController) Register(w http.ResponseWriter, r *http.Request) {
	log.Println("UserController.Register called")

	userDTO := dto.UserDTO{}
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		log.Print("UserController.Register parsing error:", err)
		pkg.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	errs := s.validator.Validate(userDTO)
	if errs != nil {
		pkg.SendValidationError(w, http.StatusBadRequest, errs)
		return
	}

	token, err := s.userService.Register(userDTO.Username, userDTO.Password)
	if err != nil {
		log.Print("UserController.Register service error:", err)
		s.handleUserError(w, err)
		return
	}

	s.setToken(w, *token)
	pkg.SendJSON(w, http.StatusCreated, userDTO)
}

// Login godoc
// @Summary Authenticate user
// @Description Authenticates the user and returns a JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserDTO true "User credentials"
// @Success 200 {object} map[string]interface{} "token"
// @Header 200 {string} Authorization "Bearer token"
// @Failure 400 {object} pkg.ErrorResponse "Invalid request"
// @Failure 404 {object} pkg.ErrorResponse "User not found"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error"
// @Router /api/login [post]
func (s *UserController) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("UserController.Login called")

	userDTO := dto.UserDTO{}
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		log.Print("UserController.Login parsing error:", err)
		pkg.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := s.userService.Login(userDTO.Username, userDTO.Password)
	if err != nil {
		log.Print("UserController.Login service error:", err)
		s.handleUserError(w, err)
		return
	}

	s.setToken(w, *token)
	pkg.SendJSON(w, http.StatusOK, map[string]interface{}{"token": *token})
}

func (s *UserController) handleUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrorUserExists):
		pkg.SendError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrorUserWithUsernameDoesNotExists):
		pkg.SendError(w, http.StatusNotFound, err.Error())
	default:
		pkg.SendError(w, http.StatusInternalServerError, err.Error())
	}
}

func (s *UserController) setToken(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
