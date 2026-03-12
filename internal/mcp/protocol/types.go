package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const JSONRPCVersion = "2.0"

const (
	MethodToolsList = "tools/list"
	MethodToolsCall = "tools/call"
)

// ID is a JSON-RPC request identifier limited to string or integer values.
type ID struct {
	String *string
	Number *int64
}

func StringID(value string) ID {
	return ID{String: &value}
}

func IntID(value int64) ID {
	return ID{Number: &value}
}

func (id ID) IsZero() bool {
	return id.String == nil && id.Number == nil
}

func (id ID) StringValue() (string, bool) {
	if id.String == nil {
		return "", false
	}
	return *id.String, true
}

func (id ID) IntValue() (int64, bool) {
	if id.Number == nil {
		return 0, false
	}
	return *id.Number, true
}

func (id ID) MarshalJSON() ([]byte, error) {
	if id.String != nil {
		return json.Marshal(*id.String)
	}
	if id.Number != nil {
		return json.Marshal(*id.Number)
	}
	return []byte("null"), nil
}

func (id *ID) UnmarshalJSON(data []byte) error {
	if id == nil {
		return errors.New("jsonrpc id target is nil")
	}

	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*id = ID{}
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(trimmed, &stringValue); err == nil {
		*id = StringID(stringValue)
		return nil
	}

	var numberValue int64
	if err := json.Unmarshal(trimmed, &numberValue); err == nil {
		*id = IntID(numberValue)
		return nil
	}

	return fmt.Errorf("jsonrpc id must be a string, integer, or null")
}

// Request is a JSON-RPC request or notification.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Validate checks whether the request matches the JSON-RPC shape AgentFence supports.
func (r Request) Validate() error {
	if r.JSONRPC != JSONRPCVersion {
		return fmt.Errorf("jsonrpc must be %q", JSONRPCVersion)
	}
	if r.Method == "" {
		return errors.New("method is required")
	}
	return nil
}

func (r Request) IsNotification() bool {
	return r.ID == nil || r.ID.IsZero()
}

// Response is a JSON-RPC success or error response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      ID              `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Validate checks whether the response matches the JSON-RPC shape AgentFence supports.
func (r Response) Validate() error {
	if r.JSONRPC != JSONRPCVersion {
		return fmt.Errorf("jsonrpc must be %q", JSONRPCVersion)
	}
	if r.ID.IsZero() {
		return errors.New("id is required")
	}

	hasResult := len(r.Result) > 0
	hasError := r.Error != nil
	if hasResult == hasError {
		return errors.New("exactly one of result or error is required")
	}
	if hasError {
		if err := r.Error.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r Response) IsError() bool {
	return r.Error != nil
}

// Error is a JSON-RPC error object.
type Error struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e Error) Validate() error {
	if e.Message == "" {
		return errors.New("error.message is required")
	}
	return nil
}

func DecodeRequest(data []byte) (Request, error) {
	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		return Request{}, fmt.Errorf("decode jsonrpc request: %w", err)
	}
	if err := request.Validate(); err != nil {
		return Request{}, err
	}
	return request, nil
}

func DecodeResponse(data []byte) (Response, error) {
	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return Response{}, fmt.Errorf("decode jsonrpc response: %w", err)
	}
	if err := response.Validate(); err != nil {
		return Response{}, err
	}
	return response, nil
}

// ToolsListParams is the parameter shape for `tools/list`.
type ToolsListParams struct{}

// ToolDefinition describes a tool exposed by an upstream MCP server.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
}

// ToolsListResult is the result shape for `tools/list`.
type ToolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolsCallParams is the parameter shape for `tools/call`.
type ToolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

func (p ToolsCallParams) Validate() error {
	if p.Name == "" {
		return errors.New("tools/call name is required")
	}
	return nil
}

// ToolContent represents one item from a tool call result content array.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolsCallResult is the result shape for `tools/call`.
type ToolsCallResult struct {
	Content           []ToolContent   `json:"content,omitempty"`
	StructuredContent json.RawMessage `json:"structuredContent,omitempty"`
	IsError           bool            `json:"isError,omitempty"`
}

func DecodeToolsCallParams(raw json.RawMessage) (ToolsCallParams, error) {
	var params ToolsCallParams
	if len(raw) == 0 {
		return ToolsCallParams{}, errors.New("tools/call params are required")
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return ToolsCallParams{}, fmt.Errorf("decode tools/call params: %w", err)
	}
	if err := params.Validate(); err != nil {
		return ToolsCallParams{}, err
	}
	return params, nil
}
