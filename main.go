package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) HandleWS(ws *websocket.Conn) {
	fmt.Println("New incoming connection from client: ", ws.RemoteAddr())
	// TODO: add mutex lock here
	s.conns[ws] = true
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from client: ", err)
			continue
		}
		msg := buf[:n]
		s.Broadcast(msg)
		fmt.Println("Broadcasted message: ", string(msg))
	}
}

func (s *Server) HandleWSOrderBook(ws *websocket.Conn) {
	fmt.Println("New incoming connection from client: to orderbook feed", ws.RemoteAddr())
	for {
		payload := fmt.Sprintf("order book data -> %d\n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(1 * time.Second)
	}
}

func (s *Server) AddConn(conn *websocket.Conn) {
	s.conns[conn] = true
}

func (s *Server) RemoveConn(conn *websocket.Conn) {
	delete(s.conns, conn)
}

func (s *Server) Broadcast(b []byte) {
	for conn := range s.conns {
		go func(conn *websocket.Conn) {
			if _, err := conn.Write(b); err != nil {
				fmt.Println("Error broadcasting message: ", err)
			}
		}(conn)
	}
}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.HandleWS))
	http.Handle("/orderbookfeed", websocket.Handler(server.HandleWSOrderBook))
	http.ListenAndServe(":3000", nil)
}
