package spdy3

import (
	"bytes"
	"github.com/markchadwick/spec"
)

var _ = spec.Suite("Header Word", func(c *spec.C) {
	c.It("when a control frame", func(c *spec.C) {
		var header HeaderWord = 0x80030002

		c.It("should be a control frame", func(c *spec.C) {
			c.Assert(header.Control()).IsTrue()
		})

		c.It("should know its version", func(c *spec.C) {
			c.Assert(header.Version()).Equals(3)
		})

		c.It("should know its type", func(c *spec.C) {
			c.Assert(header.Type()).Equals(SynReplyType)
		})
	})

	c.It("when a data frame", func(c *spec.C) {
		var header HeaderWord = 0x000A2C2A
		c.It("should not be a control frame", func(c *spec.C) {
			c.Assert(header.Control()).IsFalse()
		})
	})

	c.It("should err with UNSUPPORTED_VERSION", func(c *spec.C) {
		c.Skip("pending")
	})
})

var _ = spec.Suite("Flag/Len Word", func(c *spec.C) {
	var word FlagLenWord = 0xAB123456

	c.It("should know its flags", func(c *spec.C) {
		c.Assert(word.Flags()).Equals(uint8(171))
	})

	c.It("should know its length", func(c *spec.C) {
		c.Assert(word.Length()).Equals(uint32(1193046))
	})

	c.It("must return a protocol err if len > 2^24", func(c *spec.C) {
		c.Skip("pending")
	})
})

var _ = spec.Suite("StreamId Word", func(c *spec.C) {
	var word StreamIdWord = 0xFFFFFFFF

	c.It("should know its stream ID", func(c *spec.C) {
		// Won't actually be 4294967295 -- first bit is dropped
		c.Assert(word.StreamId()).Equals(uint32(2147483647))
	})

	c.It("Should read from an io.Reader", func(c *spec.C) {
		var streamIdWord StreamIdWord
		r := bytes.NewBuffer([]byte{0x00, 0x00, 0x02, 0x9A})
		n, err := streamIdWord.Read(r)
		c.Assert(n).Equals(4)
		c.Assert(err).IsNil()
		c.Assert(streamIdWord.StreamId()).Equals(uint32(666))
	})
})

var _ = spec.Suite("Priority Word", func(c *spec.C) {
	c.It("Should know its priority", func(c *spec.C) {
		var p PriorityWord = 0x20000000
		c.Assert(p.Priority()).Equals(uint8(1))

		p = 0xA0000000
		c.Assert(p.Priority()).Equals(uint8(5))

		p = 0xE0000000
		c.Assert(p.Priority()).Equals(uint8(7))

		p = 0xFFFFFFFF
		c.Assert(p.Priority()).Equals(uint8(7))
	})
})

var pairs = []byte{
	0x00, 0x00, 0x00, 0x02, // | Number of Name/Value pairs (int32) |
	0x00, 0x00, 0x00, 0x04, // |     Length of name (int32)         |
	0x6e, 0x61, 0x6d, 0x65, // |           Name (string)            |
	0x00, 0x00, 0x00, 0x05, // |     Length of value  (int32)       |
	0x4d, 0x61, 0x72, 0x6b, // |          Value   (string)          |
	0x21,
	0x00, 0x00, 0x00, 0x04, // |     Length of name (int32)         |
	0x6a, 0x6f, 0x62, 0x3f, // |           Name (string)            |
	0x00, 0x00, 0x00, 0x09, // |     Length of value  (int32)       |
	0x6f, 0x68, 0x2c, 0x20, // |          Value   (string)          |
	0x72, 0x69, 0x67, 0x68,
	0x74,
}

var _ = spec.Suite("Name/Value pairs", func(c *spec.C) {
	c.It("should read a basic set of pairs", func(c *spec.C) {
		r := bytes.NewBuffer(pairs)
		nameValuePairs := make(NameValuePairs)
		n, err := nameValuePairs.Read(r)

		c.Assert(n).Equals(42)
		c.Assert(err).IsNil()

		c.Assert(nameValuePairs).HasLen(2)
		name, ok := nameValuePairs["name"]
		c.Assert(ok).IsTrue()
		c.Assert(name).Equals("Mark!")

		job, ok := nameValuePairs["job?"]
		c.Assert(ok).IsTrue()
		c.Assert(job).Equals("oh, right")
	})

	c.It("should write uncompressed", func(c *spec.C) {
		nvp := make(NameValuePairs)

		// Set up exactly like above, may be re-ordered, however.
		nvp["name"] = "Mark!"
		nvp["job?"] = "oh, right"

		buf := new(bytes.Buffer)
		n, err := nvp.Write(buf)
		c.Assert(err).IsNil()
		c.Assert(n).Equals(42)
	})

	c.It("must reject a zero-length name with a stream error", func(c *spec.C) {
		c.Skip("pending")
	})

	c.It("must reject duplicate header names", func(c *spec.C) {
		c.Skip("pending")
	})

	c.It("should accept mutliple header values", func(c *spec.C) {
		c.Skip("pending")
	})
})

var _ = spec.Suite("SYN_STREAM", func(c *spec.C) {
	var synStreamBody = []byte{
		0x00, 0x00, 0x02, 0x9A, // |X|           Stream-ID (31bits)     |
		0x49, 0x96, 0x02, 0xD2, // |X| Associated-To-Stream-ID (31bits) |
		0xA0, 0x00, 0x00, 0x00, // | Pri|Unused | Slot |                |
		0x00, 0x01, 0x02, 0x03, // | Raw headers...                     |
	}
	var r = bytes.NewBuffer(synStreamBody)

	synStream := new(SynStream)
	_, err := synStream.Read(r)
	c.Assert(err).IsNil()

	c.It("should be type 1", func(c *spec.C) {
		c.Assert(SynStreamType).Equals(FrameType(1))
	})

	c.It("should parse its stream ids", func(c *spec.C) {
		c.Assert(synStream.StreamId).Equals(uint32(666))
		c.Assert(synStream.AssociatedStreamId).Equals(uint32(1234567890))
	})

	c.It("should parse its priority", func(c *spec.C) {
		c.Assert(synStream.Priority).Equals(uint8(5))
	})

	c.It("should parse its raw headers", func(c *spec.C) {
		c.Assert(synStream.CompressedHeaders).HasLen(4)
		c.Assert(synStream.CompressedHeaders[0]).Equals(byte(0x00))
		c.Assert(synStream.CompressedHeaders[1]).Equals(byte(0x01))
		c.Assert(synStream.CompressedHeaders[2]).Equals(byte(0x02))
		c.Assert(synStream.CompressedHeaders[3]).Equals(byte(0x03))
	})
})

var _ = spec.Suite("SYN_REPLY", func(c *spec.C) {
	var synReplyBody = []byte{
		0x00, 0x00, 0x02, 0x9A, // |X|           Stream-ID (31bits)     |
		0x00, 0x01, 0x02, 0x03, // | Raw headers...                     |
	}
	var r = bytes.NewBuffer(synReplyBody)
	synReply := new(SynReply)
	_, err := synReply.Read(r)
	c.Assert(err).IsNil()

	c.It("should be type 2", func(c *spec.C) {
		c.Assert(SynReplyType).Equals(FrameType(2))
	})

	c.It("should parse its stream id", func(c *spec.C) {
		c.Assert(synReply.StreamId).Equals(uint32(666))
	})

	c.It("should parse its raw headers", func(c *spec.C) {
		c.Assert(synReply.CompressedHeaders).HasLen(4)
		c.Assert(synReply.CompressedHeaders[0]).Equals(byte(0x00))
		c.Assert(synReply.CompressedHeaders[1]).Equals(byte(0x01))
		c.Assert(synReply.CompressedHeaders[2]).Equals(byte(0x02))
		c.Assert(synReply.CompressedHeaders[3]).Equals(byte(0x03))
	})
})

var _ = spec.Suite("RST_STREAM", func(c *spec.C) {
	var rstStreamBody = []byte{
		0x00, 0x00, 0x02, 0x9B, // |X|          Stream-ID (31bits)    |
		0x00, 0x00, 0x00, 0x32, // |          Status code             |
	}

	var r = bytes.NewBuffer(rstStreamBody)
	rstStream := new(RstStream)
	_, err := rstStream.Read(r)
	c.Assert(err).IsNil()

	c.It("should be type 3", func(c *spec.C) {
		c.Assert(RstStreamType).Equals(FrameType(3))
	})

	c.It("should read its stream id", func(c *spec.C) {
		c.Assert(rstStream.StreamId).Equals(uint32(667))
	})

	c.It("should read its status code", func(c *spec.C) {
		c.Assert(rstStream.StatusCode).Equals(int32(50))
	})
})

var _ = spec.Suite("SETTINGS", func(c *spec.C) {
	c.It("should be type 4", func(c *spec.C) {
		c.Assert(SettingsType).Equals(FrameType(4))
	})

	c.It("should read an empty settings frame", func(c *spec.C) {
		r := bytes.NewBuffer([]byte{
			0x00, 0x00, 0x00, 0x00, // |         Number of entries        |
		})
		settings := new(Settings)
		_, err := settings.Read(r)
		c.Assert(err).IsNil()
		c.Assert(settings.Settings).HasLen(0)
	})

	c.It("should read a populated settings frame", func(c *spec.C) {
		r := bytes.NewBuffer([]byte{
			0x00, 0x00, 0x00, 0x01, // |         Number of entries        |
			0x01, 0x00, 0x00, 0x23, // | Flags(8) |      ID (24 bits)     |
			0x00, 0x00, 0x02, 0x9C, // |          Value (32 bits)         |
		})
		settings := new(Settings)
		_, err := settings.Read(r)
		c.Assert(err).IsNil()
		c.Assert(settings.Settings).HasLen(1)
		s0 := settings.Settings[0]
		c.Assert(s0.Flags).Equals(uint8(1))
		c.Assert(s0.Id).Equals(uint32(35))
		c.Assert(s0.Value).Equals(int32(668))
	})
})

var _ = spec.Suite("PING", func(c *spec.C) {
	c.It("should be type 6", func(c *spec.C) {
		c.Assert(PingType).Equals(FrameType(6))
	})

	c.It("should read", func(c *spec.C) {
		r := bytes.NewBuffer([]byte{
			0x00, 0x00, 0x66, 0x12, // |            32-bit ID             |
		})
		ping := new(Ping)
		_, err := ping.Read(r)
		c.Assert(err).IsNil()
		c.Assert(ping.Id).Equals(uint32(26130))
	})
})

var _ = spec.Suite("GOAWAY", func(c *spec.C) {
	c.It("should be type 7", func(c *spec.C) {
		c.Assert(GoAwayType).Equals(FrameType(7))
	})

	c.It("should read", func(c *spec.C) {
		r := bytes.NewBuffer([]byte{
			0x00, 0x00, 0x02, 0x9A, // |X|  Last-good-stream-ID (31 bits) |
			0x00, 0x00, 0x24, 0x68, // |          Status code             |
		})
		goaway := new(GoAway)
		_, err := goaway.Read(r)
		c.Assert(err).IsNil()
		c.Assert(goaway.LastGoodStreamId).Equals(uint32(666))
		c.Assert(goaway.StatusCode).Equals(uint32(9320))
	})
})

var _ = spec.Suite("HEADERS", func(c *spec.C) {
	c.It("should be type 8", func(c *spec.C) {
		c.Assert(HeadersType).Equals(FrameType(8))
	})

	c.It("should read", func(c *spec.C) {
		var r = bytes.NewBuffer([]byte{
			0x00, 0x00, 0x02, 0x9A, // |X|           Stream-ID (31bits)     |
			0x03, 0x02, 0x01, 0x00, // | Raw headers...                     |
		})

		headers := new(Headers)
		_, err := headers.Read(r)
		c.Assert(err).IsNil()

		c.Assert(headers.StreamId).Equals(uint32(666))
		c.Assert(headers.CompressedHeaders).HasLen(4)
		h := headers.CompressedHeaders
		c.Assert(h[0]).Equals(byte(0x03))
		c.Assert(h[1]).Equals(byte(0x02))
		c.Assert(h[2]).Equals(byte(0x01))
		c.Assert(h[3]).Equals(byte(0x00))
	})
})

var _ = spec.Suite("WINDOW_UPDATE", func(c *spec.C) {
	c.It("should be type 9", func(c *spec.C) {
		c.Assert(WindowUpdateType).Equals(FrameType(9))
	})

	c.It("should allow a negative delta window size?", func(c *spec.C) {
		c.Skip("unsure")
	})

	c.It("should read", func(c *spec.C) {
		var r = bytes.NewBuffer([]byte{
			0x00, 0x00, 0x02, 0x99, // |X|     Stream-ID (31-bits)        |
			0x00, 0x00, 0x00, 0x99, // |X|  Delta-Window-Size (31-bits)   |
		})
		windowUpdate := new(WindowUpdate)
		_, err := windowUpdate.Read(r)
		c.Assert(err).IsNil()

		c.Assert(windowUpdate.StreamId).Equals(uint32(665))
		c.Assert(windowUpdate.DeltaWindowSize).Equals(uint32(153))
	})
})

var _ = spec.Suite("CREDENTIAL", func(c *spec.C) {
	c.Skip("impl pending")

	c.It("should be type 10", func(c *spec.C) {
		c.Assert(CredentialType).Equals(FrameType(10))
	})
})
