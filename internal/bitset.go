package internal

import (
	"fmt"
	"io"
)

type BitSetReader interface {
	io.Reader
	ReadBits(bits []bool) (n int, err error)
}

type BitSetWriter interface {
	io.Writer
	WriteBits(bits []bool) (n int, err error)
}

type BitSetBuffer struct {
	pos int
	Set []bool
}

func NewFromBytes(bytes []byte) (*BitSetBuffer, error) {
	b := BitSetBuffer{}
	defer func() {
		b.pos = 0
	}()
	n, err := b.Write(bytes)
	if err != nil {
		return nil, err
	}
	if n != len(bytes) {
		return nil, fmt.Errorf("expected to store %v bytes in buffer but stored %v instead", len(bytes), n)
	}
	return &b, nil
}

func NewFromBits(bits []bool) (*BitSetBuffer, error) {
	b := BitSetBuffer{}
	defer func() {
		b.pos = 0
	}()
	n, err := b.WriteBits(bits)
	if err != nil {
		return nil, err
	}
	if n != len(bits) {
		return nil, fmt.Errorf("expected to store %v bits in buffer but stored %v instead", len(bits), n)
	}
	return &b, nil
}

func (set *BitSetBuffer) ResetToStart() {
	set.pos = 0
}

func (set *BitSetBuffer) ResetToEnd() {
	set.pos = len(set.Set)
}

func (set *BitSetBuffer) PosAtEnd() bool {
	return set.pos == len(set.Set)
}

func (set *BitSetBuffer) Read(bytes []byte) (n int, err error) {
	if bytes == nil {
		return 0, fmt.Errorf("error nil passed in")
	}
	n = 0
	for n < len(bytes) && !set.PosAtEnd() {
		bytes[n] = set.readByte()
		n++
	}
	return n, nil
}

func (set *BitSetBuffer) ReadBits(bits []bool) (n int, err error) {
	if bits == nil {
		return 0, fmt.Errorf("error nil passed in")
	}
	for n = 0; n < len(bits); n++ {
		if set.pos >= len(set.Set) {
			return
		}
		bits[n] = set.Set[set.pos]
		set.pos++
	}
	return
}

func (set *BitSetBuffer) Write(bytes []byte) (n int, err error) {
	n = 0
	for n < len(bytes) {
		set.writeByte(bytes[n])
		n++
	}
	return n, nil
}

func (set *BitSetBuffer) WriteBits(bits []bool) (n int, err error) {
	if set.Set == nil {
		set.Set = make([]bool, 0)
	}

	for n = 0; n < len(bits); n++ {
		if set.pos < len(set.Set) {
			set.Set[set.pos] = bits[n]
		} else {
			set.Set = append(set.Set, bits[n])
		}
		set.pos++
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (set *BitSetBuffer) readByte() (b byte) {
	b = 0
	defer func() {
		set.pos = min(len(set.Set), set.pos+8)
	}()

	for i := 0; i < 8; i++ {
		index := i + set.pos
		if index >= len(set.Set) {
			return
		}
		if set.Set[index] {
			b = b | (1 << i)
		}
	}
	return
}

func (set *BitSetBuffer) writeByte(b byte) {
	for i := 0; i < 8; i++ {
		value := b&(1<<i) > 0

		if set.pos < len(set.Set) {
			set.Set[set.pos] = value
		} else {
			if set.Set == nil {
				set.Set = make([]bool, 0)
			}
			set.Set = append(set.Set, value)
		}
		set.pos++
	}
}

func (set *BitSetBuffer) Bytes() []byte {
	old := set.pos
	defer func() {
		set.pos = old
	}()
	set.ResetToStart()
	buf := make([]byte, 0)
	for !set.PosAtEnd() {
		buf = append(buf, set.readByte())
	}
	return buf
}
