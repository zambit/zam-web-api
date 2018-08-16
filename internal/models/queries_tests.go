package models

import (
	"fmt"
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/db"
	. "git.zam.io/wallet-backend/web-api/fixtures"
	"git.zam.io/wallet-backend/web-api/fixtures/database"
	"git.zam.io/wallet-backend/web-api/fixtures/database/migrations"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	validPhone1  = "+79000000001"
	validPhone2  = "+79000000002"
	validPhone3  = "+79000000003"
	validPhone4  = "+79000000004"
	validPhone5  = "+79000000005"
	validPhone6  = "+79000000006"
	invalidPhone = "+790000000000"
	pass         = "1234"
)

var _ = Describe("user related queries", func() {
	Init()
	database.Init()
	migrations.Init()

	phoneToIDMap := make(map[string]int64)
	preInsertPhones := []interface{}{validPhone2, validPhone3, validPhone4, validPhone6}
	BeforeEachCInvoke(func(d *db.Db) {
		activeStatusID, err := getUserStatusID(d, UserStatusActive)
		pendingStatusID, err := getUserStatusID(d, UserStatusPending)
		Expect(err).NotTo(HaveOccurred())

		// prepend table before test to be sure returned user id is actual
		rows, err := d.Query(
			fmt.Sprintf(
				`INSERT INTO users (phone, status_id, registered_at, created_at) VALUES 
					($1, %d, now(), now()), ($2, %d, now(), now()), ($3, %d, now(), now()), ($4, %d, now(), now()) 
				RETURNING id`,
				activeStatusID, activeStatusID, activeStatusID, pendingStatusID,
			),
			preInsertPhones...,
		)
		Expect(err).NotTo(HaveOccurred())

		defer rows.Close()
		for i := 0; rows.Next(); i++ {
			var id int64
			err := rows.Scan(&id)
			Expect(err).NotTo(HaveOccurred())
			phoneToIDMap[preInsertPhones[i].(string)] = id
		}
	})

	Describe("when updating user", func() {
		ItD("should update status field", func(d *db.Db) {
			By("selecting previously created user")
			user, err := GetUserByID(d, fmt.Sprint(phoneToIDMap[validPhone2]))
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Status).To(Equal(UserStatusActive))
			prevCreatedAt := user.CreatedAt
			Expect(prevCreatedAt.IsZero()).NotTo(Equal(true))

			By("updating user status")
			user.Status = UserStatusPending
			err = UpdateUser(d, user)
			Expect(err).NotTo(HaveOccurred())

			By("ensuring user row")
			user, err = GetUserByID(d, fmt.Sprint(phoneToIDMap[validPhone2]))
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Status).To(Equal(UserStatusPending))

			By("should not change created at field")
			Expect(user.CreatedAt).To(Equal(prevCreatedAt))
		})

	})

	Describe("when querying new user", func() {
		ItD("should return valid user", func(d *db.Db) {
			user, err := GetUserByPhone(d, validPhone2)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ID).To(Equal(phoneToIDMap[validPhone2]))
			Expect(user.Phone).To(Equal(types.Phone(validPhone2)))
		})

		ItD("should return valid user and lock for update", func(d *db.Db) {
			user, err := GetUserByPhone(d, validPhone2, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ID).To(Equal(phoneToIDMap[validPhone2]))
			Expect(user.Phone).To(Equal(types.Phone(validPhone2)))
		})
	})

	Context("when creating new user", func() {
		Describe("without referrer", func() {
			ItD("should be ok", func(d *db.Db) {
				user, err := NewUser(validPhone1, pass, UserStatusActive, nil)
				Expect(err).NotTo(HaveOccurred())

				user, err = CreateUser(d, user)
				Expect(err).NotTo(HaveOccurred())

				var storedPhone string
				err = d.QueryRow(`SELECT phone FROM users WHERE id = $1`, user.ID).Scan(&storedPhone)
				Expect(err).NotTo(HaveOccurred())
				Expect(storedPhone).To(Equal(validPhone1))
			})
			ItD("should be failed due to user already exists", func(d *db.Db) {
				user, err := NewUser(validPhone2, pass, UserStatusActive, nil)
				Expect(err).NotTo(HaveOccurred())

				user, err = CreateUser(d, user)
				Expect(err).To(Equal(ErrUserAlreadyExists))
			})
		})

		Describe("with referrer", func() {
			var database *db.Db
			BeforeEachCInvoke(func(d *db.Db) {
				database = d
			})

			table.DescribeTable(
				"with referrer",
				func(refPhone string, expectErr error) {
					user, err := NewUser(validPhone1, pass, UserStatusActive, &refPhone)
					Expect(err).NotTo(HaveOccurred())

					_, err = CreateUser(database, user)

					if expectErr != nil {
						Expect(err).To(Equal(expectErr))
					} else {
						Expect(err).NotTo(HaveOccurred())
					}
				},
				[]table.TableEntry{
					{
						Description: "should be failed because no such referrer",
						Parameters: []interface{}{
							validPhone5, ErrReferrerNotFound,
						},
					},
					{
						Description: "should be failed because referrer not in active state",
						Parameters: []interface{}{
							validPhone6, ErrReferrerNotFound,
						},
					},
					{
						Description: "should be failed because invalid referrer phone",
						Parameters: []interface{}{
							invalidPhone, ErrReferrerNotFound,
						},
					},
					{
						Description: "should be ok",
						Parameters: []interface{}{
							validPhone2, nil,
						},
					},
				}...,
			)
		})
	})
})
