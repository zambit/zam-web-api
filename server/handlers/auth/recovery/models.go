package recovery

// StartRequest
type StartRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
}

// VerifyRequest
type VerifyRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
	Code  string `json:"verification_code" validate:"required,min=6"`
}

// FinishRequest
type FinishRequest struct {
	Phone string `json:"phone" validate:"required,phone"`
	Token string `json:"recovery_token" validate:"required"`

	Password             string `validate:"required,min=6,alphanum" json:"password"`
	PasswordConfirmation string `validate:"required,eqfield=Password" json:"password_confirmation" `
}
