package base

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"io"
	"net/http"
	"git.zam.io/wallet-backend/common/pkg/merrors"
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

// WrapMiddleware do same as WrapHandler but intended to wrap middlewares.
//
// If either err or http code are returned, request will be aborted
func WrapMiddleware(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// perform handler
		val, code, err := handler(c)

		// zero values means that middleware passes further
		if val == nil && code == 0 && err == nil {
			c.Next()
			return
		}

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

	sourceErrs, ok := err.(merrors.Errors)
	if !ok {
		sourceErrs = merrors.Errors{err}
	}

	// coerce returned error and try to guess response error code
	for _, e := range sourceErrs {
		if e == io.EOF {
			code = http.StatusBadRequest
			errors = append(errors, ErrorView{Message: "empty body"})
		} else if e != nil {
			// it's expect that errors will come from validator or in form of errors views
			// other errors are interpreted as internal errors
			switch e2 := e.(type) {
			case validator.ValidationErrors:
				if len(e2) == 0 {
					continue
				}
				errors = append(errors, ViewFromValidationErrs(e2))
				code = http.StatusBadRequest
			case ErrorView:
				errors = append(errors, e)
				if e2.Code == 0 {
					code = http.StatusBadRequest
				} else {
					code = e2.Code
				}
			case FieldErrorView:
				errors = append(errors, e)
				if e2.Code == 0 {
					code = http.StatusBadRequest
				} else {
					code = e2.Code
				}
			default:
				errors = append(errors, ErrorView{Message: e.Error()})

				code = http.StatusInternalServerError
			}
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
