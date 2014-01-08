package spdy3

import (
	"github.com/markchadwick/spec"
	"io"
)

var _ = spec.Suite("Framer", func(c *spec.C) {
	rw := new(RWBuffer)
	framer := NewFramer(Spdy3, rw)

	c.It("Should EOF when there are no more frames", func(c *spec.C) {
		_, err := framer.Read()
		c.Assert(err).Equals(io.EOF)
	})

	c.It("Should read a simple frame", func(c *spec.C) {
		header := NewHeaderWord(true, Spdy3, SynStreamType)
		header.Write(rw)
	})

	c.It("Should read two frames", func(c *spec.C) {
	})
})
