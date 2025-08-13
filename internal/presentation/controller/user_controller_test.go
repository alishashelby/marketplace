package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alishashelby/marketplace/internal/application/dto"
	"github.com/alishashelby/marketplace/internal/application/service"
	"github.com/alishashelby/marketplace/internal/application/validator"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	usernameConst = "username"
	passwordConst = "12345678!"
)

type userControllerTest struct {
	ctrl           *gomock.Controller
	userRepo       *service.MockUserRepository
	userController *UserController
}

func setUpUserControllerTest(t *testing.T) *userControllerTest {
	t.Helper()

	t.Setenv(service.DotEnvJWTExpiration, "21600")
	t.Setenv(service.DotEnvJWTSecret, "jwt-secret")

	ctrl := gomock.NewController(t)

	mockUserRepo := service.NewMockUserRepository(ctrl)
	JWTService, err := service.NewJWTService()
	if err != nil {
		t.Log("Failed to create JWT service")
	}
	userService := service.NewUserService(mockUserRepo, JWTService)
	userValidator := validator.NewUserValidator()

	userController := NewUserController(userService, userValidator)

	return &userControllerTest{
		ctrl:           ctrl,
		userRepo:       mockUserRepo,
		userController: userController,
	}
}

func TestUserController_Register(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	test.userRepo.EXPECT().
		GetByUsername(usernameConst).
		Return(nil, service.ErrorUserWithUsernameDoesNotExists)

	test.userRepo.EXPECT().
		Save(gomock.Any()).
		Do(func(user *entity.User) {
			if user.Username != usernameConst {
				t.Errorf("got username %v, want %v", user.Username, usernameConst)
			}
		}).
		Return(nil)

	userDTO := bytes.NewBufferString(`{"username":"` + usernameConst + `","password":"` + passwordConst + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", userDTO)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Register)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.UserDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	assert.Equal(t, dto.UserDTO{
		Username: usernameConst,
		Password: passwordConst,
	}, resp)
	authHeader := w.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)
	assert.Contains(t, authHeader, "Bearer ")
}

func TestUserController_Register_BadJSON(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(`{ id: 123 }`)))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Register)
	handler.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "invalid character 'i'")
}

func TestUserController_Register_ValidationErrors(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	username := "Username"
	password := "Password"

	testCases := []struct {
		name           string
		payload        string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "empty dto",
			payload:        `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{username, password},
		},
		{
			name:           "invalid username",
			payload:        `{"username":"user1", "password":"12345678&"}`,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{username},
		},
		{
			name:           "invalid password",
			payload:        `{"username":"username", "password":"1234567890"}`,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{password},
		},
		{
			name:           "multiple errors",
			payload:        `{"username":"0", "password":"1234"}`,
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{username, password},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.userRepo.EXPECT().
				GetByUsername(gomock.Any()).
				Times(0)

			test.userRepo.EXPECT().
				Save(gomock.Any()).
				Times(0)

			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(tc.payload))
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(test.userController.Register)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var resp map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Errorf("error decoding response body: %v", err)
			}

			for _, field := range tc.expectedFields {
				assert.Contains(t, resp, field)
				assert.NotEmpty(t, resp[field])
			}

			assert.Equal(t, len(tc.expectedFields), len(resp))
		})
	}
}

func TestUserController_UserService_UserExists(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	test.userRepo.EXPECT().
		GetByUsername(usernameConst).
		Return(&entity.User{}, nil)

	userDTO := bytes.NewBufferString(`{"username":"` + usernameConst + `","password":"` + passwordConst + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", userDTO)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Register)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUserController_Login(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	userID := uuid.New()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordConst), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("error hashing password: %v", err)
	}

	test.userRepo.EXPECT().
		GetByUsername(usernameConst).
		Return(&entity.User{
			ID:       userID,
			Username: usernameConst,
			Password: string(hashedPassword),
		}, nil)

	userDTO := bytes.NewBufferString(fmt.Sprintf(`{"username":"%s","password":"%s"}`,
		usernameConst, passwordConst))
	req := httptest.NewRequest(http.MethodPost, "/api/login", userDTO)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Login)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	assert.NotEmpty(t, resp["token"])
	assert.Equal(t, "Bearer "+resp["token"].(string), w.Header().Get("Authorization"))
}

func TestUserController_Login_BadJSON(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(`{ id: 123 }`)))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Login)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid character 'i'")
}

func TestUserController_Login_UserService_NotFoundUsername(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	test.userRepo.EXPECT().
		GetByUsername(usernameConst).
		Return(nil, service.ErrorUserWithUsernameDoesNotExists)

	userDTO := bytes.NewBufferString(
		fmt.Sprintf(`{"username":"%s","password":"something"}`, usernameConst))
	req := httptest.NewRequest(http.MethodPost, "/api/login", userDTO)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Login)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	assert.Equal(t, service.ErrorUserWithUsernameDoesNotExists.Error(), resp["error"].(string))
}

func TestUserController_Login_InternalServerError(t *testing.T) {
	test := setUpUserControllerTest(t)
	defer test.ctrl.Finish()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("wrongpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("error hashing password: %v", err)
	}

	expectedUser := &entity.User{
		Username: usernameConst,
		Password: string(hashedPassword),
	}

	test.userRepo.EXPECT().
		GetByUsername(usernameConst).
		Return(expectedUser, nil)

	userDTO := bytes.NewBufferString(
		fmt.Sprintf(`{"username":"%s","password":"something"}`, usernameConst))
	req := httptest.NewRequest(http.MethodPost, "/api/login", userDTO)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.userController.Login)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("error decoding response body: %v", err)
	}

	assert.Equal(t, service.ErrorInvalidPassword.Error(), resp["error"].(string))
}
