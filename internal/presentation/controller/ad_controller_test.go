package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/alishashelby/marketplace/internal/application/dto"
	"github.com/alishashelby/marketplace/internal/application/service"
	"github.com/alishashelby/marketplace/internal/application/validator"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/alishashelby/marketplace/internal/infrastructure/repository/ad"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	titleConst    = "title"
	textConst     = "test text 20 symbols"
	imageUrlConst = "https://upload.wikimedia.org/wikipedia/commons/c/c7/Tabby_cat_with_blue_eyes-3336579.jpg"
	priceConst    = 1000.1
)

type adControllerTest struct {
	ctrl         *gomock.Controller
	adRepo       *service.MockAdRepository
	userRepo     *service.MockUserRepository
	adController *AdController
}

func setUpAdControllerTest(t *testing.T) *adControllerTest {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockUserRepo := service.NewMockUserRepository(ctrl)
	userService := service.NewUserService(mockUserRepo, nil)

	mockAdRepo := service.NewMockAdRepository(ctrl)
	adService := service.NewAdService(mockAdRepo)

	adValidator := validator.NewAdValidator()

	adController := NewAdController(adService, userService, adValidator)

	return &adControllerTest{
		ctrl:         ctrl,
		adRepo:       mockAdRepo,
		userRepo:     mockUserRepo,
		adController: adController,
	}
}

func TestAdController_CreateAd(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	user := &entity.User{
		ID:       uuid.New(),
		Username: usernameConst,
	}

	test.userRepo.EXPECT().
		GetByID(user.ID).
		Return(user, nil)

	test.adRepo.EXPECT().
		Save(gomock.Any()).
		Return(nil)

	testAd := &dto.AdDTO{
		Title:    titleConst,
		Text:     textConst,
		ImageURL: imageUrlConst,
		Price:    priceConst,
	}

	body, err := json.Marshal(testAd)
	if err != nil {
		t.Errorf("error marshalling ad: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/ads", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, user.ID))

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(test.adController.CreateAd)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp dto.AdResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, titleConst, resp.Title)
	assert.Equal(t, textConst, resp.Text)
	assert.Equal(t, imageUrlConst, resp.ImageURL)
	assert.Equal(t, priceConst, resp.Price)
	assert.Equal(t, user.Username, resp.Username)
}

func TestAdController_CreateAd_BadJSON(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/ads", bytes.NewReader([]byte(`{ id: 123 }`)))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.CreateAd)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid character 'i'")
}

func TestAdController_CreateAd_ValidationErrors(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/ads", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, uuid.New()))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.CreateAd)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "Title")
	assert.Contains(t, resp, "Text")
	assert.Contains(t, resp, "ImageURL")
	assert.Contains(t, resp, "Price")
}

func TestAdController_CreateAd_Unauthorized(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	testAd := &dto.AdDTO{
		Title:    titleConst,
		Text:     textConst,
		ImageURL: imageUrlConst,
		Price:    priceConst,
	}

	body, err := json.Marshal(testAd)
	if err != nil {
		t.Errorf("error marshalling ad: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/ads", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.CreateAd)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp string
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, unauthorizedError, resp)
}

func TestAdController_CreateAd_UserDoesNotExist(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	userID := uuid.New()

	test.userRepo.EXPECT().
		GetByID(userID).
		Return(nil, service.ErrorUserWithIDDoesNotExists)

	testAd := &dto.AdDTO{
		Title:    titleConst,
		Text:     textConst,
		ImageURL: imageUrlConst,
		Price:    priceConst,
	}

	body, err := json.Marshal(testAd)
	if err != nil {
		t.Errorf("error marshalling ad: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/ads", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, userID))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.CreateAd)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp string
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrorUserWithIDDoesNotExists.Error(), resp)
}

func TestAdController_CreateAd_AdServiceError(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	userID := uuid.New()
	user := &entity.User{ID: userID}

	test.userRepo.EXPECT().
		GetByID(userID).
		Return(user, nil)

	test.adRepo.EXPECT().
		Save(gomock.Any()).
		Return(ad.ErrorFailedToSaveAd)

	testAd := &dto.AdDTO{
		Title:    titleConst,
		Text:     textConst,
		ImageURL: imageUrlConst,
		Price:    priceConst,
	}

	body, err := json.Marshal(testAd)
	if err != nil {
		t.Errorf("error marshalling ad: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/ads", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, userID))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.CreateAd)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp string
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, ad.ErrorFailedToSaveAd.Error(), resp)
}

func TestAdController_GetAllAds(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	ad1 := entity.NewAd("title1", "text1", "image1", 100, &entity.User{ID: uuid.New()})
	ad2 := entity.NewAd("title2", "text2", "image2", 200, &entity.User{ID: uuid.New()})

	test.adRepo.EXPECT().
		FindAll(gomock.Any()).
		Return([]*entity.Ad{ad1, ad2}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/ads?page=1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.GetAllAds)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []dto.AdResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestAdController_GetAdsWithOwned(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	owner, other := "owner", "other"

	user := &entity.User{ID: uuid.New(), Username: owner}
	otherUser := &entity.User{ID: uuid.New(), Username: other}

	ad1 := entity.NewAd("title1", "text1", "image1", 100, user)
	ad2 := entity.NewAd("title2", "text2", "image2", 200, otherUser)

	test.adRepo.EXPECT().
		FindAll(gomock.Any()).
		Return([]*entity.Ad{ad1, ad2}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/ads/?page=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, user.ID))
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.GetAdsWithOwned)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []dto.AdResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)

	assert.True(t, resp[0].IsOwner)
	assert.False(t, resp[1].IsOwner)
	assert.Equal(t, owner, resp[0].Username)
	assert.Equal(t, other, resp[1].Username)
}

func TestAdController_GetAdsWithOwned_Unauthorized(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/ads/?page=1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.GetAdsWithOwned)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp string
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, unauthorizedError, resp)
}

func TestAdController_GetAllAds_WithOptions(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	user := &entity.User{ID: uuid.New()}
	ad1 := entity.NewAd("title1", "text1", "image1", 100, user)
	ad2 := entity.NewAd("title2", "text2", "image2", 200, user)

	test.adRepo.EXPECT().
		FindAll(gomock.Any()).
		Do(func(ops *entity.Options) {
			assert.Equal(t, 2, ops.Page)
			assert.Equal(t, 20, ops.Limit)
			assert.Equal(t, entity.SortByPrice, ops.SortBy)
			assert.Equal(t, entity.OrderByAsc, ops.OrderBy)
			assert.Equal(t, 50.0, ops.MinPrice)
			assert.Equal(t, 300.0, ops.MaxPrice)
		}).
		Return([]*entity.Ad{ad1, ad2}, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/ads?page=2&limit=20&sort_by=price&order_by=1&min_price=50&max_price=300",
		nil,
	)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.GetAllAds)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []dto.AdResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.False(t, resp[0].IsOwner)
}

func TestAdController_GetAds_InvalidOptions(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	testCases := []struct {
		name           string
		query          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid page",
			query:          "page=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
		{
			name:           "invalid limit",
			query:          "page=1&limit=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
		{
			name:           "invalid order_by",
			query:          "page=1&order_by=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
		{
			name:           "invalid min_price",
			query:          "page=1&min_price=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "strconv.ParseFloat: parsing \"invalid\": invalid syntax",
		},
		{
			name:           "invalid max_price",
			query:          "page=1&max_price=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "strconv.ParseFloat: parsing \"invalid\": invalid syntax",
		},
		{
			name:           "invalid sort_by",
			query:          "page=1&sort_by=invalid_field",
			expectedStatus: http.StatusBadRequest,
			expectedError:  validator.ReportInvalidSortBy,
		},
		{
			name:           "invalid order_by value",
			query:          "page=1&order_by=3",
			expectedStatus: http.StatusBadRequest,
			expectedError:  validator.ReportInvalidOrderBy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/ads?"+tc.query, nil)
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(test.adController.GetAllAds)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var resp string
			err := json.NewDecoder(w.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedError, resp)
		})
	}
}

func TestAdController_GetAds_NotFound(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	test.adRepo.EXPECT().
		FindAll(gomock.Any()).
		Return(nil, ad.ErrorAdsNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/ads?page=1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.GetAllAds)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp string
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, ad.ErrorAdsNotFound.Error(), resp)
}

func TestAdController_GetAds_ServiceError(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	bdErr := "database error"
	test.adRepo.EXPECT().
		FindAll(gomock.Any()).
		Return(nil, errors.New(bdErr))

	req := httptest.NewRequest(http.MethodGet, "/api/ads?page=1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(test.adController.GetAllAds)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp string
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, bdErr, resp)
}

func TestAdController_GetIDFromToken(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, userID))

	id, err := test.adController.getIDFromToken(req)

	assert.NoError(t, err)
	assert.Equal(t, userID, id)
}

func TestAdController_GetIDFromToken_NothingInContext(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id, err := test.adController.getIDFromToken(req)

	assert.Error(t, err)
	assert.Equal(t, unauthorizedError, err)
	assert.Equal(t, uuid.Nil, id)
}

func TestAdController_GetIDFromToken_InvalidType(t *testing.T) {
	test := setUpAdControllerTest(t)
	defer test.ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), service.UserIDKey, "invalid"))

	id, err := test.adController.getIDFromToken(req)

	assert.Error(t, err)
	assert.Equal(t, unauthorizedError, err)
	assert.Equal(t, uuid.Nil, id)
}
