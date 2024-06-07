package planity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Ventilateur/helia-nails/config"
	coremodels "github.com/Ventilateur/helia-nails/core/models"
	"nhooyr.io/websocket"
)

const (
	loginURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"
	tokenURL = "https://securetoken.googleapis.com/v1/token"
)

type Planity struct {
	httpClient  *http.Client
	wsConn      *websocket.Conn
	authInfo    AuthInfo
	outMsgCount *atomic.Int64
	config      *config.Config
	messages    *sync.Map
}

func New(ctx context.Context, httpClient *http.Client, config *config.Config) (*Planity, error) {
	planity := &Planity{
		httpClient: httpClient,
		authInfo: AuthInfo{
			ApiKey: config.Planity.ApiKey,
		},
		outMsgCount: &atomic.Int64{},
		config:      config,
		messages:    &sync.Map{},
	}

	if err := planity.Connect(ctx); err != nil {
		return nil, err
	}

	return planity, nil
}

func (p *Planity) Name() coremodels.Source {
	return coremodels.SourcePlanity
}

func (p *Planity) keepAlive(ctx context.Context) error {
	return p.wsConn.Write(ctx, websocket.MessageText, []byte("0"))
}

func (p *Planity) logIncomingMessages() {
	for {
		_, b, err := p.wsConn.Read(context.Background())
		if err != nil {
			slog.Error("failed to read incoming message", "error", err)
			continue
		}
		p.messages.Store(string(b), struct{}{})
	}
}

func waitForMessage[T any](buffer *sync.Map, filter func(m *T) bool) (*T, error) {
	var ret *T
	for timeout := time.After(10 * time.Second); ; {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for message")
		default:
			buffer.Range(func(msg, _ any) bool {
				m := new(T)
				if err := json.Unmarshal([]byte(msg.(string)), m); err != nil {
					return true
				}

				if filter(m) {
					buffer.Delete(msg)
					ret = m
					return false
				}

				return true
			})
			if ret != nil {
				return ret, nil
			}
		}
	}
}

func (p *Planity) nextReqId() int64 {
	return p.outMsgCount.Add(1)
}
