package models

type AuthRequest struct {
	Cred string `json:"cred,omitempty"`
}

type AuthResponse struct {
	Status string `json:"s,omitempty"`
	Auth   struct {
		UserId string `json:"user_id"`
	} `json:"auth"`
}

func NewAuthMessage(reqId int64, accessToken string) *Message[AuthRequest] {
	return &Message[AuthRequest]{
		Type: "d",
		Desc: MessageDescription[AuthRequest]{
			RequestId: reqId,
			Action:    "auth",
			Body: AuthRequest{
				Cred: accessToken,
			},
		},
	}
}

type LoginResponse struct {
	Kind         string `json:"kind"`
	LocalId      string `json:"localId"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	IdToken      string `json:"idToken"`
	Registered   bool   `json:"registered"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

type GetTokenResponse struct {
	UserId       string `json:"user_id"`
	ProjectId    string `json:"project_id"`
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	ExpiresIn    string `json:"expires_in"`
}
