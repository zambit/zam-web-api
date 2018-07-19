package signup

// TokenView
type TokenView struct {
	Token string `json:"signup_token"`
}

// FinishResponse represent finish request response
type FinishResponse struct {
	Token string `json:"token"`
}
