package auth

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/ZamzamTech/wallet-api/db"
	_ "gitlab.com/ZamzamTech/wallet-api/server/handlers"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	notifmocks "gitlab.com/ZamzamTech/wallet-api/services/notifications/mocks"
	sessmocks "gitlab.com/ZamzamTech/wallet-api/services/sessions/mocks"

	. "gitlab.com/ZamzamTech/wallet-api/fixtures"
	"gitlab.com/ZamzamTech/wallet-api/fixtures/database"
	"gitlab.com/ZamzamTech/wallet-api/fixtures/database/migrations"

	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/stretchr/testify/mock"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"net/http"
	"time"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
)

const (
	authExpire  = time.Second
	validPhone1 = "+79871111111"
	validPhone2 = "+79871111112"
	validPhone3 = "+79871111113"
	pass1       = "12345"
	pass2       = "54321"
	pass3       = "lkjhafnion2rmpu1-w0m9d12h3[f912u3nr0ym92p[,iod-0]\\/\\]"
	shortPass   = "123"
)

const tokenName = "TestBearer"
var mockedToken = sessions.Token("TOKENTOKENTOKEN")
var mockedToken2 = sessions.Token("ToKToKToKToKToKToKToKToK")

type tokenResp struct {
	Token string
}

func CreateContext(method, url string, body interface{}) *gin.Context {
	var bodyCont []byte
	var err error
	if body != nil {
		bodyCont, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyCont))
	if err != nil {
		panic(err)
	}

	return &gin.Context{Request: req}
}

func CreateContextWA(method, url string, body interface{}, tokenName, token string) *gin.Context {
	c := CreateContext(method, url, body)
	c.Request.Header.Add("Authorization", tokenName + " " + token)
	return c
}

func CreateSUContext(phone, pass, passConf string, referrer *string) *gin.Context {
	return CreateContext("POST", "signup", map[string]interface{}{
		"phone":                 phone,
		"password":              pass,
		"password_confirmation": passConf,
		"referrer_phone":        referrer,
	})
}

func CreateSIContext(phone, pass string) *gin.Context {
	return CreateContext("POST", "signin", map[string]interface{}{
		"phone":    phone,
		"password": pass,
	})
}

func CreateLOContext(token sessions.Token) *gin.Context {
	return CreateContextWA("DELETE", "signout", nil, tokenName, string(token))
}

var _ = Describe("Given the auth api", func() {
	Init()
	database.Init()
	migrations.Init()

	BeforeEachCProvide(func() (store *sessmocks.IStorage, sender *notifmocks.ISender) {
		return &sessmocks.IStorage{}, &notifmocks.ISender{}
	})
	BeforeEachCProvide(func(store *sessmocks.IStorage, sender *notifmocks.ISender) (sessions.IStorage, notifications.ISender) {
		return store, sender
	})

	Context("when querying signup", func() {
		BeforeEachCProvide(
			func(d *db.Db, sessStore sessions.IStorage, notificator notifications.ISender) base.HandlerFunc {
				return SignupHandlerFactory(d, sessStore, notificator, authExpire)
			},
		)

		Describe("with requests to iStorage and ISender", func() {
			BeforeEachCInvoke(func(sessStore *sessmocks.IStorage, sender *notifmocks.ISender) {
				// setup mocks
				sessStore.On("New", mock.Anything, authExpire).Return(mockedToken, nil)
				sender.On(
					"Send",
					notifications.ActionRegistrationCompleted,
					mock.Anything,
					notifications.Urgent,
				).Return(nil)
			})

			ItD("should be valid request", func(handler base.HandlerFunc, d *db.Db, sessStore *sessmocks.IStorage, sender *notifmocks.ISender) {
				// perform request
				data, code, err := handler(CreateSUContext(validPhone1, pass1, pass1, nil))
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(201))
				Expect(data).To(BeEquivalentTo(tokenResp{Token: string(mockedToken)}))

				// query just created user
				user, err := models.GetUserByPhone(d, validPhone1)
				Expect(err).NotTo(HaveOccurred())

				// validate calls
				Expect(len(sessStore.Calls)).To(Equal(1))
				Expect(sessStore.Calls[0].Method).To(Equal("New"))
				Expect(len(sessStore.Calls[0].Arguments)).To(Equal(2))

				sessArg := sessStore.Calls[0].Arguments[0].(map[string]interface{})
				Expect(sessArg).To(HaveKeyWithValue("id", user.ID))
			})

			Describe("with referrer", func() {
				type refererID int64

				BeforeEachCProvide(func(d *db.Db) refererID {
					referrer, err := models.NewUser(validPhone1, pass1, models.UserStatusActive, nil)
					Expect(err).NotTo(HaveOccurred())
					referrer, err = models.CreateUser(d, referrer)
					Expect(err).NotTo(HaveOccurred())

					return refererID(referrer.ID)
				})

				ItD("should be failed due to user already exists", func(handler base.HandlerFunc, _ refererID) {
					data, _, err := handler(CreateSUContext(validPhone1, pass1, pass1, nil))
					Expect(err).To(HaveOccurred())
					Expect(err).To(
						Equal(base.NewErrorsView("").AddField("body", "phone", models.ErrUserAlreadyExists.Error())),
					)
					Expect(data).To(BeNil())
				})

				ItD("should be failed due to no such referrer", func(handler base.HandlerFunc) {
					rph := validPhone3
					// perform request
					data, _, err := handler(CreateSUContext(validPhone2, pass1, pass1, &rph))
					Expect(err).To(
						Equal(base.NewErrorsView("").AddField(
							"body", "referrer_phone", models.ErrReferrerNotFound.Error()),
						),
					)
					Expect(data).To(BeNil())
				})

				ItD("should be ok with existed referrer", func(d *db.Db, handler base.HandlerFunc, refID refererID) {
					rph := validPhone1
					// perform request
					data, code, err := handler(CreateSUContext(validPhone2, pass1, pass1, &rph))
					Expect(err).NotTo(HaveOccurred())
					Expect(code).To(Equal(201))
					Expect(data).To(BeEquivalentTo(tokenResp{string(mockedToken)}))

					user, err := models.GetUserByPhone(d, validPhone2)
					Expect(err).NotTo(HaveOccurred())
					Expect(*user.ReferrerID).To(BeEquivalentTo(refID))
				})
			})
		})

		Describe("checking params", func() {
			ItD("failed due to missing all params", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateContext("POST", "signup", map[string]interface{}{}))
				Expect(data).To(BeNil())
				Expect(err).To(BeAssignableToTypeOf(validator.ValidationErrors{}))

				vErr := err.(validator.ValidationErrors)

				Expect(len(vErr)).To(Equal(3))
				for i, ns := range []string{
					"UserSignupRequest.Phone",
					"UserSignupRequest.Password",
					"UserSignupRequest.PasswordConfirmation",
				} {
					Expect(vErr[i].Tag()).To(Equal("required"))
					Expect(vErr[i].StructNamespace()).To(Equal(ns))
				}
			})

			ItD("failed due to password short", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateSUContext(validPhone1, shortPass, shortPass, nil))
				Expect(data).To(BeNil())
				Expect(err).To(BeAssignableToTypeOf(validator.ValidationErrors{}))

				vErr := err.(validator.ValidationErrors)
				Expect(len(vErr)).To(Equal(1))
				Expect(vErr[0].Tag()).To(Equal("min"))
				Expect(vErr[0].StructNamespace()).To(Equal("UserSignupRequest.Password"))
			})

			ItD("failed due to password confirmation not eq to password", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateSUContext(validPhone1, pass1, pass2, nil))
				Expect(data).To(BeNil())
				Expect(err).To(BeAssignableToTypeOf(validator.ValidationErrors{}))

				vErr := err.(validator.ValidationErrors)
				Expect(len(vErr)).To(Equal(1))
				Expect(vErr[0].Tag()).To(Equal("eqfield"))
				Expect(vErr[0].Param()).To(Equal("Password"))
				Expect(vErr[0].StructNamespace()).To(Equal("UserSignupRequest.PasswordConfirmation"))
			})
		})
	})

	Context("when querying singin requests", func() {
		BeforeEachCInvoke(func(d *db.Db) {
			user, err := models.NewUser(validPhone1, pass1, models.UserStatusActive, nil)
			Expect(err).NotTo(HaveOccurred())
			user, err = models.CreateUser(d, user)
			Expect(err).NotTo(HaveOccurred())
		})

		BeforeEachCProvide(
			func(d *db.Db, sessStore sessions.IStorage, notificator notifications.ISender) base.HandlerFunc {
				return SigninHandlerFactory(d, sessStore, authExpire)
			},
		)

		Context("IStorage occurs get", func() {
			BeforeEachCInvoke(func(sessStore *sessmocks.IStorage) {
				sessStore.On("New", mock.Anything, authExpire).Return(mockedToken, nil)
			})

			ItD("should return token", func(handler base.HandlerFunc, sessStore *sessmocks.IStorage) {
				data, _, err := handler(CreateSIContext(validPhone1, pass1))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(BeEquivalentTo(tokenResp{
					Token: string(mockedToken),
				}))

				// check user phone must be stored in session
				Expect(len(sessStore.Calls)).To(Equal(1))
				Expect(len(sessStore.Calls[0].Arguments)).To(Equal(2))

				sessPayload := sessStore.Calls[0].Arguments[0]
				Expect(sessPayload).To(HaveKeyWithValue("phone", types.Phone(validPhone1)))
			})
		})

		ItD("should fail due to wrong password", func(handler base.HandlerFunc) {
			data, _, err := handler(CreateSIContext(validPhone1, pass2))
			Expect(err).To(HaveOccurred())
			Expect(data).To(BeNil())
			Expect(err).To(Equal(base.NewErrorsView("wrong authorization data").AddField(
				"body", "phone", "either phone or password are invalid",
			)))
		})

		ItD("should fail due to wrong password v2", func(handler base.HandlerFunc) {
			data, _, err := handler(CreateSIContext(validPhone1, pass3))
			Expect(err).To(HaveOccurred())
			Expect(data).To(BeNil())
			Expect(err).To(Equal(base.NewErrorsView("wrong authorization data").AddField(
				"body", "phone", "either phone or password are invalid",
			)))
		})
	})

	Context("when querying signout request", func() {
		BeforeEachCProvide(
			func(sessStore sessions.IStorage) base.HandlerFunc {
				return SignoutHandlerFactory(sessStore, tokenName)
			},
		)

		Context("when token is stored", func() {
			BeforeEachCInvoke(func(sessStore *sessmocks.IStorage) {
				sessStore.On("Delete", sessions.Token(mockedToken)).Return(nil)
			})

			ItD("should logout", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateLOContext(mockedToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(BeNil())
			})
		})

		Context("when token expired", func() {
			BeforeEachCInvoke(func(sessStore *sessmocks.IStorage) {
				sessStore.On("Delete", sessions.Token(mockedToken)).Return(sessions.ErrExpired)
			})

			ItD("should not return error when token expired", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateLOContext(mockedToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(BeNil())
			})
		})
	})

	Context("when querying refresh token request", func() {
		BeforeEachCProvide(func(sessStore sessions.IStorage) base.HandlerFunc {
			return RefreshTokenHandlerFactory(sessStore, tokenName)
		})
		BeforeEachCInvoke(func(sessStore *sessmocks.IStorage) {
			sessStore.On("RefreshToken", sessions.Token(mockedToken)).Return(mockedToken2, nil)
		})

		Context("when token is stored", func() {
			ItD("should not return error when token expired", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateLOContext(mockedToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(BeEquivalentTo(tokenResp{
					Token: string(mockedToken2),
				}))
			})
		})
	})

	Context("when querying check request", func() {
		BeforeEachCProvide(func(sessStore sessions.IStorage) base.HandlerFunc {
			return CheckHandlerFactory()
		})

		ItD("should return phone attached to the session", func(handler base.HandlerFunc) {
			c := CreateContext("GET", "check", nil)
			c.Set("user_data", map[string]interface{}{
				"phone": types.Phone(validPhone1),
			})
			data, _, err := handler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(UserPhoneResponse{Phone: validPhone1}))
		})

		ItD("should return error because session data is missing", func(handler base.HandlerFunc) {
			_, _, err := handler(CreateContext("GET", "check", nil))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("auth passed but no user data attached"))
		})
	})
})