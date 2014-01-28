package spdy3

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"unsafe"
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
//  +----------------------------------+
//  |C| Version(15bits) | Type(16bits) |
//  +----------------------------------+
type HeaderWord uint32

func NewHeaderWord(control bool, version SpdyVersion, typ FrameType) HeaderWord {
	var header HeaderWord

	if control {
		header |= 0x80000000
	}

	header |= HeaderWord(uint32(version&0x7f) << 16)
	header |= HeaderWord(uint32(typ) & 0xff)
	return header
}

// Control bit: The 'C' bit is a single bit indicating if this is a control
// message. For control frames this value is always 1.
func (h HeaderWord) Control() bool {
	return h&0x80000000 == 0x80000000
}

// Version: The version number of the SPDY protocol. This document describes
// SPDY version 3.
func (h HeaderWord) Version() uint16 {
	return uint16((h >> 16) & 0x7F)
}

// Type: The type of control frame. See Control Frames (Section 2.6) for the
// complete list of control frames.
func (h HeaderWord) Type() FrameType {
	return FrameType(h & 0xFFFF)
}

func (h HeaderWord) Write(w io.Writer) (int, error) {
	return writeWord(w, uint32(h))
}

// ----------------------------------------------------------------------------
// Flag/Len Word
//  +----------------------------------+
//  | Flags (8)  |  Length (24 bits)   |
//  +----------------------------------+

type FlagLenWord uint32

func NewFlagLenWord(flags uint8, length uint32) FlagLenWord {
	var flagLenWord FlagLenWord
	flagLenWord |= FlagLenWord(flags << 24)
	flagLenWord |= FlagLenWord(length & 0x00ffffff)
	return flagLenWord
}

func (f FlagLenWord) Flags() uint8 {
	return uint8(f >> 24)
}

func (f FlagLenWord) Length() uint32 {
	return uint32(f & 0x00FFFFFF)
}

func (f FlagLenWord) Write(w io.Writer) (int, error) {
	return writeWord(w, uint32(f))
}

// ----------------------------------------------------------------------------
// StreamID Word
//  +------------------------------------+
//  |X|           Stream-ID (31bits)     |
//  +------------------------------------+
type StreamIdWord uint32

func (s *StreamIdWord) Read(r io.Reader) (int, error) {
	if err := binary.Read(r, binary.BigEndian, s); err != nil {
		return 0, err
	}
	return 4, nil
}

func (s StreamIdWord) StreamId() uint32 {
	return uint32(s & 0x7FFFFFFF)
}

func (s StreamIdWord) Write(w io.Writer) (int, error) {
	return writeWord(w, uint32(s))
}

// ----------------------------------------------------------------------------
// Priority Word
// As a note, this current doesn't grok "slot" because the credentials bit of
// things is pending.
//
//  +------------------------------------+
//  | Pri|Unused | Slot |                |
//  +-------------------+----------------+
type PriorityWord uint32

func (p *PriorityWord) Read(r io.Reader) (int, error) {
	if err := binary.Read(r, binary.BigEndian, p); err != nil {
		return 0, err
	}
	return 4, nil
}

func (p PriorityWord) Priority() uint8 {
	return uint8(p >> 29)
}

func (p PriorityWord) Write(w io.Writer) (int, error) {
	return writeWord(w, uint32(p))
}

// ----------------------------------------------------------------------------
// Helper Types
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// Name/Value Pairs
//
// Each header name must have at least one value. Header names are encoded using
// the US-ASCII character set [ASCII] and must be all lower case. The length of
// each name must be greater than zero. A recipient of a zero-length name MUST
// issue a stream error (Section 2.4.2) with the status code PROTOCOL_ERROR for
// the stream-id.
//
// Duplicate header names are not allowed. To send two identically named
// headers, send a header with two values, where the values are separated by a
// single NUL (0) byte. A header value can either be empty (e.g. the length is
// zero) or it can contain multiple, NUL-separated values, each with length
// greater than zero. The value never starts nor ends with a NUL character.
// Recipients of illegal value fields MUST issue a stream error (Section 2.4.2)
// with the status code PROTOCOL_ERROR for the stream-id.
//
//  +------------------------------------+
//  | Number of Name/Value pairs (int32) |   <+
//  +------------------------------------+    |
//  |     Length of name (int32)         |    | This section is the "Name/Value
//  +------------------------------------+    | Header Block", and is compressed.
//  |           Name (string)            |    |
//  +------------------------------------+    |
//  |     Length of value  (int32)       |    |
//  +------------------------------------+    |
//  |          Value   (string)          |    |
//  +------------------------------------+    |
//  |           (repeats)                |   <+
type NameValuePairs map[string]string

func (nvp NameValuePairs) Read(r io.Reader) (n int, err error) {
	var numPairs, length uint32
	var name, value []byte

	if err = binary.Read(r, binary.BigEndian, &numPairs); err != nil {
		return
	}
	n += 4

	for i := uint32(0); i < numPairs; i++ {
		if err = binary.Read(r, binary.BigEndian, &length); err != nil {
			return
		}
		n += 4 + int(length)

		name = make([]byte, length)
		if _, err = r.Read(name); err != nil {
			return
		}

		if err = binary.Read(r, binary.BigEndian, &length); err != nil {
			return
		}
		n += 4 + int(length)

		value = make([]byte, length)
		if _, err = r.Read(value); err != nil {
			return
		}

		nvp[string(name)] = string(value)
	}

	return
}

func (nvp *NameValuePairs) Write(w io.Writer) (n int, err error) {
	var i int

	if err = binary.Write(w, binary.BigEndian, uint32(len(*nvp))); err != nil {
		return
	}
	n += 4

	for name, value := range *nvp {
		if i, err = nvp.writeString(w, []byte(name)); err != nil {
			return
		}
		n += i
		if i, err = nvp.writeString(w, []byte(value)); err != nil {
			return
		}
		n += i
	}

	return
}

func (nvp *NameValuePairs) writeString(w io.Writer, s []byte) (n int, err error) {
	var i int
	var length = uint32(len(s))

	if err = binary.Write(w, binary.BigEndian, length); err != nil {
		return
	}
	n += 4

	if i, err = w.Write(s); err != nil {
		return
	}
	n += int(i)
	return
}

// ----------------------------------------------------------------------------
// Compressed Name/Value Pairs
//
// Name/Value pairs are compressed by default. This type referrs to the
// compressed value
type CompressedNameValuePairs []byte

// ----------------------------------------------------------------------------
// SYN_STREAM
//
// The SYN_STREAM control frame allows the sender to asynchronously create a
// stream between the endpoints. See Stream Creation (Section 2.3.2)
//
//  +------------------------------------+
//  |1|    version    |         1        |
//  +------------------------------------+
//  |  Flags (8)  |  Length (24 bits)    |
//  +------------------------------------+
//  |X|           Stream-ID (31bits)     |
//  +------------------------------------+
//  |X| Associated-To-Stream-ID (31bits) |
//  +------------------------------------+
//  | Pri|Unused | Slot |                |
//  +-------------------+                |
//  | Number of Name/Value pairs (int32) |   <+
//  +------------------------------------+    |
//  |     Length of name (int32)         |    | This section is the "Name/Value
//  +------------------------------------+    | Header Block", and is compressed.
//  |           Name (string)            |    |
//  +------------------------------------+    |
//  |     Length of value  (int32)       |    |
//  +------------------------------------+    |
//  |          Value   (string)          |    |
//  +------------------------------------+    |
//  |           (repeats)                |   <+
type SynStream struct {
	StreamId           uint32
	AssociatedStreamId uint32
	Priority           uint8
	CompressedHeaders  CompressedNameValuePairs
}

type synStreamFramev3 struct {
	StreamId           StreamIdWord
	AssociatedStreamId StreamIdWord
	Priority           PriorityWord
}

func (s *SynStream) Read(r io.Reader) (n int, err error) {
	frame := new(synStreamFramev3)
	if err = binary.Read(r, binary.BigEndian, frame); err != nil {
		return
	}
	n += int(unsafe.Sizeof(frame))

	s.StreamId = frame.StreamId.StreamId()
	s.AssociatedStreamId = frame.AssociatedStreamId.StreamId()
	s.Priority = frame.Priority.Priority()
	s.CompressedHeaders, err = ioutil.ReadAll(r)
	return
}

func (s *SynStream) Type() FrameType {
	return SynStreamType
}

// Ensure SynStream is a frame
var _ Frame = &SynStream{}

// ----------------------------------------------------------------------------
// SYN_REPLY
//
// SYN_REPLY indicates the acceptance of a stream creation by the recipient of a
// SYN_STREAM frame.
//
//  +------------------------------------+
//  |1|    version    |         2        |
//  +------------------------------------+
//  |  Flags (8)  |  Length (24 bits)    |
//  +------------------------------------+
//  |X|           Stream-ID (31bits)     |
//  +------------------------------------+
//  | Number of Name/Value pairs (int32) |   <+
//  +------------------------------------+    |
//  |     Length of name (int32)         |    | This section is the "Name/Value
//  +------------------------------------+    | Header Block", and is compressed.
//  |           Name (string)            |    |
//  +------------------------------------+    |
//  |     Length of value  (int32)       |    |
//  +------------------------------------+    |
//  |          Value   (string)          |    |
//  +------------------------------------+    |
//  |           (repeats)                |   <+
type SynReply struct {
	StreamId          uint32
	CompressedHeaders CompressedNameValuePairs
}

type synReplyFramev3 struct {
	StreamId StreamIdWord
}

func (s SynReply) Type() FrameType {
	return SynReplyType
}

func (s *SynReply) Read(r io.Reader) (n int, err error) {
	frame := new(synReplyFramev3)
	if err = binary.Read(r, binary.BigEndian, frame); err != nil {
		return
	}
	n += int(unsafe.Sizeof(frame))

	s.StreamId = frame.StreamId.StreamId()
	s.CompressedHeaders, err = ioutil.ReadAll(r)
	return
}

// Ensure SynReply is a frame
var _ Frame = &SynReply{}

// ----------------------------------------------------------------------------
// RST_STREAM
//
// The RST_STREAM frame allows for abnormal termination of a stream. When sent
// by the creator of a stream, it indicates the creator wishes to cancel the
// stream. When sent by the recipient of a stream, it indicates an error or that
// the recipient did not want to accept the stream, so the stream should be
// closed.
//
//  +----------------------------------+
//  |1|   version    |         3       |
//  +----------------------------------+
//  | Flags (8)  |         8           |
//  +----------------------------------+
//  |X|          Stream-ID (31bits)    |
//  +----------------------------------+
//  |          Status code             |
//  +----------------------------------+
type RstStream struct {
	StreamId   uint32
	StatusCode int32
}

type rstStreamFramev3 struct {
	StreamId   StreamIdWord
	StatusCode int32
}

func (rst RstStream) Type() FrameType {
	return RstStreamType
}

func (rst *RstStream) Read(r io.Reader) (n int, err error) {
	frame := new(rstStreamFramev3)
	if err = binary.Read(r, binary.BigEndian, frame); err != nil {
		return
	}
	n += int(unsafe.Sizeof(frame))

	rst.StreamId = frame.StreamId.StreamId()
	rst.StatusCode = frame.StatusCode
	return
}

// Ensure RstStream is a frame
var _ Frame = &RstStream{}

// ----------------------------------------------------------------------------
// SETTINGS
//
// A SETTINGS frame contains a set of id/value pairs for communicating
// configuration data about how the two endpoints may communicate. SETTINGS
// frames can be sent at any time by either endpoint, are optionally sent, and
// are fully asynchronous. When the server is the sender, the sender can request
// that configuration data be persisted by the client across SPDY sessions and
// returned to the server in future communications.
//
//  +----------------------------------+
//  |1|   version    |         4       |
//  +----------------------------------+
//  | Flags (8)  |  Length (24 bits)   |
//  +----------------------------------+
//  |         Number of entries        |
//  +----------------------------------+
//  |          ID/Value Pairs          |
//  |             ...                  |
type Settings struct {
	Settings []*Setting
}

type Setting struct {
	Flags uint8
	Id    uint32
	Value int32
}

type settingv3 struct {
	FlagId FlagLenWord
	Value  int32
}

func (s Settings) Type() FrameType {
	return SettingsType
}

func (s *Settings) Read(r io.Reader) (n int, err error) {
	var numSettings uint32
	if err = binary.Read(r, binary.BigEndian, &numSettings); err != nil {
		return
	}
	n += 4

	s.Settings = make([]*Setting, numSettings)
	for i := uint32(0); i < numSettings; i++ {
		setting := new(settingv3)
		if err = binary.Read(r, binary.BigEndian, setting); err != nil {
			return
		}
		n += 8
		s.Settings[i] = &Setting{
			Flags: setting.FlagId.Flags(),
			Id:    setting.FlagId.Length(),
			Value: setting.Value,
		}
	}
	return
}

// Ensure Settings is a frame
var _ Frame = &Settings{}

// ----------------------------------------------------------------------------
// PING
//
// The PING control frame is a mechanism for measuring a minimal round-trip time
// from the sender. It can be sent from the client or the server. Recipients of
// a PING frame should send an identical frame to the sender as soon as possible
// (if there is other pending data waiting to be sent, PING should take highest
// priority). Each ping sent by a sender should use a unique ID.
//
//  +----------------------------------+
//  |1|   version    |         6       |
//  +----------------------------------+
//  | 0 (flags) |     4 (length)       |
//  +----------------------------------|
//  |            32-bit ID             |
//  +----------------------------------+

type Ping struct {
	Id uint32
}

func (p Ping) Type() FrameType {
	return PingType
}

func (p *Ping) Read(r io.Reader) (n int, err error) {
	if err = binary.Read(r, binary.BigEndian, p); err != nil {
		return
	}
	n += 4
	return
}

// Ensure Ping is a frame
var _ Frame = &Ping{}

// ----------------------------------------------------------------------------
// GOAWAY
//
// The GOAWAY control frame is a mechanism to tell the remote side of the
// connection to stop creating streams on this session. It can be sent from the
// client or the server. Once sent, the sender will not respond to any new
// SYN_STREAMs on this session. Recipients of a GOAWAY frame must not send
// additional streams on this session, although a new session can be established
// for new streams. The purpose of this message is to allow an endpoint to
// gracefully stop accepting new streams (perhaps for a reboot or maintenance),
// while still finishing processing of previously established streams.
//
//  +----------------------------------+
//  |1|   version    |         7       |
//  +----------------------------------+
//  | 0 (flags) |     8 (length)       |
//  +----------------------------------|
//  |X|  Last-good-stream-ID (31 bits) |
//  +----------------------------------+
//  |          Status code             |
//  +----------------------------------+

type GoAway struct {
	LastGoodStreamId uint32
	StatusCode       uint32
}

type goAwayFramev3 struct {
	LastGoodStreamId StreamIdWord
	StatusCode       uint32
}

func (g GoAway) Type() FrameType {
	return GoAwayType
}

func (g *GoAway) Read(r io.Reader) (n int, err error) {
	frame := new(goAwayFramev3)
	if err = binary.Read(r, binary.BigEndian, frame); err != nil {
		return
	}
	n += int(unsafe.Sizeof(frame))

	g.LastGoodStreamId = frame.LastGoodStreamId.StreamId()
	g.StatusCode = frame.StatusCode
	return
}

// Ensure GoAway is a frame
var _ Frame = &GoAway{}

// ----------------------------------------------------------------------------
// HEADERS
//
// The HEADERS frame augments a stream with additional headers. It may be
// optionally sent on an existing stream at any time. Specific application of
// the headers in this frame is application-dependent. The name/value header
// block within this frame is compressed.
//
//  +------------------------------------+
//  |1|   version     |          8       |
//  +------------------------------------+
//  | Flags (8)  |   Length (24 bits)    |
//  +------------------------------------+
//  |X|          Stream-ID (31bits)      |
//  +------------------------------------+
//  | Number of Name/Value pairs (int32) |   <+
//  +------------------------------------+    |
//  |     Length of name (int32)         |    | This section is the "Name/Value
//  +------------------------------------+    | Header Block", and is compressed.
//  |           Name (string)            |    |
//  +------------------------------------+    |
//  |     Length of value  (int32)       |    |
//  +------------------------------------+    |
//  |          Value   (string)          |    |
//  +------------------------------------+    |
//  |           (repeats)                |   <+
type Headers struct {
	StreamId          uint32
	CompressedHeaders CompressedNameValuePairs
}

type headersFramev3 struct {
	StreamId StreamIdWord
}

func (h Headers) Type() FrameType {
	return HeadersType
}

func (h *Headers) Read(r io.Reader) (n int, err error) {
	frame := new(headersFramev3)
	if err = binary.Read(r, binary.BigEndian, frame); err != nil {
		return
	}
	n += int(unsafe.Sizeof(frame))

	h.StreamId = frame.StreamId.StreamId()
	h.CompressedHeaders, err = ioutil.ReadAll(r)
	return
}

// Ensure Headers is a frame
var _ Frame = &Headers{}

// ----------------------------------------------------------------------------
// WINDOW_UPDATE
//
// The WINDOW_UPDATE control frame is used to implement per stream flow control
// in SPDY. Flow control in SPDY is per hop, that is, only between the two
// endpoints of a SPDY connection. If there are one or more intermediaries
// between the client and the origin server, flow control signals are not
// explicitly forwarded by the intermediaries. (However, throttling of data
// transfer by any recipient may have the effect of indirectly propagating flow
// control information upstream back to the original sender.) Flow control only
// applies to the data portion of data frames. Recipients must buffer all
// control frames. If a recipient fails to buffer an entire control frame, it
// MUST issue a stream error (Section 2.4.2) with the status code
// FLOW_CONTROL_ERROR for the stream.
//
//  +----------------------------------+
//  |1|   version    |         9       |
//  +----------------------------------+
//  | 0 (flags) |     8 (length)       |
//  +----------------------------------+
//  |X|     Stream-ID (31-bits)        |
//  +----------------------------------+
//  |X|  Delta-Window-Size (31-bits)   |
//  +----------------------------------+
type WindowUpdate struct {
	StreamId        uint32
	DeltaWindowSize uint32
}

type windowUpdateFramev3 struct {
	StreamId        StreamIdWord
	DeltaWindowSize StreamIdWord
}

func (w WindowUpdate) Type() FrameType {
	return WindowUpdateType
}

func (w *WindowUpdate) Read(r io.Reader) (n int, err error) {
	frame := new(windowUpdateFramev3)
	if err = binary.Read(r, binary.BigEndian, frame); err != nil {
		return
	}
	n += int(unsafe.Sizeof(frame))

	w.StreamId = frame.StreamId.StreamId()
	w.DeltaWindowSize = frame.DeltaWindowSize.StreamId()
	return
}

// Ensure Headers is a frame
var _ Frame = &WindowUpdate{}

// ----------------------------------------------------------------------------
// Helper functions

func writeWord(w io.Writer, word uint32) (int, error) {
	return w.Write([]byte{
		byte(word & 0xff000000 >> 24),
		byte(word & 0x00ff0000 >> 16),
		byte(word & 0x0000ff00 >> 8),
		byte(word & 0x000000ff),
	})
}
