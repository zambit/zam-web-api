package base

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"testing"
)

func TestBaseHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Handlers Suite")
}

var _ = Describe("testing validator.ValidationErrors coercion into FieldsErrorView", func() {
	type exampleParam struct {
		Param1      string `validate:"required" json:"param1"`
		Param2      string `validate:"min=5" json:"param2"`
		Param3      string `json:"param3"`
		Param4      string `validate:"eqfield=Param3" json:"param4"`
		ComplexName string `validate:"required" json:"complexName"`
	}

	v := validator.New()
	initValidator(v)
	Context("when all params are invalid", func() {
		It("should coerce appropriate", func() {
			err := v.Struct(&exampleParam{
				Param1:      "",
				Param2:      "1234",
				Param3:      "example",
				Param4:      "miss_example",
				ComplexName: "",
			})
			Expect(err).To(HaveOccurred())

			vErr := err.(validator.ValidationErrors)
			Expect(ViewFromValidationErrs(vErr)).To(BeEquivalentTo(
				[]error{
					NewFieldErr("body", "param1", "field is required"),
					NewFieldErr("body", "param2", "field value must be at least 5 items long"),
					NewFieldErr("body", "param4", `this field must be equal to "param3"`),
					NewFieldErr("body", "complexName", "field is required"),
				},
			))
		})
	})
})

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
		"should return special error when body empty", nil, 0, io.EOF,
		http.StatusBadRequest,
		BaseResponse{
			false,
			[]error{
				ErrorView{
					Message: "empty body",
				},
			},
			nil,
		},
	),
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
			[]error{nil},
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
)
