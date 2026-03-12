package transport

import (
	"strings"
	"testing"
)

func TestTargetValidate(t *testing.T) {
	tests := []struct {
		name    string
		target  Target
		wantErr string
	}{
		{
			name:   "http valid",
			target: Target{Network: NetworkHTTP, Address: "http://127.0.0.1:8080/mcp"},
		},
		{
			name:   "stdio valid",
			target: Target{Network: NetworkStdio, Command: "agent-mcp"},
		},
		{
			name:    "http missing address",
			target:  Target{Network: NetworkHTTP},
			wantErr: "address is required",
		},
		{
			name:    "stdio missing command",
			target:  Target{Network: NetworkStdio},
			wantErr: "command is required",
		},
		{
			name:    "unsupported network",
			target:  Target{Network: "tcp"},
			wantErr: "unsupported target network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Validate() error = nil, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
				}
			}
		})
	}
}
