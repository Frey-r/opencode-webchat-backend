package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

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
	Exclude   *Client
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
						h.opencodeMgr.StopProcess(client.sessionID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.SessionID]
			h.mu.RUnlock()

			for client := range clients {
				if msg.Exclude != nil && client == msg.Exclude {
					continue
				}
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

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, sessionID, userID int) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := newClient(h, conn, sessionID, userID)
	h.register <- client

	go client.writePump(r.Context())
	go client.readPump(r.Context())
}

func (h *Hub) handleMessage(client *Client, msg *InboundMessage) {
	switch msg.Type {
	case TypePing:
		client.sendJSON(OutboundMessage{Type: TypePong})

	case TypePrompt:
		h.handlePrompt(client, msg)

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
	proc := h.opencodeMgr.GetProcess(client.sessionID)
	if proc == nil {
		var err error
		proc, err = h.opencodeMgr.StartProcess(client.sessionID)
		if err != nil {
			client.sendJSON(OutboundMessage{
				Type:    TypeError,
				Content: "failed to start opencode process",
			})
			return
		}
		go h.streamOutput(client, proc)
	}

	h.store.CreateMessage(context.Background(), client.sessionID, "user", msg.Content, nil, "completed")

	if err := proc.Write(context.Background(), msg.Content); err != nil {
		client.sendJSON(OutboundMessage{
			Type:    TypeError,
			Content: "failed to write to opencode",
		})
	}
}

func (h *Hub) streamOutput(client *Client, proc *opencode.Process) {
	go func() {
		if err := proc.Wait(); err != nil {
			log.Printf("opencode process ended: %v", err)
		}
	}()

	for {
		line, err := proc.ReadLine(context.Background())
		if err != nil {
			break
		}

		outMsg := OutboundMessage{
			Type:    TypeToken,
			Content: line,
		}
		data, _ := json.Marshal(outMsg)

		h.broadcast <- &BroadcastMessage{
			SessionID: client.sessionID,
			Data:      data,
		}
	}

	doneMsg := OutboundMessage{Type: TypeDone}
	data, _ := json.Marshal(doneMsg)
	h.broadcast <- &BroadcastMessage{
		SessionID: client.sessionID,
		Data:      data,
	}
}

func (h *Hub) handleCancel(client *Client) {
	h.opencodeMgr.StopProcess(client.sessionID)
	client.sendJSON(OutboundMessage{
		Type:    TypeDone,
		Content: "cancelled",
	})
}