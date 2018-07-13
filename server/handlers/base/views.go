package base

import (
	"fmt"
	"github.com/go-playground/validator"
	"strings"
)

// ErrorView
type ErrorView struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

// NewErrorView
func NewErrorView(err error, code ...int) ErrorView {
	c := 0
	if len(code) > 0 {
		c = code[0]
	}
	return ErrorView{
		Message: err.Error(),
		Code:    c,
	}
}

// ErrorView
func (err ErrorView) Error() string {
	return fmt.Sprintf("%d: %s", err.Code, err.Message)
}

// FieldErrorDescr
type FieldErrorDescr struct {
	Name    string `json:"name"`
	Input   string `json:"input"`
	Message string `json:"message"`
}

// FieldsErrorView represents errors occurred relative to the sets of fields
type FieldsErrorView struct {
	ErrorView `json:",inline"`
	Fields    []FieldErrorDescr `json:"fields"`
}

// NewFieldsErrorsView
func NewFieldsErrorsView(validationErrs validator.ValidationErrors) (view FieldsErrorView) {
	fieldErrs := make([]FieldErrorDescr, 0, len(validationErrs))
	for _, vErr := range validationErrs {
		if vErr == nil {
			continue
		}
		fieldName, message := coerceValidationErr(vErr)
		fieldErrs = append(
			fieldErrs,
			FieldErrorDescr{
				Input:   "body",
				Name:    fieldName,
				Message: message,
			},
		)
	}
	view.Fields = fieldErrs
	view.Message = "some fields contains bad formatted or invalid values"
	return
}

// coerceValidationErr coerce different types of validation to look like backend message
func coerceValidationErr(err validator.FieldError) (paramName, message string) {
	paramName = strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		message = "field is required"
	case "min":
		message = fmt.Sprintf("field value must be at least %s items long", err.Param())
	case "eqfield":
		// TODO improve fields naming especially with cross-field validations
		message = fmt.Sprintf("this field must be equal to %s", strings.ToLower(err.Param()))
	default:
		if e, ok := err.(error); ok {
			message = e.Error()
		} else {
			message = "unexpected error"
		}
	}
	return
}

// NewErrorsView
func NewErrorsView(message string) (view FieldsErrorView) {
	if message == "" {
		message = "some fields contains bad formatted or invalid values"
	}
	view.Message = message
	return
}

// AddField
func (err FieldsErrorView) AddField(input, name, message string) FieldsErrorView {
	err.Fields = append(err.Fields, FieldErrorDescr{name, input, message})
	return err
}

// ErrorView implements error interface
func (err FieldsErrorView) Error() string {
	builder := strings.Builder{}
	for _, f := range err.Fields {
		builder.WriteString(fmt.Sprintf("%s %s: %s\n", f.Input, f.Name, f.Message))
	}

	return fmt.Sprintf("%s: \n%s", err.ErrorView.Error(), builder.String())
}
