package auth

// UserTokenResponse represents user sigin and signup responses
type UserTokenResponse struct {
	Token string `json:"token"`
}

// UserPhoneResponse represents user auth check response
type UserPhoneResponse struct {
	Phone string `json:"phone"`
}
