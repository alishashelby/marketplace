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

func (s *UserController) Register(w http.ResponseWriter, r *http.Request) {
	log.Println("UserController.Register called")

	userDTO := dto.UserDTO{}
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		log.Print("UserController.Register parsing error:", err)
		pkg.SendJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	errs := s.validator.Validate(userDTO)
	if errs != nil {
		pkg.SendJSON(w, http.StatusBadRequest, errs)
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

func (s *UserController) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("UserController.Login called")

	userDTO := dto.UserDTO{}
	if err := json.NewDecoder(r.Body).Decode(&userDTO); err != nil {
		log.Print("UserController.Login parsing error:", err)
		pkg.SendJSON(w, http.StatusBadRequest, err.Error())
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
	mappedError := map[string]interface{}{"error": err.Error()}

	switch {
	case errors.Is(err, service.ErrorUserExists):
		pkg.SendJSON(w, http.StatusConflict, mappedError)
	case errors.Is(err, service.ErrorUserWithUsernameDoesNotExists):
		pkg.SendJSON(w, http.StatusNotFound, mappedError)
	default:
		pkg.SendJSON(w, http.StatusInternalServerError, mappedError)
	}
}

func (s *UserController) setToken(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
