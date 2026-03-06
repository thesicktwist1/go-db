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

const (
	defaultAddr         = ":4040"
	defaultReadTimeout  = time.Second * 30
	defaultWriteTimeout = time.Second * 30
)

// Processer interface defines the contract for processing frames from peers.
type Processer interface {
	Process(Peer, frame.Frame)
	Options
}

// Options interface defines timeouts for read and write operations.
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
	actualAddr string // Stores the actual listening address (useful when using :0)
}

type ServerOpts struct {
	ListenAddr   string
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// NewConfig creates a new server configuration with the given listen address.
func NewConfig(listenAddr string) ServerOpts {
	return ServerOpts{
		ListenAddr: listenAddr,
	}
}

// defaultOpts returns the default server options.
func defaultOpts() ServerOpts {
	return ServerOpts{
		ListenAddr:   defaultAddr,
		readTimeout:  defaultReadTimeout,
		writeTimeout: defaultWriteTimeout,
	}
}

// NewServer creates a new server with the given database file and options.
func NewServer(fileName string, ServerOpts ServerOpts) (*Server, error) {
	store, err := NewStore(fileName)
	if err != nil {
		return nil, err
	}
	return &Server{
		Store:      store,
		ServerOpts: ServerOpts,
	}, nil
}

// Start begins listening for incoming connections and processes them.
func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.actualAddr = listener.Addr().String()
	go s.Store.Run(ctx)
	slog.Info("server listening on", "addr", s.actualAddr)
	return s.acceptLoop(ctx, listener)
}

// acceptLoop accepts incoming connections and spawns handlers for them.
func (s *Server) acceptLoop(ctx context.Context, listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("listener", "err", err)
			continue
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

// GetAddr returns the actual listening address of the server.
func (s *Server) GetAddr() string {
	if s.actualAddr != "" {
		return s.actualAddr
	}
	return s.ListenAddr
}

// ReadTimeout returns the read timeout for the server.
func (s *Server) ReadTimeout() time.Duration {
	return s.readTimeout
}

// WriteTimeout returns the write timeout for the server.
func (s *Server) WriteTimeout() time.Duration {
	return s.writeTimeout
}
