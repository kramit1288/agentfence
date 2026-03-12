package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
)

func TestHTTPForwarderSuccess(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request protocol.Request
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("json.Decode() error = %v", err)
		}
		_ = json.NewEncoder(w).Encode(protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			ID:      *request.ID,
			Result:  mustMarshal(map[string]any{"ok": true}),
		})
	}))
	defer upstream.Close()

	forwarder := mustHTTPForwarder(t, upstream.URL, &http.Client{Timeout: time.Second})
	result, err := forwarder.Forward(context.Background(), "deployer", protocol.Request{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      idPtr(protocol.StringID("req-1")),
		Method:  protocol.MethodToolsCall,
		Params:  mustMarshal(map[string]any{"name": "deploy"}),
	})
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}
	if result.Outcome != OutcomeSuccess || result.HTTPStatusCode != http.StatusOK {
		t.Fatalf("result = %+v, want success 200", result)
	}
}

func TestHTTPForwarderUpstreamDecodeError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`not-jsonrpc`))
	}))
	defer upstream.Close()

	forwarder := mustHTTPForwarder(t, upstream.URL, &http.Client{Timeout: time.Second})
	result, err := forwarder.Forward(context.Background(), "deployer", protocol.Request{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      idPtr(protocol.StringID("req-1")),
		Method:  protocol.MethodToolsCall,
		Params:  mustMarshal(map[string]any{"name": "deploy"}),
	})
	if err == nil {
		t.Fatal("Forward() error = nil, want error")
	}
	if result.Outcome != OutcomeHTTPError {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeHTTPError)
	}
}

func TestHTTPForwarderTimeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			ID:      protocol.StringID("req-1"),
			Result:  mustMarshal(map[string]any{"ok": true}),
		})
	}))
	defer upstream.Close()

	forwarder := mustHTTPForwarder(t, upstream.URL, &http.Client{Timeout: 10 * time.Millisecond})
	result, err := forwarder.Forward(context.Background(), "deployer", protocol.Request{
		JSONRPC: protocol.JSONRPCVersion,
		ID:      idPtr(protocol.StringID("req-1")),
		Method:  protocol.MethodToolsCall,
		Params:  mustMarshal(map[string]any{"name": "deploy"}),
	})
	if err == nil {
		t.Fatal("Forward() error = nil, want error")
	}
	if result.Outcome != OutcomeTransportError {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeTransportError)
	}
}

func mustHTTPForwarder(t *testing.T, url string, client *http.Client) *HTTPForwarder {
	t.Helper()
	forwarder, err := NewHTTPForwarder(Target{Address: url}, client)
	if err != nil {
		t.Fatalf("NewHTTPForwarder() error = %v", err)
	}
	return forwarder
}

func mustMarshal(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func idPtr(id protocol.ID) *protocol.ID { return &id }
