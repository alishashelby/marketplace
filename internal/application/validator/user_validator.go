package validator

import (
	"errors"
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
	if err := uv.validator.Struct(dto); err != nil {
		errs := make(map[string]any)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, valErr := range validationErrors {
				switch valErr.Tag() {
				case "min":
					errs[valErr.Field()] = fmt.Sprintf(
						ReportNeedMoreCharacters, valErr.Field(), valErr.Param())
				case "max":
					errs[valErr.Field()] = fmt.Sprintf(
						ReportTooManyCharacters, valErr.Field(), valErr.Param())
				case "alpha":
					errs[valErr.Field()] = fmt.Sprintf(ReportMustBeOnlyLetters, valErr.Field())
				case "containsany":
					errs[valErr.Field()] = fmt.Sprintf(
						ReportMustBeOneSpecialSymbol, valErr.Field(), valErr.Param())
				default:
					errs[valErr.Field()] = fmt.Sprintf(ReportFailedToValidate, valErr.Field())
				}
			}
		}

		return errs
	}

	return nil
}
