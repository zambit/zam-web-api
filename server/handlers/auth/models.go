package auth

// UserSignupRequest represents user signup request TODO add phones validators
type UserSignupRequest struct {
	Phone                string  `validate:"required,min=5" json:"phone"`
	Password             string  `validate:"required,min=5" json:"password"`
	PasswordConfirmation string  `validate:"required,eqfield=Password" json:"password_confirmation" `
	ReferrerPhone        *string `json:"referrer_phone"`
}

// UserSigninRequest represents user phone and password required to perform signin
type UserSigninRequest struct {
	Phone                string  `validate:"required,min=5" json:"phone"`
	Password             string  `validate:"required,min=5,eqfield=Password" json:"password"`
}