package spdy3

import (
	"encoding/binary"
	"io"
)

type SpdyVersion uint16

const (
	Spdy3 SpdyVersion = 3
)

type Framer struct {
	Version SpdyVersion
	rw      io.ReadWriter
}

func NewFramer(version SpdyVersion, rw io.ReadWriter) *Framer {
	return &Framer{
		Version: version,
		rw:      rw,
	}
}

func (f *Framer) Read() (fr Frame, err error) {
	var header = new(HeaderWord)
	if err = binary.Read(f.rw, binary.BigEndian, header); err != nil {
		return
	}
	return nil, io.EOF
}

func (f *Framer) Write(fr Frame) (err error) {
	return nil
}
