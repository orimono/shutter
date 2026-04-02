package ws

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/orimono/shutter/internal/config"
	"github.com/orimono/shutter/internal/protocol"
	"github.com/orimono/shutter/internal/util"
)

type Client struct {
	cfg     *config.Config
	session *Session
	ready   chan struct{}
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:   cfg,
		ready: make(chan struct{}, 1),
	}
}

func (c *Client) Run(ctx context.Context) {
	dialer := websocket.DefaultDialer

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		util.DrainChan(c.ready)
		conn, _, err := dialer.Dial(c.cfg.ServerURL, nil)
		if err != nil {
			slog.Error("Failed to connect to server", "error", err)
			time.Sleep(time.Duration(c.cfg.RetryInterval))
			continue
		}
		c.session = NewSession(conn, c.cfg)
		c.ready <- struct{}{}
		c.session.run(ctx)
	}
}

func (c *Client) Send(data []byte) error {
	session := c.session
	if session == nil {
		return fmt.Errorf("no active session")
	}
	session.send <- protocol.Message{
		Type: websocket.TextMessage,
		Data: data,
	}
	return nil
}

func (c *Client) Ready() <-chan struct{} {
	return c.ready
}
