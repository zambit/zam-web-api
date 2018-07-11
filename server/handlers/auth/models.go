package auth

// UserSignupRequest represents user signup request
type UserSignupRequest struct {
	Phone                string  `json:"phone" validate:"required"`
	Password             string  `json:"password" validate:"required,eqfield=Password"`
	PasswordConfirmation string  `json:"password_confirmation" validate:"required"`
	ReferrerPhone        *string `json:"referrer_phone"`
}
