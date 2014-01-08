package spdy3

import (
	"bytes"
	"github.com/markchadwick/spec"
	"io"
	"testing"
)

func Test(t *testing.T) {
	spec.Run(t)
}

type RWBuffer struct {
	buf bytes.Buffer
}

func (rw *RWBuffer) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (rw *RWBuffer) Write(p []byte) (n int, err error) {
	return 0, io.EOF
}
