package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

type Server struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewServer() *Server {
	return &Server{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (s *Server) addClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = struct{}{}
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, conn)
	_ = conn.Close()
}

func (s *Server) broadcast(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("write message error: %v", err)
			_ = conn.Close()
			delete(s.clients, conn)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// demo 用，正式環境請限制 origin
		return true
	},
}

func (s *Server) handleResourcesWS(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade websocket failed: %v", err)
		return
	}

	s.addClient(conn)
	log.Printf("client connected: %s", conn.RemoteAddr())

	go func() {
		defer func() {
			s.removeClient(conn)
			log.Printf("client disconnected: %s", conn.RemoteAddr())
		}()

		// 持續 read 是為了能偵測 client 關閉
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}
