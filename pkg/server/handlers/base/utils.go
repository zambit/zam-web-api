package base

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"net/http"
)

// ShouldBindJson
func ShouldBindJSON(c *gin.Context, to interface{}) error {
	err := c.ShouldBindJSON(to)
	if err != nil {
		if vErr, ok := err.(validator.ValidationErrors); ok {
			return ViewFromValidationErrs(vErr)
		}

		return ErrorView{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
	}
	return nil
}
