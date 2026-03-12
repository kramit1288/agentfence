package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeRequestRoundTripToolCall(t *testing.T) {
	input := []byte(`{"jsonrpc":"2.0","id":"req-1","method":"tools/call","params":{"name":"deploy","arguments":{"env":"staging"}}}`)

	request, err := DecodeRequest(input)
	if err != nil {
		t.Fatalf("DecodeRequest() error = %v", err)
	}
	if request.Method != MethodToolsCall {
		t.Fatalf("Method = %q, want %q", request.Method, MethodToolsCall)
	}
	if request.IsNotification() {
		t.Fatal("IsNotification() = true, want false")
	}
	id, ok := request.ID.StringValue()
	if !ok || id != "req-1" {
		t.Fatalf("ID = %q, want req-1", id)
	}

	params, err := DecodeToolsCallParams(request.Params)
	if err != nil {
		t.Fatalf("DecodeToolsCallParams() error = %v", err)
	}
	if params.Name != "deploy" {
		t.Fatalf("Name = %q, want deploy", params.Name)
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(encoded), `"method":"tools/call"`) {
		t.Fatalf("json.Marshal() = %s, want method field", encoded)
	}
}

func TestDecodeResponseRoundTripResult(t *testing.T) {
	input := []byte(`{"jsonrpc":"2.0","id":7,"result":{"tools":[{"name":"search"}]}}`)

	response, err := DecodeResponse(input)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}
	if response.IsError() {
		t.Fatal("IsError() = true, want false")
	}
	id, ok := response.ID.IntValue()
	if !ok || id != 7 {
		t.Fatalf("ID = %d, want 7", id)
	}

	var result ToolsListResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
	if len(result.Tools) != 1 || result.Tools[0].Name != "search" {
		t.Fatalf("Tools = %+v, want one search tool", result.Tools)
	}
}

func TestDecodeRequestRejectsMalformedJSON(t *testing.T) {
	_, err := DecodeRequest([]byte(`{"jsonrpc":"2.0",`))
	if err == nil {
		t.Fatal("DecodeRequest() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "decode jsonrpc request") {
		t.Fatalf("DecodeRequest() error = %v, want decode context", err)
	}
}

func TestDecodeRequestRejectsMissingJSONRPC(t *testing.T) {
	_, err := DecodeRequest([]byte(`{"id":"req-1","method":"tools/list"}`))
	if err == nil {
		t.Fatal("DecodeRequest() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "jsonrpc must be \"2.0\"") {
		t.Fatalf("DecodeRequest() error = %v, want jsonrpc validation", err)
	}
}

func TestDecodeRequestRejectsMissingMethod(t *testing.T) {
	_, err := DecodeRequest([]byte(`{"jsonrpc":"2.0","id":"req-1"}`))
	if err == nil {
		t.Fatal("DecodeRequest() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "method is required") {
		t.Fatalf("DecodeRequest() error = %v, want missing method validation", err)
	}
}

func TestDecodeResponseRejectsMissingID(t *testing.T) {
	_, err := DecodeResponse([]byte(`{"jsonrpc":"2.0","result":{}}`))
	if err == nil {
		t.Fatal("DecodeResponse() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("DecodeResponse() error = %v, want missing id validation", err)
	}
}

func TestDecodeResponseRejectsResultAndErrorTogether(t *testing.T) {
	_, err := DecodeResponse([]byte(`{"jsonrpc":"2.0","id":1,"result":{},"error":{"code":-32000,"message":"boom"}}`))
	if err == nil {
		t.Fatal("DecodeResponse() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "exactly one of result or error is required") {
		t.Fatalf("DecodeResponse() error = %v, want mutually-exclusive validation", err)
	}
}

func TestDecodeRequestRejectsInvalidIDType(t *testing.T) {
	_, err := DecodeRequest([]byte(`{"jsonrpc":"2.0","id":true,"method":"tools/list"}`))
	if err == nil {
		t.Fatal("DecodeRequest() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "jsonrpc id") {
		t.Fatalf("DecodeRequest() error = %v, want id validation", err)
	}
}

func TestDecodeToolsCallParamsRejectsMissingName(t *testing.T) {
	_, err := DecodeToolsCallParams(json.RawMessage(`{"arguments":{"force":true}}`))
	if err == nil {
		t.Fatal("DecodeToolsCallParams() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("DecodeToolsCallParams() error = %v, want missing name validation", err)
	}
}

