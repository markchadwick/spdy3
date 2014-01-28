package spdy3

import (
	"github.com/markchadwick/spec"
	"testing"
)

func Test(t *testing.T) {
	spec.Run(t)
}

/*
type RWBuffer struct {
	buf bytes.Buffer
}

func (rw *RWBuffer) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (rw *RWBuffer) Write(p []byte) (n int, err error) {
	return 0, io.EOF
}
*/
