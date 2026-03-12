package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
)

// HTTPForwarder forwards MCP JSON-RPC calls to one upstream HTTP endpoint.
type HTTPForwarder struct {
	target Target
	client *http.Client
}

func NewHTTPForwarder(target Target, client *http.Client) (*HTTPForwarder, error) {
	target.Network = NetworkHTTP
	if err := target.Validate(); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &HTTPForwarder{target: target, client: client}, nil
}

func (f *HTTPForwarder) Forward(ctx context.Context, server string, request protocol.Request) (ForwardResult, error) {
	started := time.Now()
	target := sanitizeAddress(f.target.Address)
	raw, err := json.Marshal(request)
	if err != nil {
		return ForwardResult{Latency: time.Since(started), Outcome: OutcomeTransportError, Target: target, Err: err}, fmt.Errorf("encode upstream request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, f.target.Address, bytes.NewReader(raw))
	if err != nil {
		return ForwardResult{Latency: time.Since(started), Outcome: OutcomeTransportError, Target: target, Err: err}, fmt.Errorf("build upstream request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if server != "" {
		httpRequest.Header.Set("X-AgentFence-Server", server)
	}
	for key, value := range f.target.Headers {
		httpRequest.Header.Set(key, value)
	}

	httpResponse, err := f.client.Do(httpRequest)
	latency := time.Since(started)
	if err != nil {
		result := ForwardResult{Latency: latency, Outcome: OutcomeTransportError, Target: target, Err: err}
		return result, fmt.Errorf("call upstream: %w", err)
	}
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(io.LimitReader(httpResponse.Body, 1<<20))
	if err != nil {
		result := ForwardResult{Latency: latency, Outcome: OutcomeTransportError, Target: target, HTTPStatusCode: httpResponse.StatusCode, Err: err}
		return result, fmt.Errorf("read upstream response: %w", err)
	}

	response, err := protocol.DecodeResponse(body)
	if err != nil {
		result := ForwardResult{Latency: latency, Outcome: OutcomeHTTPError, Target: target, HTTPStatusCode: httpResponse.StatusCode, Err: err}
		return result, fmt.Errorf("decode upstream response: %w", err)
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		err = fmt.Errorf("upstream returned HTTP %d", httpResponse.StatusCode)
		result := ForwardResult{
			Response:       response,
			HTTPStatusCode: httpResponse.StatusCode,
			Latency:        latency,
			Outcome:        OutcomeHTTPError,
			Target:         target,
			Err:            err,
		}
		return result, err
	}

	outcome := OutcomeSuccess
	if response.IsError() {
		outcome = OutcomeRPCError
	}

	result := ForwardResult{
		Response:       response,
		HTTPStatusCode: httpResponse.StatusCode,
		Latency:        latency,
		Outcome:        outcome,
		Target:         target,
	}
	return result, nil
}

func sanitizeAddress(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}