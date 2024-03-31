package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type WebSocketManager struct {
	sync.RWMutex
	clients        ClientList
	handlers       map[string]EventHandler
	TrustedOrigins []string
	server         *Server
	log            *logger.Logger
}

func NewWebSocketManager(s *Server) *WebSocketManager {
	m := &WebSocketManager{
		log:      logger.New("WebSocket-Manager"),
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		server:   s,
	}

	m.setupEventHandlers()
	return m
}

func (m *WebSocketManager) setupEventHandlers() {
	m.handlers[EventSendMessage] = m.SendMessage
}

// SendMessage sends messages to the clients
func (m *WebSocketManager) SendMessage(event Event, c *Client) error {

	var sendEvent SendMessageEvent
	err := json.Unmarshal(event.Payload, &sendEvent)
	if err != nil {
		return fmt.Errorf("bad payload in the request: %v", err)
	}

	var broadMessage NewMessageEvent
	broadMessage.Sent = time.Now()
	broadMessage.Message = sendEvent.Message
	broadMessage.From = sendEvent.From
	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal the broadcast: %v", err)
	}

	outgoingEvent := Event{
		Type:    EventNewMessage,
		Payload: data,
	}

	for client := range c.manager.clients {
		client.egress <- outgoingEvent
	}

	return nil
}

func (m *WebSocketManager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		err := handler(event, c)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("there is no such event type: %s", event.Type)
}

func (m *WebSocketManager) serveWS(w http.ResponseWriter, r *http.Request) {

	ticket := r.URL.Query().Get("token")
	session, _, err := m.server.SessionService.Get(r.Context(), ticket)
	if err != nil {
		InvalidSession(w)
		return
	}

	if !session.User.Activated {
		NotAuthenticaded(w)
		return
	}

	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Status:  StatusErr,
			Message: "no topic_id",
		})
		return
	}

	websocketUpgrader.CheckOrigin = m.checkOrigin

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.log.Errorf("%v", err)
		return
	}

	// defer conn.Close()
	client := NewWebSocketClient(conn, m, topicID)
	m.addClient(client)

	// Start client processes
	go client.readMessages()
	go client.writeMessage()

	client.log.Debugf("client connectet to websocket")
}

func (m *WebSocketManager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = struct{}{}
}

func (m *WebSocketManager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok {
		if client.sub != nil {
			client.sub.Close()
		}
		if client.conn != nil {
			client.conn.Close()
		}
		delete(m.clients, client)
	}
}

func (m *WebSocketManager) checkOrigin(r *http.Request) bool {
	if m.server.IsDev {
		m.log.Warnf("!! Server runs in dev environment !!")
		return true
	}
	origin := r.Header.Get("Origin")
	for _, to := range m.TrustedOrigins {
		if origin == to {
			return true
		}
	}
	return false
}
