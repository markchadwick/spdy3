package spdy3

import (
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
})

var _ = spec.Suite("Flag/Len Word", func(c *spec.C) {
	var word FlagLenWord = 0xAB123456

	c.It("should know its flags", func(c *spec.C) {
		c.Assert(word.Flags()).Equals(uint8(171))
	})

	c.It("should know its length", func(c *spec.C) {
		c.Assert(word.Length()).Equals(uint32(1193046))
	})
})

var _ = spec.Suite("StreamId Word", func(c *spec.C) {
	var word StreamIdWord = 0xFFFFFFFF

	c.It("should know its stream ID", func(c *spec.C) {
		// Won't actually be 4294967295 -- first bit is dropped
		c.Assert(word.StreamId()).Equals(uint32(2147483647))
	})
})

var _ = spec.Suite("SYN_STREAM", func(c *spec.C) {
	c.It("should be type 1", func(c *spec.C) {
		c.Assert(SynStreamType).Equals(FrameType(1))
	})
})

var _ = spec.Suite("SYN_REPLY", func(c *spec.C) {
	c.It("should be type 2", func(c *spec.C) {
		c.Assert(SynReplyType).Equals(FrameType(2))
	})
})

var _ = spec.Suite("RST_STREAM", func(c *spec.C) {
	c.It("should be type 3", func(c *spec.C) {
		c.Assert(RstStreamType).Equals(FrameType(3))
	})
})

var _ = spec.Suite("SETTINGS", func(c *spec.C) {
	c.It("should be type 4", func(c *spec.C) {
		c.Assert(SettingsType).Equals(FrameType(4))
	})
})

var _ = spec.Suite("PING", func(c *spec.C) {
	c.It("should be type 6", func(c *spec.C) {
		c.Assert(PingType).Equals(FrameType(6))
	})
})

var _ = spec.Suite("GOAWAY", func(c *spec.C) {
	c.It("should be type 7", func(c *spec.C) {
		c.Assert(GoAwayType).Equals(FrameType(7))
	})
})

var _ = spec.Suite("HEADERS", func(c *spec.C) {
	c.It("should be type 8", func(c *spec.C) {
		c.Assert(HeadersType).Equals(FrameType(8))
	})
})

var _ = spec.Suite("WINDOW_UPDATE", func(c *spec.C) {
	c.It("should be type 9", func(c *spec.C) {
		c.Assert(WindowUpdateType).Equals(FrameType(9))
	})
})

var _ = spec.Suite("CREDENTIAL", func(c *spec.C) {
	c.It("should be type 10", func(c *spec.C) {
		c.Assert(CredentialType).Equals(FrameType(10))
	})
})
