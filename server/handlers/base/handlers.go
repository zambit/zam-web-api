package base

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"net/http"
	"io"
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

		// post-process response
		code, response := postProcessResult(c, val, code, err)

		// write response object
		c.JSON(code, response)
	}
}

// postProcessResult coerces handler result into base api response
func postProcessResult(c *gin.Context, val interface{}, code int, err error) (int, BaseResponse) {
	// collect errors
	var errors []error

	// coerce returned error and try to guess response error code
	if err == io.EOF {
		code = http.StatusBadRequest
		errors = append(errors, ErrorView{Message: "empty body"})
	} else if err != nil {
		// it's expect that errors will come from validator or in form of errors views
		// other errors are interpreted as internal errors
		switch e := err.(type) {
		case validator.ValidationErrors:
			errors = append(errors, NewFieldsErrorsView(e))
			code = http.StatusBadRequest
		case ErrorView:
			errors = append(errors, e)
			if e.Code == 0 {
				code = http.StatusBadRequest
			} else {
				code = e.Code
			}
		case FieldsErrorView:
			errors = append(errors, e)
			if e.Code == 0 {
				code = http.StatusBadRequest
			} else {
				code = e.Code
			}
		default:
			errors = append(errors, ErrorView{Message: e.Error()})

			code = http.StatusInternalServerError
		}
	}
	// append additional errors collected while request handling
	for _, e := range c.Errors {
		if e.Err == err {
			continue
		}

		errors = append(errors, e)
		if e.Type == gin.ErrorTypePrivate {
			// promote status to 500 in case of private error
			code = http.StatusInternalServerError
		}
	}

	// fallback onto default it nothing else determined
	if code == 0 {
		code = 200
	}

	return code, BaseResponse{
		Result: len(errors) == 0,
		Errors: errors,
		Data:   val,
	}
}