package stext

import (
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications/stext/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSimpleTextNotificator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simple Text Notificator Test Suite")
}

const (
	testRecipient = "+79999999999"
)

var _ = Describe("testing simple text notificator", func() {
	Context("when sending notification", func() {
		Context("when sending completed event", func() {
			It("should does nothing when sending registration completed event", func() {
				notificator := &sender{backend: nil}

				err := notificator.Send(
					notifications.ActionRegistrationCompleted,
					map[string]interface{}{
						"phone": testRecipient,
					},
					notifications.Urgent,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should does nothing when sending recovery completed event", func() {
				notificator := &sender{backend: nil}

				err := notificator.Send(
					notifications.ActionPasswordRecoveryCompleted,
					map[string]interface{}{
						"phone": testRecipient,
					},
					notifications.Urgent,
				)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when sending registration confirmation code notification", func() {
			var notificator *sender

			BeforeEach(func() {
				backend := mocks.IBackend{}
				notificator = &sender{backend: &backend}
				backend.On("Send", testRecipient, "Your ZamZam verification code - 556611").Return(nil)
			})

			It("should do without errors", func() {
				err := notificator.Send(
					notifications.ActionRegistrationConfirmationRequested,
					map[string]interface{}{
						"code":  "556611",
						"phone": testRecipient,
					},
					notifications.Urgent,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error when data is wrong", func() {
				By("when data not a map[string]interface{}")
				err := notificator.Send(
					notifications.ActionRegistrationConfirmationRequested,
					100,
					notifications.Urgent,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`expecting map[string]interface{} as data, not int`))

				By("when phone is missing")
				err = notificator.Send(
					notifications.ActionRegistrationConfirmationRequested,
					map[string]interface{}{
						"code": "556611",
					},
					notifications.Urgent,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`"phone" to be passed using data argument`))

				By("when code is missing")
				err = notificator.Send(
					notifications.ActionRegistrationConfirmationRequested,
					map[string]interface{}{
						"phone": testRecipient,
					},
					notifications.Urgent,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`expecting both "code" and "phone" to be passed using data argument`))
			})
		})

		Context("when sending recovery confirmation code notification", func() {
			var notificator *sender

			BeforeEach(func() {
				backend := mocks.IBackend{}
				notificator = &sender{backend: &backend}
				backend.On("Send", testRecipient, "Your password recovery code - 556611").Return(nil)
			})

			It("should do without errors", func() {
				err := notificator.Send(
					notifications.ActionPasswordRecoveryConfirmationRequested,
					map[string]interface{}{
						"phone": testRecipient,
						"code":  "556611",
					},
					notifications.Urgent,
				)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
