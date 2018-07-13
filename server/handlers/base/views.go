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
	// for name, vErr := range validationErrs {
	//
	// }
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
