package auth

// UserSignupRequest represents user signup request
type UserSignupRequest struct {
	Phone                string  `validate:"required,min=5" json:"phone"`
	Password             string  `validate:"required,min=5,eqfield=Password" json:"password"`
	PasswordConfirmation string  `validate:"required" json:"password_confirmation" `
	ReferrerPhone        *string `json:"referrer_phone"`
}
