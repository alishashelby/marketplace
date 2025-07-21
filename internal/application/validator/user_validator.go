package validator

import (
	"fmt"
	"github.com/alishashelby/marketplace/internal/application/dto"
	"github.com/go-playground/validator/v10"
)

const (
	ReportMustBeOnlyLetters      = "%s must contain only letters"
	ReportMustBeOneSpecialSymbol = "%s must contain at least one special character from %s"
)

type UserValidator struct {
	validator *validator.Validate
}

func NewUserValidator() *UserValidator {
	return &UserValidator{validator: validator.New()}
}

func (uv *UserValidator) Validate(dto dto.UserDTO) map[string]any {
	if val := uv.validator.Struct(dto); val != nil {
		errs := make(map[string]any)
		for _, err := range val.(validator.ValidationErrors) {
			switch err.Tag() {
			case "min":
				errs[err.Field()] = fmt.Sprintf(ReportNeedMoreCharacters, err.Field(), err.Param())
			case "max":
				errs[err.Field()] = fmt.Sprintf(ReportTooManyCharacters, err.Field(), err.Param())
			case "alpha":
				errs[err.Field()] = fmt.Sprintf(ReportMustBeOnlyLetters, err.Field())
			case "containsany":
				errs[err.Field()] = fmt.Sprintf(ReportMustBeOneSpecialSymbol, err.Field(), err.Param())
			default:
				errs[err.Field()] = fmt.Sprintf(ReportFailedToValidate, err.Field())
			}
		}

		return errs
	}

	return nil
}
