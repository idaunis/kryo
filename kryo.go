package kryo

import (
	"errors"
	"unicode/utf8"
)

type Kryo struct {
	buffer   []byte
	position int
	limit    int
	err      error
}

func New(buffer []byte) *Kryo {
	return &Kryo{buffer, 0, len(buffer), nil}
}

func (p *Kryo) read() byte {
	if p.position >= p.limit {
		p.position = p.limit - 1
		p.err = errors.New("Can not read more bytes from the buffer. Limit reached.")
	}
	b := p.buffer[p.position]
	if p.position < p.limit {
		p.position++
	}
	return b
}

func (p *Kryo) ReadByte() byte {
	return p.read()
}

func (p *Kryo) ReadInt() int32 {
	position := p.position
	p.position = position + 4
	return int32(p.buffer[position]&0xFF)<<24 |
		int32(p.buffer[position+1]&0xFF)<<16 |
		int32(p.buffer[position+2]&0xFF)<<8 |
		int32(p.buffer[position+3]&0xFF)
}

func (p *Kryo) ReadLong() int64 {
	position := p.position
	p.position = position + 8
	return int64(p.buffer[position])<<56 |
		int64(p.buffer[position+1]&0xFF)<<48 |
		int64(p.buffer[position+2]&0xFF)<<40 |
		int64(p.buffer[position+3]&0xFF)<<32 |
		int64(p.buffer[position+4]&0xFF)<<24 |
		int64(p.buffer[position+5]&0xFF)<<16 |
		int64(p.buffer[position+6]&0xFF)<<8 |
		int64(p.buffer[position+7]&0xFF)
}

func (p *Kryo) ReadIntWithOptimize(optimizePositive bool) int64 {
	var result int64
	var b byte

	b = p.buffer[p.position]
	p.position++
	result = int64(b & 0x3f)
	if (b & 0x80) != 0 {
		b = p.read()
		result |= int64(b&0x7F) << 7
		if (b & 0x80) != 0 {
			b = p.read()
			result |= int64(b&0x7F) << 14
			if (b & 0x80) != 0 {
				b = p.read()
				result |= int64(b&0x7F) << 21
				if (b & 0x80) != 0 {
					b = p.read()
					result |= int64(b&0x7F) << 28
				}
			}
		}
	}

	if optimizePositive {
		return result
	}
	return (result >> 1) ^ -(result & 1)
}

func (p *Kryo) ReadBytes(count int) []byte {
	ini := p.position
	end := ini + count
	p.position = end
	return p.buffer[ini:end]
}

func (p *Kryo) readUtf8Length(b byte) int {
	result := int(b & 0x3f)
	if b&0x40 != 0 {
		b = p.read()
		result |= int(b&0x7F) << 6
		if (b & 0x80) != 0 {
			b = p.read()
			result |= int(b&0x7F) << 13
			if (b & 0x80) != 0 {
				b = p.read()
				result |= int(b&0x7F) << 20
				if (b & 0x80) != 0 {
					b = p.read()
					result |= int(b&0x7F) << 27
				}
			}
		}
	}
	return result
}

func (p *Kryo) readUtf8(count int) string {
	str := make([]byte, 0)
	for i := 0; i < count; i++ {
		b := p.read()
		str = append(str, b)
	}
	for utf8.RuneCount(str) < count {
		_, size := utf8.DecodeRune(p.buffer[p.position:])
		for i := 0; i < size; i++ {
			b := p.read()
			str = append(str, b)
		}
	}

	return string(str)
}

func (p *Kryo) ReadString() string {
	var b byte
	str := make([]byte, 0)
	pos := 0
	for {
		b = p.buffer[p.position]
		p.position++
		if pos == 0 && b&0x80 != 0 {
			// Deserialize utf8 string
			count := p.readUtf8Length(b)
			if count == 1 {
				return ""
			}
			if count == 0 {
				return ""
			}
			count--
			return p.readUtf8(count)
		}
		if b&0x80 != 0 {
			// End of String
			b = b & 0x7f
			p.buffer[p.position-1] &= 0x7f
			str = append(str, b)
			p.buffer[p.position-1] |= 0x80
			break
		}
		str = append(str, b)
		pos++
	}

	return string(str)
}
