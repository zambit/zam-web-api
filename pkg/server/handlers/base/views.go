package base

import (
	"fmt"
	"git.zam.io/wallet-backend/common/pkg/merrors"
	"github.com/go-playground/validator"
	"strings"
)

// ErrorView
type ErrorView struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

// ErrorView
func (err ErrorView) Error() string {
	return fmt.Sprintf("%d: %s", err.Code, err.Message)
}

// FieldErrorView
type FieldErrorView struct {
	ErrorView `json:",inline"`

	Name  string `json:"name"`
	Input string `json:"input"`
}

// Error implements error interface
func (err FieldErrorView) Error() string {
	return fmt.Sprintf(`field error: %s{%s: %s}`, err.Input, err.Name, err.Message)
}

// NewFieldErr creates new field error
func NewFieldErr(input, name, message string) FieldErrorView {
	return FieldErrorView{
		ErrorView: ErrorView{Message: message},
		Input:     input,
		Name:      name,
	}
}

// HaveFieldErr checks is given error is list of errs, in such case scans whole list to search FieldErrorView with
// given field name.
func HaveFieldErr(err error, fieldName string) bool {
	switch errs := err.(type) {
	case merrors.Errors:
		for _, e := range errs {
			if fe, ok := e.(FieldErrorView); ok && fe.Name == fieldName {
				return true
			}
		}
	case FieldErrorView:
		return errs.Name == fieldName
	}
	return false
}

// NewFieldsErrorsView
func ViewFromValidationErrs(validationErrs validator.ValidationErrors) (view error) {
	for _, vErr := range validationErrs {
		if vErr == nil {
			continue
		}
		fieldName, message := coerceValidationErr(vErr)
		view = merrors.Append(
			view,
			NewFieldErr("body", fieldName, message),
		)
	}
	return
}

// coerceValidationErr coerce different types of validation to look like backend message
func coerceValidationErr(err validator.FieldError) (paramName, message string) {
	paramName = err.Field()

	switch err.Tag() {
	case "required":
		message = "field is required"
	case "min":
		message = fmt.Sprintf("field value must be at least %s items long", err.Param())
	case "eqfield":
		// TODO improve fields naming especially with cross-field validations
		message = fmt.Sprintf("this field must be equal to \"%s\"", strings.ToLower(err.Param()))
	case "phone":
		message = "phone is invalid"
	case "alpha":
		message = "only latin-alphabet letters allowed"
	case "alphanum":
		message = "only latin-alphabet letters or digits allowed"
	default:
		if e, ok := err.(error); ok {
			message = e.Error()
		} else {
			message = "unexpected error"
		}
	}
	return
}