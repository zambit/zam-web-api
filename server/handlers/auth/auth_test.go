package auth

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/ZamzamTech/wallet-api/db"
	_ "gitlab.com/ZamzamTech/wallet-api/server/handlers"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	notifmocks "gitlab.com/ZamzamTech/wallet-api/services/notifications/mocks"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	sessmocks "gitlab.com/ZamzamTech/wallet-api/services/sessions/mocks"

	. "gitlab.com/ZamzamTech/wallet-api/fixtures"
	"gitlab.com/ZamzamTech/wallet-api/fixtures/database"
	"gitlab.com/ZamzamTech/wallet-api/fixtures/database/migrations"

	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"net/http"
	"time"
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
	c.Request.Header.Add("Authorization", tokenName+" "+token)
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

	Context("when querying singin requests", func() {
		BeforeEachCProvide(
			func(d *db.Db, sessStore sessions.IStorage, notificator notifications.ISender) base.HandlerFunc {
				return SigninHandlerFactory(d, sessStore, authExpire)
			},
		)

		Context("when user have active status", func() {
			BeforeEachCInvoke(func(d *db.Db) {
				user, err := models.NewUser(validPhone1, pass1, models.UserStatusActive, nil)
				Expect(err).NotTo(HaveOccurred())
				user, err = models.CreateUser(d, user)
				Expect(err).NotTo(HaveOccurred())
			})

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
					Expect(sessPayload).To(HaveKeyWithValue("phone", validPhone1))
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

		Context("when user doesn't have active status", func() {
			BeforeEachCInvoke(func(d *db.Db) {
				user, err := models.NewUser(validPhone1, pass1, models.UserStatusVerified, nil)
				Expect(err).NotTo(HaveOccurred())
				user, err = models.CreateUser(d, user)
				Expect(err).NotTo(HaveOccurred())
			})

			ItD("should fail due to user isn't active", func(handler base.HandlerFunc) {
				data, _, err := handler(CreateSIContext(validPhone1, pass1))
				Expect(err).To(HaveOccurred())
				Expect(data).To(BeNil())
				Expect(err).To(Equal(base.NewErrorsView("wrong authorization data").AddField(
					"body", "phone", "either phone or password are invalid",
				)))
			})
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
			return RefreshTokenHandlerFactory(sessStore, tokenName, authExpire)
		})
		BeforeEachCInvoke(func(sessStore *sessmocks.IStorage) {
			sessStore.On("RefreshToken", sessions.Token(mockedToken), authExpire).Return(mockedToken2, nil)
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
				"phone": validPhone1,
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
