package auth

// UserTokenResponse represents user sigin and signup responses
type UserTokenResponse struct {
	Token string `json:"token"`
}
