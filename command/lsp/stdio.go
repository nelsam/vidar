// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package lsp

import "io"

type buffPipe struct {
	pipe   chan []byte
	cached []byte
}

func (p *buffPipe) Write(b []byte) (n int, err error) {
	cb := make([]byte, len(b))
	copy(cb, b)
	p.pipe <- cb
	return len(b), nil
}

func (p *buffPipe) Read(b []byte) (n int, err error) {
	if len(p.cached) == 0 {
		p.cached = <-p.pipe
	}
	copy(b, p.cached)
	if len(p.cached) > len(b) {
		p.cached = p.cached[len(b)+1:]
		return len(b), nil
	}
	n = len(p.cached)
	p.cached = nil
	return n, nil
}

type stdioconn struct {
	out   io.Writer
	in    io.Reader
	close func() error
}

func (c *stdoutinconn) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *stdoutinconn) Write(p []byte) (n int, err error) {
	return c.out.Write(p)
}

func (c *stdoutinconn) Close() error { return c.close() }
