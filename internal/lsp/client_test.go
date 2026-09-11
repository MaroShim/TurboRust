package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPathURIRoundtrip(t *testing.T) {
	testPaths := []string{
		"/Users/maro/Projects/tr/main.rs",
		"/tmp/test file with spaces.rs",
		"src/main.rs",
	}

	for _, p := range testPaths {
		uri := PathToURI(p)
		if !strings.HasPrefix(uri, "file://") {
			t.Errorf("expected file:// prefix, got %s", uri)
		}
		back := URIToPath(uri)
		absP, _ := filepath.Abs(p)
		if back != absP && back != p {
			t.Errorf("roundtrip mismatch: orig=%s, uri=%s, back=%s, abs=%s", p, uri, back, absP)
		}
	}
}

func TestHoverSnippetCleaning(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "```rust\npub fn hello(name: &str) -> String\n```\nDocumentation here",
			expected: "pub fn hello(name: &str) -> String",
		},
		{
			input:    "```rust\nstruct Point {\n    x: i32,\n    y: i32,\n}\n```",
			expected: "struct Point {",
		},
		{
			input:    "fn test()",
			expected: "fn test()",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for i, tc := range cases {
		got := cleanHoverSnippet(tc.input)
		if got != tc.expected {
			t.Errorf("case %d: expected %q, got %q", i, tc.expected, got)
		}
	}
}

// TestMockLSPClientInteraction verifies JSON-RPC stdio protocol framing using mock pipe pair.
func TestMockLSPClientInteraction(t *testing.T) {
	// Create piped readers and writers to emulate a full LSP server process
	serverInR, clientInW := io.Pipe()   // client writes to clientInW, server reads from serverInR
	clientOutR, serverOutW := io.Pipe() // server writes to serverOutW, client reads from clientOutR

	var wg sync.WaitGroup
	wg.Add(1)

	// Mock Server Goroutine
	go func() {
		defer wg.Done()
		defer serverOutW.Close()
		defer serverInR.Close()

		reader := bufio.NewReader(serverInR)
		for {
			// Read framing
			contentLength := -1
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					break
				}
				if strings.HasPrefix(strings.ToLower(line), "content-length:") {
					val := strings.TrimSpace(line[len("content-length:"):])
					contentLength, _ = strconv.Atoi(val)
				}
			}
			if contentLength <= 0 {
				continue
			}

			body := make([]byte, contentLength)
			if _, err := io.ReadFull(reader, body); err != nil {
				return
			}

			var req struct {
				ID     int64           `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			_ = json.Unmarshal(body, &req)

			// Route responses
			var result interface{}
			switch req.Method {
			case "initialize":
				result = map[string]interface{}{
					"capabilities": map[string]interface{}{
						"definitionProvider": true,
						"hoverProvider":      true,
					},
				}
			case "textDocument/definition":
				result = map[string]interface{}{
					"uri": "file:///workspace/target.rs",
					"range": map[string]interface{}{
						"start": map[string]interface{}{"line": 15, "character": 4},
						"end":   map[string]interface{}{"line": 15, "character": 12},
					},
				}
			case "textDocument/hover":
				result = map[string]interface{}{
					"contents": map[string]interface{}{
						"kind":  "markdown",
						"value": "```rust\npub fn calculate(val: i32) -> i32\n```",
					},
				}
			case "shutdown":
				result = nil
			case "exit":
				return
			default:
				// Ignore notifications like initialized, didOpen, didChange
				continue
			}

			resJSON, _ := json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  result,
			})
			header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(resJSON))
			_, _ = serverOutW.Write([]byte(header))
			_, _ = serverOutW.Write(resJSON)
		}
	}()

	// Construct Client wired to the mock pipes
	client := &Client{
		serverName: "mock-rust-analyzer",
		binPath:    "/mock/rust-analyzer",
		rootDir:    "/workspace",
		stdin:      clientInW,
		stdout:     clientOutR,
		reader:     bufio.NewReader(clientOutR),
		pending:    make(map[int64]chan *rpcResponse),
	}

	client.closeWg.Add(1)
	go client.readLoop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Initialize
	if err := client.initialize(ctx); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	client.isReady.Store(true)

	if !client.IsAvailable() {
		t.Fatalf("expected client to be available")
	}

	// 2. DidOpen & DidChange
	if err := client.DidOpen("/workspace/main.rs", "fn main() {}"); err != nil {
		t.Errorf("DidOpen failed: %v", err)
	}
	if err := client.DidChange("/workspace/main.rs", "fn main() {\n    let x = 1;\n}"); err != nil {
		t.Errorf("DidChange failed: %v", err)
	}

	// 3. Definition query
	targetFile, line, col, found := client.Definition(ctx, "/workspace/main.rs", 5, 2)
	if !found {
		t.Fatalf("expected definition to be found")
	}
	if targetFile != "/workspace/target.rs" || line != 16 || col != 5 {
		t.Errorf("expected (/workspace/target.rs, 16, 5), got (%s, %d, %d)", targetFile, line, col)
	}

	// 4. Hover query
	hoverSnippet, found := client.Hover(ctx, "/workspace/main.rs", 5, 2)
	if !found {
		t.Fatalf("expected hover to be found")
	}
	if hoverSnippet != "pub fn calculate(val: i32) -> i32" {
		t.Errorf("expected 'pub fn calculate(val: i32) -> i32', got %q", hoverSnippet)
	}

	// 5. Clean teardown
	client.isClosed.Store(true)
	_ = clientInW.Close()
	_ = clientOutR.Close()
	wg.Wait()
}
