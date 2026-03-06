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
	Process(Peer, frame.Frame)
	Options
}

// Options interface defines timeouts and buffer configurations for operations.
type Options interface {
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
	BufferSize() int
	ChannelSize() int
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
	shutdown chan struct{} // Signals graceful shutdown
	listener net.Listener  // Active listener for cleanup
}

type ServerOpts struct {
	ListenAddr   string        // Address to listen on
	readTimeout  time.Duration // Timeout for read operations
	writeTimeout time.Duration // Timeout for write operations
	bufferSize   int           // Size of request/response buffers (bytes)
	channelSize  int           // Size of async operation channels
}

// NewConfig creates a new server configuration with the given listen address.
func NewConfig(listenAddr string) ServerOpts {
	return ServerOpts{
		ListenAddr:   listenAddr,
		readTimeout:  defaultReadTimeout,
		writeTimeout: defaultWriteTimeout,
		bufferSize:   defaultBufferSize,
		channelSize:  defaultChannelSize,
	}
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

// NewServer creates a new server with the given database file and options.
func NewServer(fileName string, opts ServerOpts) (*Server, error) {
	if opts.bufferSize <= 0 {
		opts.bufferSize = defaultBufferSize
	}
	if opts.channelSize <= 0 {
		opts.channelSize = defaultChannelSize
	}
	store, err := NewStore(fileName)
	if err != nil {
		return nil, err
	}
	return &Server{
		Store:      store,
		ServerOpts: opts,
		shutdown:   make(chan struct{}),
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
	slog.Info("server listening on", "addr", listener.Addr().String())
	return s.acceptLoop(ctx, listener)
}

// acceptLoop accepts incoming connections and spawns handlers for them.
func (s *Server) acceptLoop(ctx context.Context, listener net.Listener) error {
	for {
		select {
		case <-s.shutdown:
			slog.Info("server shutting down, stopping accepting connections")
			return listener.Close()
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			// Check if this is due to shutdown
			select {
			case <-s.shutdown:
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
func (s *Server) Process(peer Peer, f frame.Frame) {
	pid := NewPID(peer.RemoteAddr(), f.Id)
	switch f.Type {
	case frame.TypeControl:
		if err := handleControl(peer, f); err != nil {
			slog.Error("control", "err", err)
		}
	case frame.TypeQuery:
		if err := s.QueueTx(transaction{
			pid:   pid,
			peer:  peer,
			Frame: f,
		}); err != nil {
			slog.Error("queue", "err", err)
		}
	}
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

// processControlMessage processes control frames such as auth, ping, and closing.
func handleControl(peer Peer, req frame.Frame) error {
	var payload []byte
	switch req.Op {
	case frame.OpAuth:
		// auth
	case frame.OpPing:
		payload = frame.Pong
	case frame.OpClosing:
		peer.Close()
		return nil
	}
	return peer.Respond(payload)
}

// ReadTimeout returns the read timeout for the server.
func (s *Server) ReadTimeout() time.Duration {
	return s.readTimeout
}

// WriteTimeout returns the write timeout for the server.
func (s *Server) WriteTimeout() time.Duration {
	return s.writeTimeout
}

// BufferSize returns the buffer size for the server.
func (s *Server) BufferSize() int {
	return s.bufferSize
}

// ChannelSize returns the channel size for the server.
func (s *Server) ChannelSize() int {
	return s.channelSize
}
func (s *Server) Shutdown() error {
	slog.Info("initiating graceful shutdown")
	close(s.shutdown)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
