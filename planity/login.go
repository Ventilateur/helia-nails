package planity

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/planity/models"
	"github.com/Ventilateur/helia-nails/utils"
	"github.com/golang-jwt/jwt/v5"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type AuthInfo struct {
	ApiKey      string
	AccessToken string
	UserId      string
}

func (p *Planity) Login() error {
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
			p.config.Planity.Username, p.config.Planity.Password,
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
	var (
		login = true
		err   error
	)
	if p.AccessToken() != "" {
		token, _, err := new(jwt.Parser).ParseUnverified(p.AccessToken(), jwt.MapClaims{})
		if err != nil {
			login = true
		} else {
			exp, err := token.Claims.GetExpirationTime()
			if err != nil {
				return fmt.Errorf("failed to get access token's expiration time: %w", err)
			}

			if time.Now().Add(10 * time.Minute).Before(exp.Time) {
				login = false
			}
		}
	}

	if login {
		slog.Info("Token expired, logging in...")
		err := p.Login()
		if err != nil {
			return fmt.Errorf("failed to login: %w", err)
		}
	} else {
		slog.Info("Reuse access token...")
	}

	p.wsConn, _, err = websocket.Dial(ctx, p.config.Planity.WebsocketUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	p.wsConn.SetReadLimit(-1)

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
