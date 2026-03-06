package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"thesicktwist1/go-db/Internal/frame"
	"thesicktwist1/go-db/Internal/logs"

	"time"
)

var (
	ErrNotExists = errors.New("key doesn't exists")
)

const (
	syncInterval = time.Millisecond * 5
	compInterval = time.Minute * 5
	txChSize     = 32
)

type transaction struct {
	pid  *pid
	peer Peer
	frame.Frame
}

// Store interface defines the contract for database storage operations.
type Store interface {
	QueueTx(transaction) error
	Run(context.Context)
}

// store implements the Store interface for in-memory database operations.
type store struct {
	txCh chan transaction
	mem  map[string][]byte
	log  *logs.Log
}

// QueueTx queues a transaction for processing.
func (s *store) QueueTx(tx transaction) error {
	select {
	case s.txCh <- tx:
		slog.Info("transaction added to queue", "PID", tx.pid.String())
		return nil
	default:
		return ErrClosedCh
	}
}

// Run starts the main transaction processing loop.
func (s *store) Run(ctx context.Context) {
	resp := bytes.NewBuffer(nil)
loop:
	for {
		select {
		case tx := <-s.txCh:
			slog.Info("processing transaction", "pid", tx.pid)
			val, err := s.processTask(tx)
			if err != nil {
				resp.Write(frame.Error(tx.Id, err))
				slog.Error("transaction processing", "err", err, "pid", tx.pid)
			} else {
				resp.Write(frame.Payload(tx.Id, val))
			}
			payload := make([]byte, resp.Len())
			resp.Read(payload)
			if err := tx.peer.Respond(payload); err != nil {
				slog.Error("couldn't reach peer",
					"err", err, "peer", tx.peer.RemoteAddr())
				continue
			}
			slog.Info("transaction successful", "pid", tx.pid)
			resp.Reset()
		case <-ctx.Done():
			break loop
		}
	}
}

// processTask processes a single transaction and returns the result.
func (s *store) processTask(tx transaction) ([]byte, error) {
	var (
		val    []byte
		exists bool
	)
	payload := tx.Buffer
	switch tx.Op {
	case frame.OpGet:
		key := payload.String()
		val, exists = s.mem[key]
		if !exists {
			return nil, ErrNotExists
		}
	case frame.OpDel:
		_, exists := s.mem[payload.String()]
		if !exists {
			return nil, ErrNotExists
		}
		if err := s.log.Append(payload.Bytes(), tx.KeyLen); err != nil {
			return nil, err
		}
		delete(s.mem, payload.String())
	case frame.OpSet:
		key := tx.Buffer.String()[:tx.KeyLen]
		newVal := tx.Buffer.Bytes()[tx.KeyLen:]
		if err := s.log.Append(tx.Buffer.Bytes(), tx.KeyLen); err != nil {
			return nil, err
		}
		s.mem[key] = newVal
	default:
		return nil, frame.ErrMalformed
	}
	return val, nil
}

// NewStore creates a new store from the given database file.
func NewStore(filePath string) (*store, error) {
	mem := make(map[string][]byte)
	l, err := logs.New(filePath, mem)
	if err != nil {
		return nil, err
	}
	s := &store{
		txCh: make(chan transaction, txChSize),
		mem:  mem,
		log:  l,
	}
	return s, nil
}
