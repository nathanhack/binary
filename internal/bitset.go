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

func (bsb *BitSetBuffer) ResetToStart() {
	bsb.pos = 0
}

func (bsb *BitSetBuffer) ResetToEnd() {
	bsb.pos = len(bsb.Set)
}

func (bsb *BitSetBuffer) PosAtEnd() bool {
	return bsb.pos == len(bsb.Set)
}

func (bsb *BitSetBuffer) Read(bytes []byte) (n int, err error) {
	if bytes == nil {
		return 0, fmt.Errorf("error nil passed in")
	}
	if bsb.PosAtEnd() {
		return 0, io.EOF
	}

	n = 0
	for n < len(bytes) && !bsb.PosAtEnd() {
		bytes[n] = bsb.readByte()
		n++
	}
	return n, nil
}

func (bsb *BitSetBuffer) ReadBits(bits []bool) (n int, err error) {
	if bits == nil {
		return 0, fmt.Errorf("error nil passed in")
	}
	for n = 0; n < len(bits); n++ {
		if bsb.pos >= len(bsb.Set) {
			return
		}
		bits[n] = bsb.Set[bsb.pos]
		bsb.pos++
	}
	return
}

func (bsb *BitSetBuffer) Write(bytes []byte) (n int, err error) {
	n = 0
	for n < len(bytes) {
		bsb.writeByte(bytes[n])
		n++
	}
	return n, nil
}

func (bsb *BitSetBuffer) WriteBits(bits []bool) (n int, err error) {
	if bsb.Set == nil {
		bsb.Set = make([]bool, 0)
	}

	for n = 0; n < len(bits); n++ {
		if bsb.pos < len(bsb.Set) {
			bsb.Set[bsb.pos] = bits[n]
		} else {
			bsb.Set = append(bsb.Set, bits[n])
		}
		bsb.pos++
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (bsb *BitSetBuffer) readByte() (b byte) {
	b = 0
	defer func() {
		bsb.pos = min(len(bsb.Set), bsb.pos+8)
	}()

	for i := 0; i < 8; i++ {
		index := i + bsb.pos
		if index >= len(bsb.Set) {
			return
		}
		if bsb.Set[index] {
			b = b | (1 << i)
		}
	}
	return
}

func (bsb *BitSetBuffer) writeByte(b byte) {
	for i := 0; i < 8; i++ {
		value := b&(1<<i) > 0

		if bsb.pos < len(bsb.Set) {
			bsb.Set[bsb.pos] = value
		} else {
			if bsb.Set == nil {
				bsb.Set = make([]bool, 0)
			}
			bsb.Set = append(bsb.Set, value)
		}
		bsb.pos++
	}
}

func (bsb *BitSetBuffer) Bytes() []byte {
	old := bsb.pos
	defer func() {
		bsb.pos = old
	}()
	bsb.ResetToStart()
	buf := make([]byte, 0)
	for !bsb.PosAtEnd() {
		buf = append(buf, bsb.readByte())
	}
	return buf
}
