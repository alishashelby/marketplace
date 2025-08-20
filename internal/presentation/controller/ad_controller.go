package controller

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/alishashelby/marketplace/internal/application/dto"
	"github.com/alishashelby/marketplace/internal/application/service"
	"github.com/alishashelby/marketplace/internal/application/validator"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/alishashelby/marketplace/internal/infrastructure/repository/ad"
	"github.com/alishashelby/marketplace/pkg"
	"github.com/google/uuid"
)

const (
	unauthorizedError = "invalid or missing user ID"
)

type AdController struct {
	adService   *service.AdService
	userService *service.UserService
	validator   *validator.AdValidator
}

func NewAdController(adService *service.AdService,
	userService *service.UserService, validator *validator.AdValidator) *AdController {
	return &AdController{
		adService:   adService,
		userService: userService,
		validator:   validator,
	}
}

// CreateAd godoc
//
//	@Summary		Create a new advertisement
//	@Description	Publishes a new ad for the authenticated user
//	@Tags			Ads
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			ad	body		dto.AdDTO	true	"Ad data"
//	@Success		201	{object}	dto.AdResponse
//	@Failure		400	{object}	pkg.ValidationErrorResponse	"Validation error"
//	@Failure		401	{object}	pkg.ErrorResponse			"Unauthorized"
//	@Failure		500	{object}	pkg.ErrorResponse			"Internal server error"
//	@Router			/api/publish [post]
func (ac *AdController) CreateAd(w http.ResponseWriter, r *http.Request) {
	log.Print("AdController.CreateAd called")

	var adDTO dto.AdDTO
	if err := json.NewDecoder(r.Body).Decode(&adDTO); err != nil {
		pkg.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	errs := ac.validator.Validate(adDTO)
	if errs != nil {
		pkg.SendValidationError(w, http.StatusBadRequest, errs)
		return
	}

	userID, err := ac.getIDFromToken(r)
	if err != nil {
		pkg.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := ac.userService.GetByID(userID)
	if err != nil {
		pkg.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	newAdd := entity.NewAd(
		adDTO.Title,
		adDTO.Text,
		adDTO.ImageURL,
		adDTO.Price,
		user,
	)

	if err = ac.adService.Create(newAdd); err != nil {
		pkg.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	adResp := dto.NewAdResponse(newAdd)
	pkg.SendJSON(w, http.StatusCreated, adResp)
}

// GetAdsWithOwned godoc
//
//	@Summary		Get ads with ownership info
//	@Description	Returns a list of advertisements with ownership flag for the authenticated user
//	@Tags			Ads
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int		false	"Page number"						default(1)
//	@Param			limit		query		int		false	"Items per page"					default(10)	minimum(1)	maximum(40)
//	@Param			sortBy		query		string	false	"Sort field (created_at, price)"	default(created_at)
//	@Param			orderBy		query		int		false	"Order (1 asc, -1 desc)"			default(-1)
//	@Param			minPrice	query		number	false	"Minimum price"
//	@Param			maxPrice	query		number	false	"Maximum price"
//	@Success		200			{array}		dto.AdResponse
//	@Failure		400			{object}	pkg.ErrorResponse	"Invalid query params"
//	@Failure		401			{object}	pkg.ErrorResponse	"Unauthorized"
//	@Failure		404			{object}	pkg.ErrorResponse	"No ads found"
//	@Failure		500			{object}	pkg.ErrorResponse	"Internal server error"
//	@Router			/api/ads/ [get]
func (ac *AdController) GetAdsWithOwned(w http.ResponseWriter, r *http.Request) {
	log.Print("AdController.GetAdsWithOwned called")

	userID, err := ac.getIDFromToken(r)
	if err != nil {
		pkg.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	ac.getAds(w, r, userID)
}

// GetAllAds godoc
//
//	@Summary		Get all ads
//	@Description	Returns a list of all published ads
//	@Tags			Ads
//	@Produce		json
//	@Param			page		query		int		false	"Page number"						default(1)
//	@Param			limit		query		int		false	"Items per page"					default(10)	minimum(1)	maximum(40)
//	@Param			sortBy		query		string	false	"Sort field (created_at, price)"	default(created_at)
//	@Param			orderBy		query		int		false	"Order (1 asc, -1 desc)"			default(-1)
//	@Param			minPrice	query		number	false	"Minimum price"
//	@Param			maxPrice	query		number	false	"Maximum price"
//	@Success		200			{array}		dto.AdResponse
//	@Failure		400			{object}	pkg.ErrorResponse	"Invalid query params"
//	@Failure		404			{object}	pkg.ErrorResponse	"No ads found"
//	@Failure		500			{object}	pkg.ErrorResponse	"Internal server error"
//	@Router			/api/ads [get]
func (ac *AdController) GetAllAds(w http.ResponseWriter, r *http.Request) {
	log.Print("AdController.GetAllAds called")

	ac.getAds(w, r, uuid.Nil)
}

func (ac *AdController) getIDFromToken(r *http.Request) (uuid.UUID, error) {
	userID, ok := r.Context().Value(service.UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New(unauthorizedError)
	}

	return userID, nil
}

func (ac *AdController) parseOptions(r *http.Request) (*entity.Options, error) {
	query := r.URL.Query()
	ops := &entity.Options{
		Page:     0,
		Limit:    entity.LimitDefaultValue,
		SortBy:   entity.SortByCreatedAt,
		OrderBy:  entity.OrderByDesc,
		MinPrice: 0,
		MaxPrice: 0,
	}

	if pageStr := query.Get(entity.ParamPage); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, err
		}
		ops.Page = page
	}

	if limitStr := query.Get(entity.ParamLimit); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, err
		}
		ops.Limit = limit
	}

	if sortBy := query.Get(entity.ParamSortBy); sortBy != "" {
		ops.SortBy = sortBy
	}

	if orderByStr := query.Get(entity.ParamOrderBy); orderByStr != "" {
		orderBy, err := strconv.Atoi(orderByStr)
		if err != nil {
			return nil, err
		}
		ops.OrderBy = orderBy
	}

	if minPriceStr := query.Get(entity.ParamMinPrice); minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			return nil, err
		}
		ops.MinPrice = minPrice
	}

	if maxPriceStr := query.Get(entity.ParamMaxPrice); maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			return nil, err
		}
		ops.MaxPrice = maxPrice
	}

	err := ac.validator.ValidateOptions(ops)
	if err != nil {
		return nil, err
	}

	return ops, nil
}

func (ac *AdController) getAds(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	ops, err := ac.parseOptions(r)
	if err != nil {
		pkg.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	ads, err := ac.adService.GetAds(ops)
	if err != nil {
		if errors.Is(err, ad.ErrorAdsNotFound) {
			pkg.SendError(w, http.StatusNotFound, err.Error())
			return
		}

		pkg.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	adsResp := make([]*dto.AdResponse, 0, len(ads))
	for _, a := range ads {
		resp := dto.NewAdResponse(a)
		if userID != uuid.Nil {
			resp.ProcessOwner(a, userID)
		}
		adsResp = append(adsResp, resp)
	}

	pkg.SendJSON(w, http.StatusOK, adsResp)
}
