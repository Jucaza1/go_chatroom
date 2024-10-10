package main

import (
	"context"
	"fmt"
)

type Hub struct {
	ctx        context.Context
	clients    map[*Client]bool
	broadcast  chan *ClientMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub(ctx context.Context) Hub {
	return Hub{
        ctx: ctx,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *ClientMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	fmt.Println("--hub running")
	defer h.shutdown()
	for {
		select {
		case <-h.ctx.Done():
			return
		case client, ok := <-h.register:
			if !ok {
				return
			}
			h.clients[client] = true
		case client, ok := <-h.unregister:
			if !ok {
				return
			}
			delete(h.clients, client)
		case message, ok := <-h.broadcast:
			if !ok {
				return
			}
			for client := range h.clients {
                client.send<-message
			}
		}
	}
}

func (h *Hub) unreg(client *Client) {
	if _, ok := h.clients[client]; ok {
		client.cancelCtx()
		delete(h.clients, client)
	}
}

func (h *Hub) reg(client *Client) {
	h.clients[client] = true
}

func (h *Hub) shutdown() {
	close(h.register)
	close(h.unregister)
	close(h.broadcast)
}
