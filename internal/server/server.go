package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

// Contains the state of the server
type Server struct {
	Listener net.Listener
	IsClosed atomic.Bool
}

// Creates a net.Listener and returns a new Server instance.
// Starts listening for requests inside a goroutine.
func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
		Listener: l,
	}

	go server.listen()

	return &server, nil
}

// Closes the listener and the server
func (s *Server) Close() error {
	s.IsClosed.Store(true)
	err := s.Listener.Close()
	if err != nil {
		return err
	}

	return nil
}

// Uses a loop to .Accept new connections as they come in, and handles each one in a new goroutine.
// I used an atomic.Bool to track whether the server is closed or not
// so that I can ignore connection errors after the server is closed.
// https://pkg.go.dev/net#Listener.Accept
// https://pkg.go.dev/sync/atomic#Bool
func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.IsClosed.Load() {
				// Server was closed, exit gracefully
				return
			}
			log.Fatal(err)
		}

		go s.handle(conn)
	}
}

// Handles a single connection by writing the following response and then closing the connection:
/*
HTTP/1.1 200 OK
Content-Type: text/plain

Hello World!
*/
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	defaultResponse := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!")

	conn.Write(defaultResponse)
}
