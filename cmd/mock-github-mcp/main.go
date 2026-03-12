package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/agentfence/agentfence/internal/mcp/protocol"
)

func main() {
	var addr string
	flags := flag.NewFlagSet("mock-github-mcp", flag.ExitOnError)
	flags.StringVar(&addr, "addr", ":8090", "listen address")
	_ = flags.Parse(os.Args[1:])

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleMCP)

	log.Printf("mock GitHub MCP listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var request protocol.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, protocol.ID{}, -32700, "invalid JSON-RPC request")
		return
	}
	if err := request.Validate(); err != nil {
		writeError(w, valueOrZero(request.ID), -32600, err.Error())
		return
	}
	if request.ID == nil || request.ID.IsZero() {
		writeError(w, protocol.ID{}, -32600, "notifications are not supported")
		return
	}

	switch request.Method {
	case protocol.MethodToolsList:
		result := protocol.ToolsListResult{Tools: []protocol.ToolDefinition{
			{Name: "repos/get", Description: "Fetch repository metadata."},
			{Name: "pulls/list", Description: "List pull requests for a repository."},
			{Name: "issues/comment", Description: "Create a GitHub issue comment."},
			{Name: "pulls/merge", Description: "Merge a pull request."},
			{Name: "repos/delete", Description: "Delete a GitHub repository."},
		}}
		writeResult(w, *request.ID, result)
	case protocol.MethodToolsCall:
		handleToolsCall(w, request)
	default:
		writeError(w, *request.ID, -32601, fmt.Sprintf("method %q not found", request.Method))
	}
}

func handleToolsCall(w http.ResponseWriter, request protocol.Request) {
	params, err := protocol.DecodeToolsCallParams(request.Params)
	if err != nil {
		writeError(w, *request.ID, -32602, err.Error())
		return
	}

	var arguments map[string]any
	if len(params.Arguments) > 0 {
		if err := json.Unmarshal(params.Arguments, &arguments); err != nil {
			writeError(w, *request.ID, -32602, "invalid tool arguments")
			return
		}
	}
	if arguments == nil {
		arguments = map[string]any{}
	}

	var summary string
	switch params.Name {
	case "repos/get":
		summary = fmt.Sprintf("Read repository %s", repoValue(arguments))
	case "pulls/list":
		summary = fmt.Sprintf("List pull requests for %s", repoValue(arguments))
	case "issues/comment":
		summary = fmt.Sprintf("Comment on issue %v in %s", arguments["issue_number"], repoValue(arguments))
	case "pulls/merge":
		summary = fmt.Sprintf("Merge PR %v in %s", arguments["pull_number"], repoValue(arguments))
	case "repos/delete":
		summary = fmt.Sprintf("Delete repository %s", repoValue(arguments))
	default:
		writeError(w, *request.ID, -32601, fmt.Sprintf("tool %q not found", params.Name))
		return
	}

	result := protocol.ToolsCallResult{
		Content: []protocol.ToolContent{{Type: "text", Text: summary}},
		StructuredContent: mustJSON(map[string]any{
			"tool": params.Name,
			"received_arguments": arguments,
			"server": "mock-github",
		}),
	}
	writeResult(w, *request.ID, result)
}

func repoValue(arguments map[string]any) string {
	owner, _ := arguments["owner"].(string)
	repo, _ := arguments["repo"].(string)
	if owner == "" && repo == "" {
		return "unknown/unknown"
	}
	return owner + "/" + repo
}

func valueOrZero(id *protocol.ID) protocol.ID {
	if id == nil {
		return protocol.ID{}
	}
	return *id
}

func writeResult(w http.ResponseWriter, id protocol.ID, result any) {
	response := protocol.Response{JSONRPC: protocol.JSONRPCVersion, ID: id, Result: mustJSON(result)}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, id protocol.ID, code int64, message string) {
	response := protocol.Response{JSONRPC: protocol.JSONRPCVersion, ID: id, Error: &protocol.Error{Code: code, Message: message}}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(response)
}

func mustJSON(value any) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}