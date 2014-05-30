package spdy3

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Framer", func() {
	rw := new(bytes.Buffer)
	framer := NewFramer(Spdy3, rw)

	Describe("Should EOF when there are no more frames", func() {
		_, err := framer.Read()
		Expect(err).To(Equal(io.EOF))
	})

	Describe("Should read a simple frame", func() {
		NewHeaderWord(true, Spdy3, SynStreamType).Write(rw)
		NewFlagLenWord(0, 0).Write(rw)
		StreamIdWord(666).Write(rw)
		StreamIdWord(0).Write(rw)
		PriorityWord(0).Write(rw)

		_, err := framer.Read()
		Expect(err).To(BeNil())
	})
})
