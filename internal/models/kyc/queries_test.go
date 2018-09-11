package kyc_test

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/db"
	. "git.zam.io/wallet-backend/web-api/fixtures"
	"git.zam.io/wallet-backend/web-api/fixtures/database"
	"git.zam.io/wallet-backend/web-api/fixtures/database/migrations"
	"git.zam.io/wallet-backend/web-api/internal/models/kyc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestKycModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KYC Models Suite")
}

const (
	user1Phone = "+79000000001"
	user1FN    = "fn1"
	user1LN    = "ln1"
	user1S     = "male"
	user1C     = "RUSSIA MTF!"
	user2Phone = "+79000000002"
)

var (
	user1BD = time.Date(2010, 12, 1, 1, 1, 1, 1, time.Local)
	user1A  = map[string]interface{}{
		"city":   "some city",
		"street": "some street",
	}
)

var _ = Describe("user related queries", func() {
	Init()
	database.Init()
	migrations.Init()

	preInsertPhones := []interface{}{user1Phone, user2Phone}

	type (
		phoneToIDMapT  map[string]int64
		statusToIDMapT map[kyc.StatusType]int64
	)

	BeforeEachCProvide(func(d *db.Db) (res phoneToIDMapT) {
		// prepend table before test to be sure returned user id is actual
		rows, err := d.Query(
			fmt.Sprintf(
				`INSERT INTO users (phone) VALUES 
					($1), ($2)
				RETURNING id`,
			),
			preInsertPhones...,
		)
		Expect(err).NotTo(HaveOccurred())

		res = phoneToIDMapT{}
		defer rows.Close()
		for i := 0; rows.Next(); i++ {
			var id int64
			err := rows.Scan(&id)
			Expect(err).NotTo(HaveOccurred())
			res[preInsertPhones[i].(string)] = id
		}
		return res
	})

	BeforeEachCProvide(func(d *db.Db) statusToIDMapT {
		res := statusToIDMapT{}
		for _, s := range []kyc.StatusType{kyc.StatusPending, kyc.StatusVerified, kyc.StatusDeclined} {
			var id int64
			err := d.QueryRow(`select id from personal_data_statuses where name = $1`, s).Scan(&id)
			Expect(err).NotTo(HaveOccurred())
			res[s] = id
		}
		return res
	})

	Describe("when creating user kyc data", func() {
		ItD("should create kyc record with appropriate values", func(d *db.Db, users phoneToIDMapT, statuses statusToIDMapT) {
			id, err := kyc.Create(d, &kyc.Data{
				UserID:    users[user1Phone],
				Status:    kyc.StatusPending,
				FirstName: user1FN,
				LastName:  user1LN,
				BirthDate: user1BD,
				Sex:       user1S,
				Country:   user1C,
				Address:   user1A,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(id).NotTo(BeEquivalentTo(0))

			By("ensuring database state")
			insertedData := map[string]interface{}{}
			err = d.QueryRowx(`select * from personal_data where id = $1`, id).MapScan(insertedData)
			Expect(err).NotTo(HaveOccurred())

			Expect(insertedData["user_id"]).To(BeEquivalentTo(users[user1Phone]))
			Expect(insertedData["first_name"]).To(BeEquivalentTo(user1FN))
			Expect(insertedData["last_name"]).To(BeEquivalentTo(user1LN))
			Expect(insertedData["sex"]).To(BeEquivalentTo("male"))
		})
	})
})
