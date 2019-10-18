// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// package lsp contains language server protocol logic.
package lsp

import (
	"context"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/nelsam/vidar/commander/input"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// ConnOpt is an option function used to tell New how to connect to
// the language server.
type ConnOpt func(Conn) (Conn, error)

//  WithAddr returns a ConnOpt that will tell New to connect to addr.
func WithAddr(network, addr string) ConnOpt {
	return func(c Conn) (Conn, error) {
		conn, err := net.Dial(network, addr)
		if err != nil {
			return Conn{}, errors.Wrapf(err, "could not connect to language server at %s", addr)
		}
		return WithConn(conn)(c)
	}
}

// WithConn returns a ConnOpt that will tell New to use conn as the
// language server connection.
func WithConn(conn io.ReadWriteCloser) ConnOpt {
	return func(c Conn) (Conn, error) {
		c.rpc = jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), nil)
		return c, nil
	}
}

// WithCommand returns a ConnOpt that will spawn a command and use
// the process's STDOUT/STDIN as the connection's Read and Write
// methods, respectively.
func WithCommand(cmd *exec.Cmd, env ...string) ConnOpt {
	cmd.Env = env
	cmd.Stderr = os.Stderr // TODO: use UI elements.
	toCmd := &buffPipe{pipe: make(chan []byte, 3)}
	fromCmd := &buffPipe{pipe: make(chan []byte, 3)}
	golsconn := stdioconn{
		out: toCmd,
		in:  fromCmd,
		close: func() error {
			return cmd.Process.Kill()
		},
	}
	cmd.Stdin = toCmd
	cmd.Stdout = fromCmd
	return func(c Conn, err error) {
		if err := cmd.Start(); err != nil {
			return Conn{}, err
		}
		return WithConn(golsconn)(c)
	}
}

// Conn is a connection to a language server.
type Conn struct {
	rpc *jsonrpc2.Conn

	curr input.Editor
}

// New returns a new Conn.
func New(opts ...ConnOpt) (*Conn, error) {
	var (
		c   Conn
		err error
	)
	for _, o := range opts {
		c, err = o(c)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Close is used to close the language server connection.
func (c *Conn) Close() error {
	return c.rpc.Close()
}

// Initialize runs the 'initialize' rpc.
func (c *Conn) Initialize(initopts interface{}) (res lsp.InitializeResult, err error) {
	p := lsp.InitializeParams{
		ProcessID:             os.Getpid(),
		InitializationOptions: initopts,
	}
	if err := c.rpc.Call(context.Background(), "initialize", p, &res); err != nil {
		return res, err
	}
	return res, nil
}

// Completion runs the 'completion' rpc.
func (c *Conn) Completion(e input.Editor) (res lsp.CompletionList, err error) {
	p := lsp.CompletionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: lsp.DocumentURI("file://" + c.currFile)},
	}
}
