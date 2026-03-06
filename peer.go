package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"thesicktwist1/go-db/Internal/frame"
	"time"
)

type Peer interface {
	RemoteAddr() string
	Respond([]byte) error
	io.Closer
}

type peer struct {
	conn    Connection
	respsCh chan []byte
	cancel  context.CancelFunc
	Processer
}

// NewPeer creates a new peer with the given processer and connection.
func NewPeer(p Processer, conn Connection) *peer {
	return &peer{
		conn:      conn,
		respsCh:   make(chan []byte, 32),
		Processer: p,
	}
}

// readLoop continuously reads incoming messages from the connection and processes them.
func (p *peer) readLoop(ctx context.Context) {
	buf := make([]byte, 1024)
	bufLen := 0
	parser := frame.NewParser()
outer:
	for {
		p.conn.SetReadDeadline(time.Now().Add(p.ReadTimeout()))
		for !parser.Done() {
			n, err := p.conn.Read(buf[bufLen:])
			if err != nil {
				logReadError(ctx, err)
				break outer
			}
			bufLen += n
			readN, err := parser.Parse(buf[:bufLen])
			if err != nil {
				slog.Error("parse", "err", err, "peer", p.RemoteAddr())
				break outer
			}
			copy(buf, buf[readN:bufLen])
			bufLen -= readN
		}
		slog.Info("message received from", "peer", p.RemoteAddr())
		reqCopy := parser.Frame
		reqCopy.Buffer = *bytes.NewBuffer(parser.Frame.Buffer.Bytes())
		p.Process(p, reqCopy)
		parser.Reset()
	}
}

// Close closes the peer connection by canceling its context.
func (p *peer) Close() error {
	p.cancel()
	return nil
}

// writeLoop continuously writes response messages to the connection.
func (p *peer) writeLoop(ctx context.Context) {
loop:
	for {
		p.conn.SetWriteDeadline(time.Now().Add(p.WriteTimeout()))
		select {
		case resp := <-p.respsCh:
			_, err := p.conn.Write(resp)
			if err != nil {
				slog.Error("write", "err", err, "peer", p.RemoteAddr())
				break loop
			}
		case <-ctx.Done():
			break loop
		}
	}
}

// Respond sends a response message to the peer.
func (p *peer) Respond(b []byte) error {
	select {
	case p.respsCh <- b:
		slog.Info("message sent to", "peer", p.RemoteAddr())
		return nil
	default:
		return ErrClosedCh
	}
}

// RemoteAddr returns the remote address of the peer.
func (p *peer) RemoteAddr() string {
	return p.conn.RemoteAddr().String()
}

// logReadError logs read errors, ignoring EOF errors and errors from canceled contexts.
func logReadError(ctx context.Context, err error) {
	select {
	case <-ctx.Done():
	default:
		if !errors.Is(err, io.EOF) {
			slog.Error("read", "err", err)
		}
	}
}
