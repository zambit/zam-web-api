package base

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

// HandlerFunc specific project-wide handler function, must return nil or object which will be json-serialized,
// return code (0 mean 200 with resp provided or 204 otherwise) and error.
type HandlerFunc func(c *gin.Context) (resp interface{}, code int, err error)

// BaseResponse
type BaseResponse struct {
	Result bool        `json:"result"`
	Errors []error     `json:"errors,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// WrapHandler wraps our into gin form, dealing with returned values
func WrapHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// perform handler
		val, code, err := handler(c)
		if code == 0 {
			// fallback onto default
			// it's safe due to errors always overrides returned code
			if val == nil {
				code = 204
			} else {
				code = 200
			}
		}

		// collect errors
		errors := make([]error, 0, 1 + len(c.Errors))
		if err != nil {
			// it's expect that errors will come from validator or in form of errors views
			// other errors are interpreted as internal errors
			switch e := err.(type) {
			case validator.ValidationErrors:
				errors = append(errors, NewFieldsErrorsView(e))
				code = http.StatusBadRequest
			case ErrorView:
				errors = append(errors, e)
				code = e.Code
			case FieldsErrorView:
				errors = append(errors, e)
				code = e.Code
			default:
				// TODO HIJACK find the way te determine empty request body
				if e.Error() == "EOF" {
					errors = append(errors, ErrorView{Message: "Unexpected end of body"})
				} else {
					errors = append(errors, e)
				}

				code = http.StatusInternalServerError
			}
		}
		// append additional errors collected while request handling
		for _, e := range c.Errors {
			errors = append(errors, e)
			if e.Type == gin.ErrorTypePrivate {
				// promote status to 500 in case of private error
				code = http.StatusInternalServerError
			}
		}

		// write response object
		c.JSON(code, BaseResponse{
			Result: len(errors) == 0,
			Errors: errors,
			Data:   val,
		})
	}
}
