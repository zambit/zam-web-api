package base

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"reflect"
	"strings"
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
	v := ginValidatorV9{validator: validator.New()}

	// init validator
	initValidator(v.validator)

	// bind
	binding.Validator = v
}

func initValidator(v *validator.Validate) {
	// init custom validators
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		val := fl.Field()
		vall := val.Interface()
		_ = vall
		var phone string
		switch val.Type().Kind() {
		case reflect.String:
			phone = val.String()
			if len(phone) == 0 {
				return true
			}
		case reflect.Ptr:
			if val.Elem().Kind() != reflect.String {
				return false
			}
			if val.Elem().IsNil() {
				return true
			}
			phone = val.Elem().String()
		}
		// validate phone
		_, err := types.NewPhone(phone)
		return err == nil
	})

	// init field func to obtain field json names rather then original names
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		// lookup json tag
		jsonTag, ok := field.Tag.Lookup("json")
		if ok && len(jsonTag) > 0 {
			// in case where tag defined as `json:"field_name,..."`
			if idx := strings.Index(jsonTag, ","); idx != -1 && len(jsonTag[:idx]) > 0 {
				return jsonTag[:idx]
			}
			// in case when tag defined as `json:"field_name"`
			return jsonTag
		}

		// either tag is empty or not defined at all, so it fullbacks on struct field name
		return field.Name
	})
}
