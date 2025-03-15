package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/ryzmae/hyperatomic/internal/config"
	"github.com/ryzmae/hyperatomic/internal/logger"
)

type TCPServer struct {
	listener net.Listener
	log      *logger.Logger
}

func NewServer(cfg *config.Config, log *logger.Logger) (*TCPServer, error) {
	port := cfg.TCP.Port

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	return &TCPServer{listener: listener, log: log}, nil
}

func (s *TCPServer) HandleConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.log.Log("ERROR", "Failed to accept connection: %v", err)
			continue
		}
		go s.handleRequest(conn)
	}
}

func (s *TCPServer) handleRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		data, err := reader.ReadString('\n')
		if err != nil {
			s.log.Log("ERROR", "Failed to read data: %v", err)
			return
		}

		message := strings.TrimSpace(data)
		s.log.Log("INFO", "Received: %s", message)

		// Send response
		conn.Write([]byte("ACK\n"))
	}
}
