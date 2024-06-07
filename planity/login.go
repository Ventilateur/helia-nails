package planity

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Ventilateur/helia-nails/planity/models"
	"github.com/Ventilateur/helia-nails/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type AuthInfo struct {
	ApiKey      string
	AccessToken string
	UserId      string
}

func (p *Planity) Login(email, password string) error {
	loginResp, err := utils.SendRequest[models.LoginResponse](
		p.httpClient,
		http.MethodPost,
		fmt.Sprintf("%s?key=%s", loginURL, p.authInfo.ApiKey),
		map[string]string{
			"accept":          "*/*",
			"accept-language": "en-GB,en;q=0.9",
			"content-type":    "application/json",
			"origin":          "https://pro.planity.com",
			"priority":        "u=1, i",
			"sec-fetch-dest":  "empty",
			"sec-fetch-mode":  "cors",
			"sec-fetch-site":  "cross-site",
		},
		strings.NewReader(fmt.Sprintf(
			`{"returnSecureToken":true,"email":"%s","password":"%s","clientType":"CLIENT_TYPE_WEB"}`,
			email, password,
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	tokenResp, err := utils.SendRequest[models.GetTokenResponse](
		p.httpClient,
		http.MethodPost,
		fmt.Sprintf("%s?key=%s", tokenURL, p.authInfo.ApiKey),
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		strings.NewReader(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", loginResp.RefreshToken)),
	)
	if err != nil {
		return fmt.Errorf("failed to get tokens: %w", err)
	}

	p.authInfo.AccessToken = tokenResp.AccessToken
	return nil
}

func (p *Planity) Connect(ctx context.Context) error {
	err := p.Login(p.config.Planity.Username, p.config.Planity.Password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	p.wsConn, _, err = websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	go p.logIncomingMessages()

	if err := wsjson.Write(ctx, p.wsConn, models.NewAuthMessage(p.nextReqId(), p.authInfo.AccessToken)); err != nil {
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	msg, err := waitForMessage[models.Message[models.AuthResponse]](
		p.messages,
		func(m *models.Message[models.AuthResponse]) bool {
			return m.Desc.RequestId == p.outMsgCount.Load() && m.Desc.Body.Status == "ok"
		},
	)
	if err != nil {
		return fmt.Errorf("failed to wait for auth response: %w", err)
	}

	p.authInfo.UserId = msg.Desc.Body.Auth.UserId

	return nil
}
