package signup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.zam.io/wallet-backend/web-api/db"
	. "git.zam.io/wallet-backend/web-api/fixtures"
	"git.zam.io/wallet-backend/web-api/fixtures/database"
	"git.zam.io/wallet-backend/web-api/fixtures/database/migrations"
	"git.zam.io/wallet-backend/web-api/internal/models"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	nosqlmock "git.zam.io/wallet-backend/web-api/pkg/services/nosql/mocks"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	notifmock "git.zam.io/wallet-backend/web-api/internal/services/notifications/mocks"
	iscmock "git.zam.io/wallet-backend/web-api/internal/services/isc/mocks"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	sessmock "git.zam.io/wallet-backend/web-api/pkg/services/sessions/mocks"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
	"time"
	"git.zam.io/wallet-backend/web-api/internal/services/isc"
)

const (
	invalidPhone = "+7(999)000-00-000"
	validPhone1  = "+79871111111"
	validPhone2  = "+79871111112"
	validPhone3  = "+79871111113"
	pass1        = "123451"
	pass2        = "543211"
	confirmCode  = "CONFIRMATIONCODE"
	confirmCode2 = "222222222CONFIRMATIONCODE2222222222"
	signUpToken  = "SIGNUPTOKENTOKENTOKEN"
	signUpToken2 = "2222SIGNUPTOKENTOKENTOKEN222"
	authToken    = "AUTH TOKEN"
)

func TestSignUpHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SignUp Handlers Suite")
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

func createSimpleContext(body interface{}) *gin.Context {
	return CreateContext("POST", "NOT DEFINED", body)
}

var _ = Describe("Given user signup flow", func() {
	Init()
	database.Init()
	migrations.Init()

	BeforeEachCProvide(func() (*nosqlmock.IStorage, nosql.IStorage) {
		s := &nosqlmock.IStorage{}
		return s, s
	})

	BeforeEachCProvide(func() (*iscmock.IEventNotificator, isc.IEventNotificator) {
		s := &iscmock.IEventNotificator{}
		return s, s
	})

	BeforeEachCProvide(func() (*notifmock.IGenerator, notifications.IGenerator) {
		g := &notifmock.IGenerator{}
		return g, g
	})

	BeforeEachCProvide(func() (*sessmock.IStorage, sessions.IStorage) {
		s := &sessmock.IStorage{}
		return s, s
	})

	Context("when querying auth/signup/start", func() {
		BeforeEachCProvide(
			func(
				d *db.Db,
				storage nosql.IStorage,
				notifier isc.IEventNotificator,
				generator notifications.IGenerator,
			) base.HandlerFunc {
				return StartHandlerFactory(d, notifier, generator, storage, time.Minute)
			},
		)
		BeforeEachCProvide(func(d *db.Db) models.User {
			referrer, err := models.NewUser(validPhone1, pass1, models.UserStatusActive, nil)
			Expect(err).NotTo(HaveOccurred())

			referrer, err = models.CreateUser(d, referrer)
			Expect(err).NotTo(HaveOccurred())

			return referrer
		})

		Context("when start performed without errors", func() {
			BeforeEachCInvoke(func(
				storage *nosqlmock.IStorage,
				generator *notifmock.IGenerator,
			) {
				// setup mocks
				generator.On("RandomCode").Return(confirmCode)
				storage.On("SetWithExpire", "user:"+validPhone2+":signup:code", confirmCode, mock.Anything).Return(nil)
				storage.On("Delete", "user:"+validPhone2+":signup:token").Return(nil)
			})

			for _, state := range []models.UserStatusName{models.UserStatusPending, models.UserStatusVerified} {
				Context(fmt.Sprintf("when user already in %s", state), func() {
					type providedUser models.User
					BeforeEachCProvide(func(d *db.Db, notifSender *iscmock.IEventNotificator) providedUser {
						user, err := models.NewUser(validPhone2, pass2, state, nil)
						Expect(err).NotTo(HaveOccurred())

						user, err = models.CreateUser(d, user)
						Expect(err).NotTo(HaveOccurred())

						notifSender.On(
							"RegistrationVerificationRequested",
							fmt.Sprint(user.ID),
							validPhone2,
							confirmCode,
						).Return(nil)

						return providedUser(user)
					})

					ItD(fmt.Sprintf("should translate to prending state from %s", state), func(
						d *db.Db,
						handler base.HandlerFunc,
						user providedUser,
						notifSender *iscmock.IEventNotificator,
					) {
						_, _, err := handler(createSimpleContext(gin.H{
							"phone": user.Phone,
						}))
						Expect(err).NotTo(HaveOccurred())

						u, err := models.GetUserByID(d, fmt.Sprintf("%d", user.ID))
						Expect(err).NotTo(HaveOccurred())
						Expect(u.Status).To(Equal(models.UserStatusPending))
					})
				})
			}

			ItD(
				"should return ok due to phone valid and no such user registered",
				func(handler base.HandlerFunc, d *db.Db, referrer models.User, notifSender *iscmock.IEventNotificator) {
					notifSender.On(
						"RegistrationVerificationRequested",
						mock.Anything,
						validPhone2,
						confirmCode,
					).Return(nil)

					val, _, err := handler(createSimpleContext(gin.H{
						"phone": validPhone2,
					}))
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(BeNil())

					By("verifying db state")
					user, err := models.GetUserByPhone(d, validPhone2)
					Expect(err).NotTo(HaveOccurred())

					By("verifying notifier mock calls")
					Expect(len(notifSender.Calls)).To(Equal(1))
					Expect(len(notifSender.Calls[0].Arguments)).To(Equal(3))
					Expect(notifSender.Calls[0].Arguments[0]).To(Equal(fmt.Sprint(user.ID)))
				},
			)

			ItD(
				"should return ok due to phone valid, no such user registered and referrer exists",
				func(handler base.HandlerFunc, d *db.Db, referrer models.User, notifSender *iscmock.IEventNotificator) {
					notifSender.On(
						"RegistrationVerificationRequested",
						mock.Anything,
						validPhone2,
						confirmCode,
					).Return(nil)

					val, _, err := handler(createSimpleContext(gin.H{
						"phone":          validPhone2,
						"referrer_phone": referrer.Phone,
					}))
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(BeNil())

					By("verifying db state")
					user, err := models.GetUserByPhone(d, validPhone2)
					Expect(err).NotTo(HaveOccurred())
					Expect(user.ReferrerID).To(Equal(&referrer.ID))

					By("verifying notifier mock calls")
					Expect(len(notifSender.Calls)).To(Equal(1))
					Expect(len(notifSender.Calls[0].Arguments)).To(Equal(3))
					Expect(notifSender.Calls[0].Arguments[0]).To(Equal(fmt.Sprint(user.ID)))
				},
			)
		})

		Context("when testing errors", func() {
			var d *db.Db
			var handler base.HandlerFunc
			var referrer models.User
			BeforeEachCInvoke(func(iDb *db.Db, iHandler base.HandlerFunc, r models.User) {
				d = iDb
				handler = iHandler
				referrer = r
			})

			table.DescribeTable(
				"should return error due to",
				func(body interface{}, expectErr error) {
					_, _, err := handler(createSimpleContext(body))
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(expectErr))
				},
				table.Entry(
					"such user already exists",
					gin.H{
						"phone": validPhone1,
					},
					base.NewErrorsView("wrong parameters").AddField(
						"body", "phone", "user already exists",
					),
				),
				table.Entry(
					"no such referrer",
					gin.H{
						"phone":          validPhone2,
						"referrer_phone": validPhone3,
					},
					base.NewErrorsView("wrong parameters").AddField(
						"body", "referrer_phone", "referrer not found",
					),
				),
				table.Entry(
					"phone format is invalid",
					gin.H{
						"phone": invalidPhone,
					},
					base.NewErrorsView("wrong parameters").AddField(
						"body", "phone", "phone is invalid",
					),
				),
				table.Entry(
					"referrer phone format is invalid",
					gin.H{
						"phone":          validPhone3,
						"referrer_phone": invalidPhone,
					},
					base.NewErrorsView("wrong parameters").AddField(
						"body", "referrer_phone", "phone is invalid",
					),
				),
				table.Entry(
					"phone format is invalid while referrer not found",
					gin.H{
						"phone":          invalidPhone,
						"referrer_phone": validPhone2,
					},
					base.NewErrorsView("wrong parameters").AddField(
						"body", "phone", "phone is invalid",
					).AddField(
						"body", "referrer_phone", "referrer not found",
					),
				),
				table.Entry(
					"user already exists while referrer phone format is invalid",
					gin.H{
						"phone":          validPhone1,
						"referrer_phone": invalidPhone,
					},
					base.NewErrorsView("wrong parameters").AddField(
						"body", "referrer_phone", "phone is invalid",
					).AddField(
						"body", "phone", "user already exists",
					),
				),
			)
		})
	})

	Context("when querying /auth/signup/verify", func() {
		BeforeEachCProvide(
			func(d *db.Db, storage nosql.IStorage, generator notifications.IGenerator) base.HandlerFunc {
				return VerifyHandlerFactory(d, generator, storage, time.Minute)
			},
		)
		BeforeEachCProvide(func(d *db.Db) models.User {
			referrer, err := models.NewUser(validPhone1, pass1, models.UserStatusPending, nil)
			Expect(err).NotTo(HaveOccurred())

			referrer, err = models.CreateUser(d, referrer)
			Expect(err).NotTo(HaveOccurred())

			return referrer
		})

		Context("when code is valid", func() {
			BeforeEachCInvoke(func(storage *nosqlmock.IStorage, generator *notifmock.IGenerator) {
				storage.On("Get", "user:"+validPhone1+":signup:code").Return(confirmCode, nil)
				storage.On("Delete", "user:"+validPhone1+":signup:code").Return(nil)
				storage.On("SetWithExpire", "user:"+validPhone1+":signup:token", signUpToken, mock.Anything).Return(nil)
				generator.On("RandomToken").Return(signUpToken)
			})

			ItD("should be ok since code is valid", func(d *db.Db, handler base.HandlerFunc, user models.User) {
				val, _, err := handler(createSimpleContext(map[string]interface{}{
					"phone":             validPhone1,
					"verification_code": confirmCode,
				}))
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(BeEquivalentTo(struct {
					Token string
				}{
					Token: signUpToken,
				}))

				u, err := models.GetUserByID(d, fmt.Sprintf("%d", user.ID))
				Expect(err).NotTo(HaveOccurred())
				Expect(u.Status).To(Equal(models.UserStatusVerified))
			})
		})

		Context("when code is wrong", func() {
			BeforeEachCInvoke(func(storage *nosqlmock.IStorage, generator *notifmock.IGenerator) {
				storage.On("Get", "user:"+validPhone1+":signup:code").Return(confirmCode2, nil)
			})

			ItD("should return verification code error", func(d *db.Db, handler base.HandlerFunc, user models.User) {
				_, _, err := handler(createSimpleContext(map[string]interface{}{
					"phone":             validPhone1,
					"verification_code": confirmCode,
				}))
				Expect(err).To(Equal(base.NewErrorsView("").AddField(
					"body", "verification_code", "code is wrong",
				)))
			})
		})
	})

	Context("when querying /auth/signup/finish", func() {
		BeforeEachCProvide(
			func(
				d *db.Db,
				storage nosql.IStorage,
				generator notifications.IGenerator,
				notifier isc.IEventNotificator,
				sessStorage sessions.IStorage,
			) base.HandlerFunc {
				return FinishHandlerFactory(d, storage, notifier, sessStorage, time.Minute)
			},
		)
		BeforeEachCProvide(func(d *db.Db) models.User {
			referrer, err := models.NewUser(validPhone1, pass1, models.UserStatusVerified, nil)
			Expect(err).NotTo(HaveOccurred())

			referrer, err = models.CreateUser(d, referrer)
			Expect(err).NotTo(HaveOccurred())

			return referrer
		})

		Context("when signup token is valid", func() {
			BeforeEachCInvoke(func(
				user models.User,
				storage *nosqlmock.IStorage,
				generator *notifmock.IGenerator,
				notifier *iscmock.IEventNotificator,
				sessStorage *sessmock.IStorage,
			) {
				storage.On("Get", "user:"+validPhone1+":signup:token").Return(signUpToken, nil)
				sessStorage.On(
					"New", map[string]interface{}{
						"id":    user.ID,
						"phone": user.Phone,
					}, time.Minute,
				).Return(sessions.Token(authToken), nil)
				storage.On("Delete", "user:"+validPhone1+":signup:token").Return(nil)
				notifier.On("RegistrationCompleted", fmt.Sprint(user.ID)).Return(nil)
			})

			ItD("should return ok because token is valid", func(d *db.Db, user models.User, handler base.HandlerFunc) {
				val, _, err := handler(createSimpleContext(gin.H{
					"phone":                 validPhone1,
					"signup_token":          signUpToken,
					"password":              pass1,
					"password_confirmation": pass1,
				}))
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(BeEquivalentTo(struct {
					Token string
				}{
					Token: authToken,
				}))

				u, err := models.GetUserByID(d, fmt.Sprintf("%d", user.ID))
				Expect(err).NotTo(HaveOccurred())
				Expect(u.Status).To(Equal(models.UserStatusActive))
			})
		})

		Context("when signup token is wrong", func() {
			BeforeEachCInvoke(func(
				user models.User,
				storage *nosqlmock.IStorage,
				generator *notifmock.IGenerator,
				notifier *iscmock.IEventNotificator,
				sessStorage *sessmock.IStorage,
			) {
				storage.On("Get", "user:"+validPhone1+":signup:token").Return(signUpToken, nil)
			})

			ItD("should return ok because token is valid", func(d *db.Db, user models.User, handler base.HandlerFunc) {
				val, _, err := handler(createSimpleContext(gin.H{
					"phone":                 validPhone1,
					"signup_token":          signUpToken2,
					"password":              pass1,
					"password_confirmation": pass1,
				}))
				Expect(err).To(Equal(base.NewErrorsView("").AddField(
					"body", "signup_token", "signup_token is wrong",
				)))
				Expect(val).To(BeNil())
			})
		})
	})
})
