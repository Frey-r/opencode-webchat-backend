package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ebachmann/opencode-webchat/internal/apikey"
	"github.com/ebachmann/opencode-webchat/internal/opencode"
	"github.com/ebachmann/opencode-webchat/internal/store"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	opencodeMgr *opencode.Manager
	store       *store.Store

	clients    map[int]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage

	mu sync.RWMutex
}

type BroadcastMessage struct {
	SessionID int
	Data      []byte
}

func NewHub(om *opencode.Manager, s *store.Store) *Hub {
	return &Hub{
		opencodeMgr: om,
		store:       s,
		clients:     make(map[int]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.sessionID]; !ok {
				h.clients[client.sessionID] = make(map[*Client]bool)
			}
			h.clients[client.sessionID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.sessionID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.sessionID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.SessionID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- msg.Data:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, sessionID, userID int, workDir string) {
	log.Printf("[WS] HandleWS upgrading connection for session=%d user=%d dir=%s", sessionID, userID, workDir)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] upgrade failed: %v (conn=%v)", err, conn)
		return
	}

	log.Printf("[WS] upgrade success, creating client for session=%d user=%d", sessionID, userID)

	client := newClient(h, conn, sessionID, userID, workDir)
	h.register <- client

	ctx, cancel := context.WithCancel(context.Background())
	go client.writePump(ctx)
	go client.readPump(ctx)
	go func() {
		<-ctx.Done()
		cancel()
	}()
}

func (h *Hub) handleMessage(client *Client, msg *InboundMessage) {
	switch msg.Type {
	case TypePing:
		client.sendJSON(OutboundMessage{Type: TypePong})

	case TypePrompt:
		go h.handlePrompt(client, msg)

	case TypeCancel:
		h.handleCancel(client)

	default:
		client.sendJSON(OutboundMessage{
			Type:    TypeError,
			Content: "unknown message type",
		})
	}
}

func (h *Hub) handlePrompt(client *Client, msg *InboundMessage) {
	settings, _ := h.store.GetUserSettings(context.Background(), client.userID)
	env := apikey.BuildEnv(settings)

	h.store.CreateMessage(context.Background(), client.sessionID, "user", msg.Content, nil, "completed")

	model := ""
	if m, ok := settings["OPENCODE_MODEL"]; ok && m != "" {
		model = m
	}

	sessionInfo := h.opencodeMgr.GetProcessInfo(client.sessionID)
	if sessionInfo == nil {
		h.opencodeMgr.RegisterSession(client.sessionID, env)
		sessionInfo = h.opencodeMgr.GetProcessInfo(client.sessionID)
	}

	ocSessionID := ""
	if sessionInfo != nil {
		ocSessionID = sessionInfo.OpenCodeSessionID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := h.opencodeMgr.RunPrompt(ctx, msg.Content, model, ocSessionID, client.workDir, env, func(result opencode.RunResult) {
		switch result.EventType {
		case "text", "assistant", "content":
			if result.Content == "" {
				break
			}
			h.broadcastToSession(client.sessionID, OutboundMessage{
				Type:    TypeToken,
				Content: result.Content,
			})
		case "step_start", "step_finish":
			// silently ignore step events

		case "tool-call", "tool_call", "tool_use":
			h.broadcastToSession(client.sessionID, OutboundMessage{
				Type:    TypeToolCall,
				Content: result.Content,
				Data:    result.Data,
			})
		case "tool-result", "tool_result", "tool_output":
			h.broadcastToSession(client.sessionID, OutboundMessage{
				Type:    TypeToolResult,
				Content: result.Content,
				Data:    result.Data,
			})
		case "error":
			h.broadcastToSession(client.sessionID, OutboundMessage{
				Type:    TypeError,
				Content: result.Content,
			})
		}
	})

	if err != nil {
		log.Printf("opencode run failed: %v", err)
		h.broadcastToSession(client.sessionID, OutboundMessage{
			Type:    TypeError,
			Content: "failed to get response from opencode",
		})
	}

	h.broadcastToSession(client.sessionID, OutboundMessage{
		Type: TypeDone,
	})
}

func (h *Hub) handleCancel(client *Client) {
	client.sendJSON(OutboundMessage{
		Type:    TypeDone,
		Content: "cancelled",
	})
}

func (h *Hub) broadcastToSession(sessionID int, msg OutboundMessage) {
	data, _ := json.Marshal(msg)
	h.broadcast <- &BroadcastMessage{
		SessionID: sessionID,
		Data:      data,
	}
}