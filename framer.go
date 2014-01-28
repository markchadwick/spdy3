package spdy3

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
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
	log.Printf("--------------------------")
	log.Printf("Reading frame")
	var header = new(HeaderWord)
	if err = binary.Read(f.rw, binary.BigEndian, header); err != nil {
		return
	}
	log.Printf("# Header")
	log.Printf("    Control:  %v", header.Control())
	log.Printf("    Version:  %v", header.Version())
	log.Printf("    Type:     %v", header.Type())
	log.Printf("--------------------------")

	if header.Control() {
		return f.readControlFrame()
	}

	return nil, errors.New("Can only read ctrl frames")
}

func (f *Framer) readControlFrame() (fr Frame, err error) {
	log.Printf("Reading control frame")
	flagLen := new(FlagLenWord)
	if err = binary.Read(f.rw, binary.BigEndian, flagLen); err != nil {
		return
	}
	log.Printf("# Flag Len")
	log.Printf("    Flags:  %v", flagLen.Flags())
	log.Printf("    Length: %d", flagLen.Length())
	bs := make([]byte, flagLen.Length())
	if _, err = io.ReadFull(f.rw, bs); err != nil {
		return
	}
	log.Printf("    bs: %v", bs)
	return nil, nil
}

func (f *Framer) Write(fr Frame) (err error) {
	return nil
}
