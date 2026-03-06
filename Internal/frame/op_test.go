package frame

import "testing"

func TestOp_Has(t *testing.T) {
	tests := []struct {
		name     string
		op       Op
		checkOp  Op
		expected bool
	}{
		{"OpGet has OpGet", OpGet, OpGet, true},
		{"OpGet does not have OpSet", OpGet, OpSet, false},
		{"OpDefault has OpDefault", OpDefault, OpDefault, true},
		{"OpSet has OpSet", OpSet, OpSet, true},
		{"OpDel does not have OpAuth", OpDel, OpAuth, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.Has(tt.checkOp); got != tt.expected {
				t.Errorf("Op(%d).Has(%d) = %v, want %v", tt.op, tt.checkOp, got, tt.expected)
			}
		})
	}
}

func TestOp_String(t *testing.T) {
	tests := []struct {
		name     string
		op       Op
		expected string
	}{
		{"OpDefault", OpDefault, "[INVALID]"},
		{"OpGet", OpGet, "GET"},
		{"OpDel", OpDel, "DELETE"},
		{"OpSet", OpSet, "SET"},
		{"OpAuth", OpAuth, "AUTH"},
		{"OpPing", OpPing, "PING"},
		{"OpPong", OpPong, "PONG"},
		{"OpClosing", OpClosing, "CLOSING"},
		{"Unknown Op", Op(99), "[INVALID]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.String(); got != tt.expected {
				t.Errorf("Op(%d).String() = %v, want %v", tt.op, got, tt.expected)
			}
		})
	}
}
