package main

import "fmt"

type pid struct {
	RemoteAddr string
	ID         uint32
}

// NewPID creates a new process ID with the given remote address and ID.
func NewPID(addr string, id uint32) *pid {
	return &pid{
		RemoteAddr: addr,
		ID:         id,
	}
}

// String returns the string representation of the PID.
func (p *pid) String() string {
	return fmt.Sprintf("addr=%s id=%d", p.RemoteAddr, p.ID)
}
