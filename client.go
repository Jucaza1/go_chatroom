package main

import (
	"bytes"
	"context"
	"math/rand"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	messageSizeMax = 512
)

type ClientMessage struct {
	id      uint32
	message []byte
}

type Client struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	id        uint32
	hub       *Hub
	send      chan *ClientMessage
	conn      *websocket.Conn
}

func NewClient(ctx context.Context, cancel context.CancelFunc, hub *Hub, conn *websocket.Conn) Client {
	return Client{
		ctx:       ctx,
		cancelCtx: cancel,
        id: rand.Uint32(),
		hub:       hub,
		send:      make(chan *ClientMessage,256), conn: conn,
	}
}

func (c *Client) writeConn() {
	ticker := time.NewTicker(pingPeriod)
	defer c.shutdown()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(pingPeriod))
			if !ok {
				//send channel is closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte("disconnected"))
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				//TODO handle error
				return
			}
			w.Write([]byte(fmt.Sprintf("%d", message.id)))
			w.Write([]byte{' ', ':', '\n'})
			w.Write(message.message)
			w.Write([]byte{'\n'})
            //important to close the writer
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(pingPeriod))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				//TODO handle error
				return
			}
		case <-c.ctx.Done():
			return
		}
	}

}

func (c *Client) readConn() {
	defer c.shutdown()
	c.conn.SetReadLimit(messageSizeMax)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				//TODO handle error
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}
			if messageType == websocket.TextMessage {
				c.hub.broadcast <- &ClientMessage{
					id:      c.id,
					message: bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1)),
				}
			}
		}
	}
}

func (c *Client) shutdown() {
	c.hub.unregister <- c
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	c.conn.Close()
	if _, ok := <-c.send; ok {
		close(c.send)
	}
}

func MakeWSHandler(ctx context.Context, hub *Hub) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithCancel(ctx)
		client := NewClient(ctx, cancel, hub, conn)
		hub.register <- &client
		go client.readConn()
		go client.writeConn()
	}
}
