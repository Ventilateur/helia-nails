package planity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
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
			ApiKey:      config.Planity.ApiKey,
			AccessToken: config.Planity.AccessToken,
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

func (p *Planity) AccessToken() string {
	return p.authInfo.AccessToken
}

func (p *Planity) keepAlive() {
	for {
		time.Sleep(5 * time.Second)
		if err := p.wsConn.Write(context.Background(), websocket.MessageText, []byte("0")); err != nil {
			slog.Error("failed to send keep alive message")
		}
	}
}

func (p *Planity) logIncomingMessages() {
	for {
		_, b, err := p.wsConn.Read(context.Background())
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("connection closed, reconnect...")
				break
			} else {
				slog.Error("failed to read incoming message", "error", err)
				panic(err)
			}
		}

		// Multiple messages
		var msg string
		if nb, err := strconv.Atoi(string(b)); err == nil {
			var buff []byte
			for i := 0; i < nb; i++ {
				_, b, err := p.wsConn.Read(context.Background())
				if err != nil {
					slog.Error("failed to read incoming message", "error", err)
					panic(err)
				}
				buff = append(buff, b...)
			}
			msg = string(buff)
		} else {
			msg = string(b)
		}

		p.messages.Store(msg, struct{}{})
	}

	if err := p.Connect(context.Background()); err != nil {
		slog.Error("failed to reconnect", "error", err)
		panic(err)
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
