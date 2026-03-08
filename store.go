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
	compInterval = storeCompactionInterval
	txChSize     = storeChannelSize
)

type transaction struct {
	peer  Peer
	Frame frame.Query
	*pid
}

// Store interface defines the contract for database storage operations.
type Store interface {
	Queue(transaction) error
	Run(context.Context)
}

// store implements the Store interface for in-memory database operations.
type store struct {
	txCh chan transaction
	mem  map[string][]byte
	log  *logs.Log
}

// QueueTx queues a transaction for processing.
func (s *store) Queue(tx transaction) error {
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
			val, err := s.executeTransaction(tx)
			if err != nil {
				resp.Write(frame.NewError(tx.id, err).Bytes())
				slog.Error("transaction processing", "err", err, "pid", tx.pid)
			} else {
				resp.Write(frame.NewPayload(tx.id, val).Bytes())
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

// executeTransaction processes a single transaction and returns the result.
func (s *store) executeTransaction(tx transaction) ([]byte, error) {
	var (
		value  []byte
		exists bool
	)
	f := tx.Frame
	switch f.Op {
	case frame.OpGet:
		value, exists = s.mem[string(f.Buffer)]
		if !exists {
			return nil, ErrNotExists
		}
	case frame.OpDel:
		key := string(f.Buffer)
		_, exists := s.mem[key]
		if !exists {
			return nil, ErrNotExists
		}
		if err := s.log.Append(f.Buffer, int(f.KeyLen)); err != nil {
			return nil, err
		}
		delete(s.mem, key)
	case frame.OpSet:
		key := string(f.Buffer[:f.KeyLen])
		newValue := f.Buffer[f.KeyLen:]
		if err := s.log.Append(f.Buffer, int(f.KeyLen)); err != nil {
			return nil, err
		}
		s.mem[key] = newValue
	default:
		return nil, frame.ErrMalformed
	}
	return value, nil
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
