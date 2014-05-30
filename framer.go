package spdy3

import (
	"encoding/binary"
	"errors"
	"io"
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

	if header.Control() {
		return f.readControlFrame()
	}

	return nil, errors.New("Can only read ctrl frames")
}

func (f *Framer) readControlFrame() (fr Frame, err error) {
	flagLen := new(FlagLenWord)
	if err = binary.Read(f.rw, binary.BigEndian, flagLen); err != nil {
		return
	}
	bs := make([]byte, flagLen.Length())
	if _, err = io.ReadFull(f.rw, bs); err != nil {
		return
	}
	return nil, nil
}

func (f *Framer) Write(fr Frame) (err error) {
	return nil
}
