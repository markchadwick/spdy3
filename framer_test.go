package spdy3

import (
	"bytes"
	"github.com/markchadwick/spec"
	"io"
)

var _ = spec.Suite("Framer", func(c *spec.C) {
	rw := new(bytes.Buffer)
	framer := NewFramer(Spdy3, rw)

	c.It("Should EOF when there are no more frames", func(c *spec.C) {
		_, err := framer.Read()
		c.Assert(err).Equals(io.EOF)
	})

	c.It("Should read a simple frame", func(c *spec.C) {
		NewHeaderWord(true, Spdy3, SynStreamType).Write(rw)
		NewFlagLenWord(0, 0).Write(rw)
		StreamIdWord(666).Write(rw)
		StreamIdWord(0).Write(rw)
		PriorityWord(0).Write(rw)

		_, err := framer.Read()
		c.Assert(err).IsNil()
		// c.Assert(frame).NotNil()
	})

	c.It("Should read two frames", func(c *spec.C) {
	})
})
