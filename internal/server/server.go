package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/neixir/httpfromtcp/internal/request"
	"github.com/neixir/httpfromtcp/internal/response"
)

// Contains the state of the server
type Server struct {
	Listener net.Listener
	IsClosed atomic.Bool
	Handler  HandlerFunc
}

type HandlerFunc func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode int
	Message    string
}

// Creates a net.Listener and returns a new Server instance.
// Starts listening for requests inside a goroutine.
func Serve(port int, handler HandlerFunc) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := Server{
		Listener: l,
		Handler:  handler,
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

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Parse the request from the connection
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println(err)
		// TODO De fet, envia 500 Bad Request
		return
	}

	// Create a new empty bytes.Buffer for the handler to write to
	res := response.NewWriter(conn)

	// Call the handler function
	s.Handler(res, req)

}
