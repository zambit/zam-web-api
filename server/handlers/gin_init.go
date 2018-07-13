package handlers

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
)

// ginValidatorV9 overrides default gin validator
type ginValidatorV9 struct {
	validator *validator.Validate
}

func (v ginValidatorV9) ValidateStruct(val interface{}) error {
	return v.validator.Struct(val)
}

func (v ginValidatorV9) Engine() interface{} {
	return v.validator
}

// init overrides gin validator
func init() {
	binding.Validator = ginValidatorV9{validator: validator.New()}
}
