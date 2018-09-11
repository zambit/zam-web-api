package kyc

import (
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/internal/models/kyc"
)

// CreateRequest
type CreateRequest struct {
	Email      string              `json:"email" validate:"required,email"`
	FirstName  string              `json:"first_name" validate:"required,alpha"`
	LastName   string              `json:"last_name" validate:"required,alpha"`
	BirthDate  *types.UnixTimeView `json:"birth_date" validate:"required"`
	Sex        string              `json:"sex" validate:"required,oneof=male female undefined"`
	Country    string              `json:"country" validate:"required,alpha"`
	City       string              `json:"city" validate:"required,alphawithspaces"`
	Region     string              `json:"region" validate:"required,alphawithspaces"`
	Street     string              `json:"street" validate:"required,alphawithspaces"`
	House      string              `json:"house" validate:"required,alphanum"`
	PostalCode int                 `json:"postal_code"  validate:"required"`
}

// View user personal data representation
type View struct {
	Status    string                 `json:"status"`
	Email     string                 `json:"email"`
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	BirthDate *types.UnixTimeView    `json:"birth_date"`
	Sex       string                 `json:"sex"`
	Country   string                 `json:"country"`
	Address   map[string]interface{} `json:"address"`
}

// GetResponse
type GetResponse struct {
	PersonalData *View `json:"personal_data"`
}

// ViewFromModel
func ViewFromModel(data *kyc.Data) *View {
	if data == nil {
		return nil
	}

	bd := types.UnixTimeView(data.BirthDate)
	return &View{
		Status:    string(data.Status),
		Email:     data.Email,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		BirthDate: &bd,
		Sex:       data.Sex,
		Country:   data.Country,
		Address:   data.Address,
	}
}
