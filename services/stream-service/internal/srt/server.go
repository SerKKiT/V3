package srt

import (
	"context"
	"fmt"
	"log"
	"time"

	gosrt "github.com/datarhei/gosrt"
)

type Server struct {
	listener gosrt.Listener
	config   *Config
	handler  *Handler
}

type Config struct {
	Address string
	Latency uint // в миллисекундах
}

func NewServer(cfg *Config, handler *Handler) (*Server, error) {
	config := gosrt.DefaultConfig()
	// Конвертируем uint в time.Duration (миллисекунды)
	config.Latency = time.Duration(cfg.Latency) * time.Millisecond

	ln, err := gosrt.Listen("srt", cfg.Address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start SRT listener: %w", err)
	}

	log.Printf("✅ SRT listener created on %s", cfg.Address)
	return &Server{
		listener: ln,
		config:   cfg,
		handler:  handler,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	log.Printf("🚀 SRT server started on %s with latency %dms", s.config.Address, s.config.Latency)

	for {
		select {
		case <-ctx.Done():
			log.Println("⏹️  SRT server shutting down")
			return ctx.Err()
		default:
			// Accept incoming connection requests
			req, err := s.listener.Accept2()
			if err != nil {
				log.Printf("❌ Accept error: %v", err)
				continue
			}

			// Handle connection in separate goroutine
			go s.handleConnection(req)
		}
	}
}

func (s *Server) handleConnection(req gosrt.ConnRequest) {
	streamID := req.StreamId()
	log.Printf("📡 New SRT connection request, StreamID: %s", streamID)

	// Validate stream key before accepting
	if !s.handler.ValidateStreamKey(streamID) {
		log.Printf("❌ Rejecting connection: invalid stream key %s", streamID)
		req.Reject(gosrt.REJ_PEER)
		return
	}

	// Pass the ConnRequest directly to handler
	// Handler will accept the connection internally
	s.handler.HandlePublish(req)
}

func (s *Server) Stop() error {
	if s.listener != nil {
		// Close() does not return a value in gosrt
		s.listener.Close()
		log.Println("✅ SRT listener closed")
	}
	return nil
}
