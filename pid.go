package main

import "fmt"

type pid struct {
	addr string
	id   uint32
}

// NewPID creates a new process ID with the given remote address and ID.
func NewPID(addr string, id uint32) *pid {
	return &pid{
		addr: addr,
		id:   id,
	}
}

// String returns the string representation of the PID.
func (p *pid) String() string {
	return fmt.Sprintf("addr=%s id=%d", p.addr, p.id)
}

// RemoteAddr returns the addr as a string
func (p *pid) RemoteAddr() string {
	return p.addr
}

// ID returns the id as a string
func (p *pid) ID() string {
	return fmt.Sprintf("%d", p.id)
}
