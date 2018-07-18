package signup

// TokenView
type TokenView struct {
	Token string `json:"token"`
}

// FinishResponse represent finish request response
type FinishResponse struct {
	Token string `json:"token"`
}
