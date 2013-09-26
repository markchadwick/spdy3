package spdy3

import (
	"io"
)

type FrameType uint16

const (
	_                       = iota
	SynStreamType FrameType = iota
	SynReplyType
	RstStreamType
	SettingsType
	_
	PingType
	GoAwayType
	HeadersType
	WindowUpdateType
	CredentialType
)

type Frame interface {
	Type() FrameType
}

// ----------------------------------------------------------------------------
// Header Word
// +----------------------------------+
// |C| Version(15bits) | Type(16bits) |
// +----------------------------------+

type HeaderWord uint32

// Control bit: The 'C' bit is a single bit indicating if this is a control
// message. For control frames this value is always 1.
func (h HeaderWord) Control() bool {
	return h&0x80000000 == 0x80000000
}

// Version: The version number of the SPDY protocol. This document describes
// SPDY version 3.
func (h HeaderWord) Version() int {
	return int((h >> 16) & 0x7F)
}

// Type: The type of control frame. See Control Frames (Section 2.6) for the
// complete list of control frames.
func (h HeaderWord) Type() FrameType {
	return FrameType(h & 0xFFFF)
}

// ----------------------------------------------------------------------------
// # Flag/Len Word
// +----------------------------------+
// | Flags (8)  |  Length (24 bits)   |
// +----------------------------------+

type FlagLenWord uint32

func (f FlagLenWord) Flags() uint8 {
	return uint8(f >> 24)
}

func (f FlagLenWord) Length() uint32 {
	return uint32(f & 0x00FFFFFF)
}

// ----------------------------------------------------------------------------
// # StreamID Word
// +------------------------------------+
// |X|           Stream-ID (31bits)     |
// +------------------------------------+

type StreamIdWord uint32

func (f StreamIdWord) StreamId() uint32 {
	return uint32(f & 0x7FFFFFFF)
}

func (f StreamIdWord) Read(r io.Reader) (int, error) {
	return 0, nil
}

func (f StreamIdWord) Write(w io.Reader) (int, error) {
	return 0, nil
}

// ----------------------------------------------------------------------------
// # SYN_STREAM
// The SYN_STREAM control frame allows the sender to asynchronously create a
// stream between the endpoints. See Stream Creation (Section 2.3.2)
//
// +------------------------------------+
// |1|    version    |         1        |
// +------------------------------------+
// |  Flags (8)  |  Length (24 bits)    |
// +------------------------------------+
// |X|           Stream-ID (31bits)     |
// +------------------------------------+
// |X| Associated-To-Stream-ID (31bits) |
// +------------------------------------+
// | Pri|Unused | Slot |                |
// +-------------------+                |
// | Number of Name/Value pairs (int32) |   <+
// +------------------------------------+    |
// |     Length of name (int32)         |    | This section is the "Name/Value
// +------------------------------------+    | Header Block", and is compressed.
// |           Name (string)            |    |
// +------------------------------------+    |
// |     Length of value  (int32)       |    |
// +------------------------------------+    |
// |          Value   (string)          |    |
// +------------------------------------+    |
// |           (repeats)                |   <+

type SynStream struct {
	StreamId           StreamIdWord
	AssociatedStreamId StreamIdWord
}

func (s *SynStream) Read(r io.Reader) (n int, err error) {
	var i int

	if i, err = s.StreamId.Read(r); err != nil {
		return
	}
	n += i

	if i, err = s.AssociatedStreamId.Read(r); err != nil {
		return
	}
	n += i

	return
}

func (s *SynStream) Type() FrameType {
	return SynStreamType
}

// Ensure SynStream is a frame
var _ Frame = &SynStream{}
