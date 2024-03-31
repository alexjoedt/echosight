package http

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/alexjoedt/echosight/internal/eventflow"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

var (
	pongWait     time.Duration = time.Second * 10
	pingInterval time.Duration = (pongWait * 9) / 10
)

// ClientList.
// TODO: user map[string]*Client (clientID)
type ClientList map[*Client]struct{}

type Client struct {
	ID      string
	topicID string // topic is the topic where the client subscribes on the internal pub-sub/event engine
	sub     *eventflow.Subscription
	conn    *websocket.Conn
	manager *WebSocketManager
	egress  chan Event
	log     *logger.Logger
}

func NewWebSocketClient(conn *websocket.Conn, manager *WebSocketManager, topicID string) *Client {
	c := &Client{
		ID:      xid.New().String(),
		topicID: topicID,
		conn:    conn,
		manager: manager,
		egress:  make(chan Event),
	}

	c.log = logger.New("WebSocket-Client", logger.Str("client_id", c.ID), logger.Str("topic_id", topicID))
	sub, err := manager.server.EventHandler.Subscribe(context.Background(), topicID, c.onCheckEvent)
	if err != nil {
		c.log.Errorf("failed to subscribe to topic: %v", err)
	} else {
		c.sub = sub
	}
	return c
}

func (c *Client) onCheckEvent(ctx context.Context, event *eventflow.Event) error {
	c.egress <- Event{Type: event.Type, Payload: event.Payload}
	return nil
}

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		c.log.Errorf("failed to set read deadline: %v", err)
		return
	}

	// sets a read limit
	c.conn.SetReadLimit(512)

	// PongHandler to keep connection alive, or check if the client is still connected
	c.conn.SetPongHandler(c.pongHandler)

	for {
		messageType, payload, err := c.conn.ReadMessage()
		if err != nil {
			c.log.Debugf("websocket connection closed: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error(err, "failed to read websocket message")
				c.conn.Close()
			}

			if errors.Is(err, websocket.ErrReadLimit) {
				c.log.Errorf("%v", err)
			}

			break
		}

		switch messageType {
		case websocket.PongMessage:

		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			c.log.Errorf("failed to unmarshal event: %v", err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			c.log.Errorf("failed to route event; %v", err)
		}

	}
}

func (c *Client) writeMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					c.log.Errorf("websocket connection closed: %v", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				c.log.Errorf("failed to marshal message to write: %v", err)
				continue
			}

			// Message
			err = c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				c.log.Errorf("failed to send message: %v", err)
			} else {
				c.log.Debugf("message sent: '%s'", string(data))
			}

		case <-ticker.C:
			c.log.Debugf("send ping to client")
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				c.log.Errorf("failed to ping client: %v", err)
				return
			}
		}
	}
}

func (c *Client) pongHandler(msg string) error {
	c.log.Debugf("received pong from client")
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
