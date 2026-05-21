package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = 30 * time.Second
	maxMessageSize = 512 * 1024
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	sessionID int
	userID    int
	workDir   string
	send      chan []byte
	mu        sync.Mutex
}

func newClient(hub *Hub, conn *websocket.Conn, sessionID, userID int, workDir string) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		sessionID: sessionID,
		userID:    userID,
		workDir:   workDir,
		send:      make(chan []byte, 256),
	}
}

func (c *Client) readPump(ctx context.Context) {
	log.Printf("[WS] readPump started for user=%d session=%d", c.userID, c.sessionID)
	defer func() {
		log.Printf("[WS] readPump ended for user=%d session=%d", c.userID, c.sessionID)
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws error: %v", err)
			}
			break
		}

		var msg InboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}

		c.hub.handleMessage(c, &msg)
	}
}

func (c *Client) writePump(ctx context.Context) {
	log.Printf("[WS] writePump started for user=%d session=%d", c.userID, c.sessionID)
	ticker := time.NewTicker(pingInterval)
	defer func() {
		log.Printf("[WS] writePump ended for user=%d session=%d", c.userID, c.sessionID)
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.mu.Unlock()
				return
			}
			w.Write(msg)
			c.mu.Unlock()

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) sendJSON(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return ErrSendBufferFull
	}
}

var ErrSendBufferFull = context.DeadlineExceeded