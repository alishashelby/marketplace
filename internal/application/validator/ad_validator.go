package validator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/alishashelby/marketplace/internal/application/dto"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/go-playground/validator/v10"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	maxImageSize   = 5 * 1024 * 1024
	timeout        = 15 * time.Second
	ImageURLField  = "image_url"
	ContentTypeKey = "Content-Type"
)

const (
	ReportErrorFetchingImgFromURL     = "error fetching image from url: %v"
	ReportErrorUnavailableImg         = "error - image is not available: %d"
	ReportErrorUnsupportedContentType = "error - unsupported content type: %s"
	ReportErrorReadingImg             = "error reading image: %v"
	ReportErrorLargeImg               = "error - image too large: %d, but need %d"
	ReportInvalidFormatImg            = "error - invalid image format: %v"
	ReportNeedMoreCharacters          = "%s must be at least %s"
	ReportTooManyCharacters           = "%s must be at most %s"
	ReportNeedURL                     = "%s url is required"
	ReportNeedPositive                = "%s must be greater then 0"
	ReportFailedToValidate            = "failed to validate field %s"
	ReportErrorInComparePrices        = "min_price cannot be greater than max_price"
	ReportInvalidSortBy               = "invalid sort_by parameter"
)

type AdValidator struct {
	validator    *validator.Validate
	contentTypes map[string]struct{}
}

func NewAdValidator() *AdValidator {
	return &AdValidator{
		validator: validator.New(),
		contentTypes: map[string]struct{}{
			"image/jpeg": {},
			"image/png":  {},
			"image/gif":  {},
		},
	}
}

func (v *AdValidator) validateImgFormat(imgURL string, errors map[string]any) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(imgURL)
	if err != nil {
		errors[ImageURLField] = fmt.Sprintf(ReportErrorFetchingImgFromURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errors[ImageURLField] = fmt.Sprintf(ReportErrorUnavailableImg, resp.StatusCode)
		return
	}

	contentType := resp.Header.Get(ContentTypeKey)
	if _, ok := v.contentTypes[contentType]; !ok {
		errors[ImageURLField] = fmt.Sprintf(ReportErrorUnsupportedContentType, contentType)
		log.Printf("unsupported content type: %s", contentType)
		return
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		errors[ImageURLField] = fmt.Sprintf(ReportErrorReadingImg, err)
		return
	}

	log.Printf("image size: %d", n)
	if n > maxImageSize {
		errors[ImageURLField] = fmt.Sprintf(ReportErrorLargeImg, n, maxImageSize)
		return
	}

	_, _, err = image.DecodeConfig(bytes.NewReader(buf.Bytes()))
	if err != nil {
		errors[ImageURLField] = fmt.Sprintf(ReportInvalidFormatImg, err)
		return
	}
}

func (v *AdValidator) Validate(dto dto.AdDTO) map[string]any {
	errs := make(map[string]any)

	if val := v.validator.Struct(dto); val != nil {
		for _, err := range val.(validator.ValidationErrors) {
			switch err.Tag() {
			case "min":
				errs[err.Field()] = fmt.Sprintf(ReportNeedMoreCharacters, err.Field(), err.Param())
			case "max":
				errs[err.Field()] = fmt.Sprintf(ReportTooManyCharacters, err.Field(), err.Param())
			case "url":
				errs[err.Field()] = fmt.Sprintf(ReportNeedURL, err.Field())
			case "gt":
				errs[err.Field()] = fmt.Sprintf(ReportNeedPositive, err.Field())
			default:
				errs[err.Field()] = fmt.Sprintf(ReportFailedToValidate, err.Field())
			}
		}
	}

	v.validateImgFormat(dto.ImageURL, errs)
	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (v *AdValidator) ValidateOptions(ops *entity.Options) error {
	if ops.Page < 1 {
		return fmt.Errorf(ReportNeedPositive, entity.ParamPage)
	}

	if ops.Limit < 1 {
		return fmt.Errorf(ReportNeedPositive, entity.ParamLimit)
	}
	if ops.Limit > entity.LimitMaxValue {
		ops.Limit = entity.LimitMaxValue
	}

	if ops.MinPrice < 0 {
		return fmt.Errorf(ReportNeedPositive, entity.ParamMinPrice)
	}
	if ops.MaxPrice < 0 {
		return fmt.Errorf(ReportNeedPositive, entity.ParamMaxPrice)
	}
	if ops.MinPrice > ops.MaxPrice {
		return errors.New(ReportErrorInComparePrices)
	}

	if ops.SortBy != "" && ops.SortBy != entity.SortByCreatedAt && ops.SortBy != entity.SortByPrice {
		return errors.New(ReportInvalidSortBy)
	}

	if ops.OrderBy != 0 && ops.OrderBy != entity.OrderByAsc && ops.OrderBy != entity.OrderByDesc {
		return errors.New("invalid order_by parameter")
	}

	return nil
}
