package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ole-larsen/binance-subscriber/internal/helpers"
	"github.com/ole-larsen/binance-subscriber/internal/log"
)

var (
	logger             = log.NewLogger("info", log.DefaultBuildLogger)
	defaultIDSize      = 10
	defaultPingTimeout = 1 * time.Minute
)

type BinanceRequest struct {
	ID     interface{} `json:"id"`
	Method string      `json:"method"`
	Params []string    `json:"params"`
}

type BinancePoller struct {
	Conn         *websocket.Conn
	msg          chan []byte
	BaseEndpoint string
}

func NewBinancePoller() *BinancePoller {
	return &BinancePoller{
		BaseEndpoint: "wss://stream.binance.com:9443/stream",
		msg:          make(chan []byte),
	}
}

func (c *BinancePoller) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.BaseEndpoint, nil)
	if err != nil {
		return NewError(err)
	}

	c.Conn = conn

	c.Ping()

	return nil
}

func (c *BinancePoller) Close() error {
	if c.Conn == nil {
		return ErrConnectionNotInitialized
	}

	return c.Conn.Close()
}

func (c *BinancePoller) Read() error {
	if c.Conn == nil {
		return ErrConnectionNotInitialized
	}

	defer close(c.msg)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			return NewError(fmt.Errorf("failed to read message: %w", err))
		}

		c.msg <- message
	}
}

func (c *BinancePoller) GetMsg() chan []byte {
	return c.msg
}

func (c *BinancePoller) Ping() {
	if c.Conn == nil {
		return
	}

	c.Conn.SetPingHandler(func(appData string) error {
		if err := c.Conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(defaultPingTimeout)); err != nil {
			logger.Errorln(NewError(err))
			return err
		}

		return nil
	})
}

func (c *BinancePoller) Subscribe(streams []string) error {
	if c.Conn == nil {
		return ErrConnectionNotInitialized
	}

	req := BinanceRequest{
		Method: "SUBSCRIBE",
		Params: streams,
		ID:     helpers.RandStringBytes(defaultIDSize),
	}

	return c.Conn.WriteJSON(req)
}

func (c *BinancePoller) Unsubscribe(streams []string) error {
	if c.Conn == nil {
		return ErrConnectionNotInitialized
	}

	req := BinanceRequest{
		Method: "UNSUBSCRIBE",
		Params: streams,
		ID:     helpers.RandStringBytes(defaultIDSize),
	}

	return c.Conn.WriteJSON(req)
}
