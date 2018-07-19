package signup

// StartRequest
type StartRequest struct {
	Phone         string `json:"phone" validate:"required,phone"`
	ReferrerPhone string `json:"referrer_phone" validate:"phone"`
}

// VerifyRequest
type VerifyRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
	Code  string `json:"verification_code" validate:"required,min=6"`
}

// FinishRequest
type FinishRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
	Token string `json:"signup_token" validate:"required"`

	Password             string `validate:"required,min=6,alphanum" json:"password"`
	PasswordConfirmation string `validate:"required,eqfield=Password" json:"password_confirmation" `
}
