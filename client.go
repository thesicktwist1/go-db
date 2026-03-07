package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"thesicktwist1/go-db/Internal/frame"

	"time"
)

type responses map[uint32]chan response

type response struct {
	val []byte
	err error
}

type Client struct {
	conn   Connection
	resps  responses
	reqCh  chan []byte
	cancel context.CancelFunc
	ClientOpts
	sync.RWMutex
}

type ClientOpts struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	BufferSize   int // Size of read/write buffers (bytes)
	ChannelSize  int // Size of request/response channels
}

// NewClient creates and starts a new client connected to the given address with the given options.
func NewClient(address string, opts ClientOpts) (*Client, error) {
	clientOptsDefaults(&opts)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	client := &Client{
		conn:       conn,
		resps:      make(map[uint32]chan response),
		reqCh:      make(chan []byte, opts.ChannelSize),
		cancel:     cancel,
		ClientOpts: opts,
	}
	go client.start(ctx)
	return client, nil
}

// Get retrieves the value for the given key from the server.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	id := rand.Uint32()
	request := frame.Query(frame.OpGet, id, key, nil)
	respCh, err := c.newRequestChannel(id, request)
	if err != nil {
		return nil, err
	}
	select {
	case resp := <-respCh:
		return resp.val, resp.err
	case <-ctx.Done():
		return nil, context.Canceled
	}
}

// Set sets the value for the given key on the server.
func (c *Client) Set(ctx context.Context, key string, val []byte) error {
	id := rand.Uint32()
	request := frame.Query(frame.OpSet, id, key, val)
	respCh, err := c.newRequestChannel(id, request)
	if err != nil {
		return err
	}
	select {
	case resp := <-respCh:
		return resp.err
	case <-ctx.Done():
		return context.Canceled
	}
}

// Delete deletes the key from the server.
func (c *Client) Delete(ctx context.Context, key string) error {
	id := rand.Uint32()
	request := frame.Query(frame.OpDel, id, key, nil)
	respCh, err := c.newRequestChannel(id, request)
	if err != nil {
		return err
	}
	select {
	case resp := <-respCh:
		return resp.err
	case <-ctx.Done():
		return context.Canceled
	}
}

// newRequestChannel creates a new response channel for a request and sends it.
func (c *Client) newRequestChannel(id uint32, request []byte) (chan response, error) {
	c.Lock()
	defer c.Unlock()
	_, exist := c.resps[id]
	if exist {
		return nil, fmt.Errorf("request id [%d] already exist", id)
	}
	respCh := make(chan response, 1)
	c.resps[id] = respCh
	select {
	case c.reqCh <- request:
	default:
		delete(c.resps, id)
		close(respCh)
		return nil, ErrClosedCh
	}
	return respCh, nil
}

// start initializes the client's read and write loops.
func (c *Client) start(ctx context.Context) {
	defer func() {
		slog.Info("closing client gracefully")
		c.cancel()
		c.conn.Write(frame.Closing)
		c.conn.Close()
	}()
	go c.readLoop(ctx)
	c.writeLoop(ctx)
}

// readLoop continuously reads responses from the server.
func (c *Client) readLoop(ctx context.Context) {
	buf := make([]byte, c.BufferSize)
	bufLen := 0
	parser := frame.NewParser()
outer:
	for {
		for !parser.Done() {
			c.conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
			n, err := c.conn.Read(buf[bufLen:])
			if err != nil {
				logReadError(ctx, err)
				break outer
			}
			bufLen += n
			readN, err := parser.Parse(buf[:bufLen])
			if err != nil {
				slog.Error("parse", "err", err)
				break outer
			}
			copy(buf, buf[readN:bufLen])
			bufLen -= readN
		}
		frameCopy := parser.Frame
		frameCopy.Buffer = *bytes.NewBuffer(parser.Frame.Buffer.Bytes())
		c.handleResponse(frameCopy)
		parser.Reset()
	}
}

// writeLoop continuously writes requests to the server and sends periodic pings.
func (c *Client) writeLoop(ctx context.Context) {
	pingTicker := time.NewTicker(clientPingInterval)
	defer pingTicker.Stop()
loop:
	for {
		c.conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
		select {
		case <-pingTicker.C:
			select {
			case c.reqCh <- frame.Ping:
			default:
				break loop
			}
		case <-ctx.Done():
			break loop
		case req := <-c.reqCh:
			_, err := c.conn.Write(req)
			if err != nil {
				slog.Error("write", "err", err)
				break loop
			}
		}
	}
}

// sendResponse sends a response to the corresponding request channel.
func (c *Client) sendResponse(id uint32, resp response) {
	c.RLock()
	defer c.RUnlock()
	ch, exist := c.resps[id]
	if !exist {
		slog.Error("response channel", "err", fmt.Errorf("id [%d] doesn't exist", id))
		return
	}
	select {
	case ch <- resp:
		slog.Info("response successfully received")
		close(ch)
		delete(c.resps, id)
	default:
		slog.Error("response channel", "err", ErrClosedCh)
		return
	}
}

// handleResponse processes different types of frames received from the server.
func (c *Client) handleResponse(f frame.Frame) {
	var resp response
	switch f.Type {
	case frame.TypePayload:
		resp.val = f.Buffer.Bytes()
		c.sendResponse(f.Id, resp)
	case frame.TypeError:
		resp.err = errors.New(f.Buffer.String())
		c.sendResponse(f.Id, resp)
	case frame.TypeControl:
		c.handleControl(f)
	}
}

// handleControl handles control frames from the server.
func (c *Client) handleControl(f frame.Frame) {
	switch f.Op {
	case frame.OpAuth:
		// to do
	case frame.OpClosing:
		c.cancel()
	}
}
