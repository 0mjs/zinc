---
slug: /cookbook/realtime-chat-app
title: Realtime Chat App
description: Build a small live chat UI with Zinc templates, static assets, and a WebSocket endpoint.
---

This recipe combines:

- server-rendered HTML
- static browser assets
- a normal JSON health route
- a WebSocket endpoint for realtime updates

## Setup

```bash
go mod init zinc-realtime-chat
go get github.com/0mjs/zinc
```

## Project layout

```text
.
├── main.go
├── public
│   └── chat.js
└── templates
    └── chat.html
```

## Application

```go
package main

import (
	"html/template"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0mjs/zinc"
)

type ChatMessage struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time string `json:"time"`
}

type Hub struct {
	mu      sync.Mutex
	clients map[*zinc.WebSocketConn]string
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*zinc.WebSocketConn]string)}
}

func (h *Hub) Add(conn *zinc.WebSocketConn, user string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = user
}

func (h *Hub) Remove(conn *zinc.WebSocketConn) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	user := h.clients[conn]
	delete(h.clients, conn)
	return user
}

func (h *Hub) Broadcast(message ChatMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		if err := conn.WriteJSON(message); err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func main() {
	views := template.Must(template.ParseGlob("templates/*.html"))
	hub := NewHub()

	app := zinc.NewWithConfig(zinc.Config{
		Renderer: zinc.NewHTMLTemplateRenderer(
			views,
			zinc.WithTemplateSuffixes(".html"),
		),
	})

	if err := app.Static("/static", "./public"); err != nil {
		log.Fatal(err)
	}

	app.Get("/", func(c *zinc.Context) error {
		return c.Render("chat", zinc.Map{
			"Title": "Realtime Chat App",
			"Room":  "Lobby",
		})
	})

	app.Get("/health", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{"ok": true})
	})

	if err := app.WS("/ws", func(c *zinc.Context, conn *zinc.WebSocketConn) error {
		user := strings.TrimSpace(c.Query("user"))
		if user == "" {
			user = "guest-" + strconv.FormatInt(time.Now().Unix()%1000, 10)
		}

		conn.SetReadLimit(4096)
		hub.Add(conn, user)
		hub.Broadcast(ChatMessage{
			User: "system",
			Text: user + " joined the room",
			Time: time.Now().Format("15:04:05"),
		})

		defer func() {
			leftUser := hub.Remove(conn)
			if leftUser != "" {
				hub.Broadcast(ChatMessage{
					User: "system",
					Text: leftUser + " left the room",
					Time: time.Now().Format("15:04:05"),
				})
			}
		}()

		for {
			var incoming struct {
				Text string `json:"text"`
			}

			if err := conn.ReadJSON(&incoming); err != nil {
				return nil
			}

			incoming.Text = strings.TrimSpace(incoming.Text)
			if incoming.Text == "" {
				continue
			}

			hub.Broadcast(ChatMessage{
				User: user,
				Text: incoming.Text,
				Time: time.Now().Format("15:04:05"),
			})
		}
	}, zinc.WebSocketConfig{
		CheckOrigin: func(c *zinc.Context) bool {
			origin := c.GetHeader(zinc.HeaderOrigin)
			if origin == "" {
				return true
			}
			host := c.Request().Host
			return origin == "http://"+host || origin == "https://"+host
		},
	}); err != nil {
		log.Fatal(err)
	}

	app.Listen()
}
```

## What this demonstrates

- templates for the first render
- `Static` for browser assets
- a normal JSON route next to the socket route
- `app.WS(...)` for WebSocket lifecycle and origin checks

## Run it

```bash
go run .
# open http://localhost:8080
```

Open a second tab with a different `?user=` query value and you have a compact realtime example to extend.
