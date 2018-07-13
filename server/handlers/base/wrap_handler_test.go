package base

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/ginkgo/extensions/table"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"net/http"
	"github.com/pkg/errors"
)

func TestBaseHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Handlers Suite")
}

var _ = table.DescribeTable(
	"testing base handler wrapper",
	func(val interface{}, code int, err error, genCode int, generatedResp BaseResponse, additionalErrs ...error) {
		c := &gin.Context{}
		for _, e := range additionalErrs {
			c.Error(e)
		}

		gCode, gResp := postProcessResult(c, val, code, err)
		Expect(gCode).To(Equal(genCode))
		Expect(gResp).To(Equal(generatedResp))
	},
	table.Entry(
		"should return 200 when no error and content present", "CONTENT", 0, nil,
		http.StatusOK,
		BaseResponse{
			true,
			nil,
			"CONTENT",
		},
	),
	table.Entry(
		"should return 200 when no error and no content present", nil, 0, nil,
		http.StatusOK,
		BaseResponse{
			true,
			nil,
			nil,
		},
	),
	table.Entry(
		"should return internal error on unexpected error", nil, 0, errors.New("UNEXPECTED ERROR"),
		http.StatusInternalServerError,
		BaseResponse{
			false,
			[]error{
				ErrorView{
					Message: "UNEXPECTED ERROR",
				},
			},
			nil,
		},
	),
	table.Entry(
		"should give bad request on validation errors", nil, 0, validator.ValidationErrors{validator.FieldError(nil)},
		http.StatusBadRequest, BaseResponse{
			false,
			[]error{
				NewFieldsErrorsView(nil),
			},
			nil,
		},
	),
	table.Entry(
		"should give bad request on ErrorView error", nil, 0, ErrorView{Message: "err", Code: 430},
		430, BaseResponse{
			false,
			[]error{
				ErrorView{
					Message: "err", Code: 430,
				},
			},
			nil,
		},
	),
	table.Entry(
		"should give bad request on FieldsErrorView error", nil, 0, NewErrorsView("fields err"),
		http.StatusBadRequest, BaseResponse{
			false,
			[]error{
				NewErrorsView("fields err"),
			},
			nil,
		},
	),
)