package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrServerNotFound = errors.New("lsp server binary not found")
	ErrClientClosed   = errors.New("lsp client is closed")
	ErrTimeout        = errors.New("lsp request timed out")
)

// FindRustAnalyzer locates the rust-analyzer binary across standard rustup and cargo bin paths.
func FindRustAnalyzer() (string, bool) {
	candidates := []string{"rust-analyzer"}
	if home := os.Getenv("HOME"); home != "" {
		candidates = append(candidates, filepath.Join(home, ".cargo", "bin", "rust-analyzer"))
	}
	candidates = append(candidates,
		"/usr/local/bin/rust-analyzer",
		"/opt/homebrew/bin/rust-analyzer",
	)

	for _, cand := range candidates {
		if path, err := exec.LookPath(cand); err == nil {
			return path, true
		}
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand, true
		}
	}
	return "", false
}

// Client implements a lightweight, pure-Go JSON-RPC 2.0 LSP client over stdio.
type Client struct {
	serverName string
	binPath    string
	rootDir    string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	reader     *bufio.Reader

	muWrite   sync.Mutex
	muPending sync.Mutex
	pending   map[int64]chan *rpcResponse
	seq       atomic.Int64
	docVer    sync.Map // uri (string) -> *atomic.Int64

	isReady  atomic.Bool
	isClosed atomic.Bool
	closeWg  sync.WaitGroup
}

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type rpcNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StartRustAnalyzerClient attempts to locate and launch rust-analyzer for the workspace.
func StartRustAnalyzerClient(rootDir string) (*Client, error) {
	bin, found := FindRustAnalyzer()
	if !found {
		return nil, ErrServerNotFound
	}
	return StartClient("rust-analyzer", bin, rootDir)
}

// StartClient spawns an LSP process and starts the stdio message loop.
func StartClient(serverName, binPath, rootDir string) (*Client, error) {
	cmd := exec.Command(binPath)

	cmd.Dir = rootDir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, err
	}

	c := &Client{
		serverName: serverName,
		binPath:    binPath,
		rootDir:    rootDir,
		cmd:        cmd,
		stdin:      stdin,
		stdout:     stdout,
		reader:     bufio.NewReader(stdout),
		pending:    make(map[int64]chan *rpcResponse),
	}

	c.closeWg.Add(1)
	go c.readLoop()

	// Initialize handshake with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.initialize(ctx); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("lsp initialize failed: %w", err)
	}

	c.isReady.Store(true)
	return c, nil
}

func (c *Client) IsAvailable() bool {
	return c != nil && c.isReady.Load() && !c.isClosed.Load()
}

func (c *Client) ServerName() string {
	if c == nil {
		return "None"
	}
	return c.serverName
}

func (c *Client) BinPath() string {
	if c == nil {
		return ""
	}
	return c.binPath
}

func (c *Client) RootDir() string {
	if c == nil {
		return ""
	}
	return c.rootDir
}

func (c *Client) readLoop() {
	defer c.closeWg.Done()
	for {
		if c.isClosed.Load() {
			return
		}

		contentLength := -1
		// 1. Read headers until blank line (\r\n\r\n or \n\n)
		for {
			line, err := c.reader.ReadString('\n')
			if err != nil {
				c.failAllPending(err)
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			if strings.HasPrefix(strings.ToLower(line), "content-length:") {
				val := strings.TrimSpace(line[len("content-length:"):])
				if n, err := strconv.Atoi(val); err == nil {
					contentLength = n
				}
			}
		}

		if contentLength <= 0 {
			continue
		}

		// 2. Read exact payload bytes
		body := make([]byte, contentLength)
		if _, err := io.ReadFull(c.reader, body); err != nil {
			c.failAllPending(err)
			return
		}

		// 3. Parse JSON response or notification
		var resp rpcResponse
		if err := json.Unmarshal(body, &resp); err == nil && resp.ID != 0 {
			c.muPending.Lock()
			ch, ok := c.pending[resp.ID]
			if ok {
				delete(c.pending, resp.ID)
			}
			c.muPending.Unlock()

			if ok && ch != nil {
				ch <- &resp
			}
		}
	}
}

func (c *Client) failAllPending(err error) {
	c.muPending.Lock()
	defer c.muPending.Unlock()
	for id, ch := range c.pending {
		delete(c.pending, id)
		close(ch)
	}
}

func (c *Client) writeFrame(payload []byte) error {
	c.muWrite.Lock()
	defer c.muWrite.Unlock()

	if c.isClosed.Load() {
		return ErrClientClosed
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	if _, err := io.WriteString(c.stdin, header); err != nil {
		return err
	}
	if _, err := c.stdin.Write(payload); err != nil {
		return err
	}
	return nil
}

func (c *Client) call(ctx context.Context, method string, params interface{}) (*rpcResponse, error) {
	if !c.isReady.Load() && method != "initialize" {
		return nil, ErrClientClosed
	}

	id := c.seq.Add(1)
	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	respCh := make(chan *rpcResponse, 1)
	c.muPending.Lock()
	c.pending[id] = respCh
	c.muPending.Unlock()

	if err := c.writeFrame(payload); err != nil {
		c.muPending.Lock()
		delete(c.pending, id)
		c.muPending.Unlock()
		return nil, err
	}

	select {
	case <-ctx.Done():
		c.muPending.Lock()
		delete(c.pending, id)
		c.muPending.Unlock()
		return nil, ErrTimeout
	case resp, ok := <-respCh:
		if !ok || resp == nil {
			return nil, errors.New("lsp server connection closed")
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("lsp error (%d): %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	}
}

func (c *Client) notify(method string, params interface{}) error {
	req := rpcNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.writeFrame(payload)
}

func (c *Client) initialize(ctx context.Context) error {
	rootURI := PathToURI(c.rootDir)
	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"definition": map[string]interface{}{
					"linkSupport": true,
				},
				"hover": map[string]interface{}{
					"contentFormat": []string{"plaintext", "markdown"},
				},
				"synchronization": map[string]interface{}{
					"openClose": true,
					"change":    1, // Full document sync (1)
				},
			},
		},
	}

	_, err := c.call(ctx, "initialize", params)
	if err != nil {
		return err
	}

	// Send initialized notification
	_ = c.notify("initialized", map[string]interface{}{})
	return nil
}

// DidOpen notifies the server that a text document was opened.
func (c *Client) DidOpen(filePath, content string) error {
	if !c.IsAvailable() {
		return nil
	}
	uri := PathToURI(filePath)
	lang := "rust"

	verPtr := &atomic.Int64{}
	verPtr.Store(1)
	c.docVer.Store(uri, verPtr)

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": lang,
			"version":    1,
			"text":       content,
		},
	}
	return c.notify("textDocument/didOpen", params)
}

// DidChange notifies the server of full content replacement.
func (c *Client) DidChange(filePath, content string) error {
	if !c.IsAvailable() {
		return nil
	}
	uri := PathToURI(filePath)
	val, ok := c.docVer.Load(uri)
	var ver int64 = 1
	if ok {
		ver = val.(*atomic.Int64).Add(1)
	} else {
		verPtr := &atomic.Int64{}
		verPtr.Store(1)
		c.docVer.Store(uri, verPtr)
	}

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     uri,
			"version": ver,
		},
		"contentChanges": []map[string]interface{}{
			{"text": content},
		},
	}
	return c.notify("textDocument/didChange", params)
}

// DidClose notifies the server that a document was closed.
func (c *Client) DidClose(filePath string) error {
	if !c.IsAvailable() {
		return nil
	}
	uri := PathToURI(filePath)
	c.docVer.Delete(uri)

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}
	return c.notify("textDocument/didClose", params)
}

// Definition queries definition location (targetFile, 1-based targetLine, 1-based targetCol).
func (c *Client) Definition(ctx context.Context, filePath string, line0, col0 int) (string, int, int, bool) {
	if !c.IsAvailable() {
		return "", 0, 0, false
	}

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": PathToURI(filePath),
		},
		"position": map[string]interface{}{
			"line":      line0,
			"character": col0,
		},
	}

	resp, err := c.call(ctx, "textDocument/definition", params)
	if err != nil || resp == nil || len(resp.Result) == 0 || string(resp.Result) == "null" {
		return "", 0, 0, false
	}

	// Result can be Location, []Location, or []LocationLink
	type pos struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	}
	type lRange struct {
		Start pos `json:"start"`
		End   pos `json:"end"`
	}
	type loc struct {
		URI         string `json:"uri"`
		TargetURI   string `json:"targetUri"`
		Range       lRange `json:"range"`
		TargetRange lRange `json:"targetRange"`
	}

	// 1. Try single Location
	var single loc
	if err := json.Unmarshal(resp.Result, &single); err == nil && (single.URI != "" || single.TargetURI != "") {
		targetURI := single.URI
		if targetURI == "" {
			targetURI = single.TargetURI
		}
		targetRange := single.Range
		if single.TargetURI != "" {
			targetRange = single.TargetRange
		}
		return URIToPath(targetURI), targetRange.Start.Line + 1, targetRange.Start.Character + 1, true
	}

	// 2. Try slice of Locations
	var multi []loc
	if err := json.Unmarshal(resp.Result, &multi); err == nil && len(multi) > 0 {
		targetURI := multi[0].URI
		if targetURI == "" {
			targetURI = multi[0].TargetURI
		}
		targetRange := multi[0].Range
		if multi[0].TargetURI != "" {
			targetRange = multi[0].TargetRange
		}
		return URIToPath(targetURI), targetRange.Start.Line + 1, targetRange.Start.Character + 1, true
	}

	return "", 0, 0, false
}

// Hover queries hover documentation/type signature for position (0-based line, 0-based col).
func (c *Client) Hover(ctx context.Context, filePath string, line0, col0 int) (string, bool) {
	if !c.IsAvailable() {
		return "", false
	}

	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": PathToURI(filePath),
		},
		"position": map[string]interface{}{
			"line":      line0,
			"character": col0,
		},
	}

	resp, err := c.call(ctx, "textDocument/hover", params)
	if err != nil || resp == nil || len(resp.Result) == 0 || string(resp.Result) == "null" {
		return "", false
	}

	var res struct {
		Contents json.RawMessage `json:"contents"`
	}
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		return "", false
	}

	// Contents can be: string, MarkupContent {"kind": "...", "value": "..."}, or slice
	var markup struct {
		Kind  string `json:"kind"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(res.Contents, &markup); err == nil && markup.Value != "" {
		return cleanHoverSnippet(markup.Value), true
	}

	var strVal string
	if err := json.Unmarshal(res.Contents, &strVal); err == nil && strVal != "" {
		return cleanHoverSnippet(strVal), true
	}

	return "", false
}

func cleanHoverSnippet(s string) string {
	s = strings.TrimSpace(s)
	// Strip markdown code fence markers (e.g. ```rust ... ```)
	lines := strings.Split(s, "\n")
	var clean []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}
		if trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		return s
	}
	// Return first non-empty representative signature line
	return clean[0]
}

// Close gracefully terminates the server process.
func (c *Client) Close() error {
	if c == nil || c.isClosed.Swap(true) {
		return nil
	}

	c.isReady.Store(false)

	// Send shutdown request with tight 500ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, _ = c.call(ctx, "shutdown", nil)
	_ = c.notify("exit", nil)

	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.stdout != nil {
		_ = c.stdout.Close()
	}

	// Terminate process with 1.5s escalation (Rule 4)
	if c.cmd != nil {
		done := make(chan struct{})
		go func() {
			_ = c.cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(1500 * time.Millisecond):
			if c.cmd.Process != nil {
				_ = c.cmd.Process.Kill()
			}
		}
	}

	c.failAllPending(ErrClientClosed)
	return nil
}

// PathToURI converts a local filesystem path to a file:// URI.
func PathToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return "file://" + filepath.ToSlash(abs)
}

// URIToPath converts a file:// URI back to a local filesystem path.
func URIToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		u, err := url.Parse(uri)
		if err == nil {
			return filepath.Clean(u.Path)
		}
		return strings.TrimPrefix(uri, "file://")
	}
	return uri
}
