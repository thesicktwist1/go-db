package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"thesicktwist1/go-db/Internal/frame"
	"time"
)

var (
	ErrPeerDisconnected = errors.New("peer disconnected")
	ErrClosedCh         = errors.New("channel closed or full")
	ErrServerClosed     = errors.New("server closed")
)

// Processor interface defines the contract for processing frames from peers.
type Processor interface {
	Process(Peer, frame.Frame) error
	Options
}

// Options interface defines timeouts and buffer configurations for operations.
type Options interface {
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
}

// Connection interface defines the contract for network connections.
type Connection interface {
	io.ReadWriteCloser
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
	RemoteAddr() net.Addr
}

// Server represents the database server.
type Server struct {
	Store
	ServerOpts
	listener net.Listener // Active listener for cleanup
}

type ServerOpts struct {
	ListenAddr   string        // Address to listen on
	readTimeout  time.Duration // Timeout for read operations
	writeTimeout time.Duration // Timeout for write operations
	bufferSize   int           // Size of request/response buffers (bytes)
	channelSize  int           // Size of async operation channels
}

// NewServer creates a new server with the given database file and options.
func NewServer(fileName string, opts ServerOpts) (*Server, error) {
	serverOptsDefaults(&opts)
	store, err := NewStore(fileName)
	if err != nil {
		return nil, err
	}
	return &Server{
		Store:      store,
		ServerOpts: opts,
	}, nil
}

// Start begins listening for incoming connections and processes them.
func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = listener
	go s.Store.Run(ctx)
	slog.Info("server listening on", "addr", s.ListenAddr)
	return s.acceptLoop(ctx, listener)
}

// acceptLoop accepts incoming connections and spawns handlers for them.
func (s *Server) acceptLoop(ctx context.Context, listener net.Listener) error {
	for {
		select {
		case <-ctx.Done():
			slog.Info("server shutting down, stopping accepting connections")
			return listener.Close()
		default:
		}
		conn, err := listener.Accept()
		if err != nil {
			// Check if this is due to shutdown
			select {
			case <-ctx.Done():
				return nil
			default:
				slog.Error("listener", "err", err)
				continue
			}
		}
		go s.handleConnection(ctx, conn)
	}
}

// Process handles incoming frames from a peer by queuing them as transactions or control messages.
func (s *Server) Process(peer Peer, frm frame.Frame) error {
	switch f := frm.(type) {
	case *frame.Control:
		var payload []byte
		switch f.Op {
		case frame.OpAuth:
			// todo
		case frame.OpPing:
			payload = frame.Pong
		case frame.OpClosing:
			peer.Close()
			return nil
		}
		return peer.Respond(payload)
	case *frame.Query:
		pid := NewPID(peer.RemoteAddr(), f.ID)
		if err := s.QueueTx(transaction{
			peer:  peer,
			Frame: *f,
			pid:   pid,
		}); err != nil {
			return err
		}
	default:
		return errors.New("invalid frame type")
	}
	return nil
}

// handleConnection manages a peer connection, handling read and write operations.
func (s *Server) handleConnection(ctx context.Context, conn Connection) {
	addr := conn.RemoteAddr().String()
	peer := NewPeer(s, conn)
	slog.Info("connection received", "addr", addr)

	newCtx, cancel := context.WithCancel(ctx)
	defer func() {
		slog.Info("closing connection", "addr", addr)
		conn.Write(frame.Closing)
		conn.Close()
	}()
	peer.cancel = cancel

	go peer.readLoop(newCtx)
	peer.writeLoop(newCtx)
}

// ReadTimeout returns the read timeout for the server.
func (s *Server) ReadTimeout() time.Duration {
	return s.readTimeout
}

// WriteTimeout returns the write timeout for the server.
func (s *Server) WriteTimeout() time.Duration {
	return s.writeTimeout
}

// DefaultServerOpts returns the default server options.
func DefaultServerOpts() ServerOpts {
	return ServerOpts{
		ListenAddr:   defaultServerAddr,
		readTimeout:  defaultReadTimeout,
		writeTimeout: defaultWriteTimeout,
		bufferSize:   defaultBufferSize,
		channelSize:  defaultChannelSize,
	}
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	slog.Info("initiating graceful shutdown")
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}
	slog.Info("graceful shutdown successful")
	return nil
}
