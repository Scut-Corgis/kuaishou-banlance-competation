// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha256 implements the SHA224 and SHA256 hash algorithms as defined
// in FIPS 180-4.
package sha256

import (
	"crypto"
	"crypto/internal/boring"
	"encoding/binary"
	"errors"
	"hash"
	"unsafe"
)

func init() {
	crypto.RegisterHash(crypto.SHA224, New224)
	crypto.RegisterHash(crypto.SHA256, New)

	// 查看是否支持AVX2指令集
	//fmt.Println("useAVX2:", useAVX2)
}

// The size of a SHA256 checksum in bytes.
const Size = 32

// The size of a SHA224 checksum in bytes.
const Size224 = 28

// The blocksize of SHA256 and SHA224 in bytes.
const BlockSize = 64

const (
	chunk     = 64
	init0     = 0x6A09E667
	init1     = 0xBB67AE85
	init2     = 0x3C6EF372
	init3     = 0xA54FF53A
	init4     = 0x510E527F
	init5     = 0x9B05688C
	init6     = 0x1F83D9AB
	init7     = 0x5BE0CD19
	init0_224 = 0xC1059ED8
	init1_224 = 0x367CD507
	init2_224 = 0x3070DD17
	init3_224 = 0xF70E5939
	init4_224 = 0xFFC00B31
	init5_224 = 0x68581511
	init6_224 = 0x64F98FA7
	init7_224 = 0xBEFA4FA4
)

var abc = []byte{
	0x80, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x02, 0x00,
}

// digest represents the partial evaluation of a checksum.
type digest struct {
	h     [8]uint32
	x     [chunk]byte
	nx    int
	len   uint64
	is224 bool // mark if this digest is SHA-224
}

const (
	magic224      = "sha\x02"
	magic256      = "sha\x03"
	marshaledSize = len(magic256) + 8*4 + chunk + 8
)

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, marshaledSize)
	if d.is224 {
		b = append(b, magic224...)
	} else {
		b = append(b, magic256...)
	}
	b = binary.BigEndian.AppendUint32(b, d.h[0])
	b = binary.BigEndian.AppendUint32(b, d.h[1])
	b = binary.BigEndian.AppendUint32(b, d.h[2])
	b = binary.BigEndian.AppendUint32(b, d.h[3])
	b = binary.BigEndian.AppendUint32(b, d.h[4])
	b = binary.BigEndian.AppendUint32(b, d.h[5])
	b = binary.BigEndian.AppendUint32(b, d.h[6])
	b = binary.BigEndian.AppendUint32(b, d.h[7])
	b = append(b, d.x[:d.nx]...)
	b = b[:len(b)+len(d.x)-d.nx] // already zero
	b = binary.BigEndian.AppendUint64(b, d.len)
	return b, nil
}

func (d *digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic224) || (d.is224 && string(b[:len(magic224)]) != magic224) || (!d.is224 && string(b[:len(magic256)]) != magic256) {
		return errors.New("crypto/sha256: invalid hash state identifier")
	}
	if len(b) != marshaledSize {
		return errors.New("crypto/sha256: invalid hash state size")
	}
	b = b[len(magic224):]
	b, d.h[0] = consumeUint32(b)
	b, d.h[1] = consumeUint32(b)
	b, d.h[2] = consumeUint32(b)
	b, d.h[3] = consumeUint32(b)
	b, d.h[4] = consumeUint32(b)
	b, d.h[5] = consumeUint32(b)
	b, d.h[6] = consumeUint32(b)
	b, d.h[7] = consumeUint32(b)
	b = b[copy(d.x[:], b):]
	b, d.len = consumeUint64(b)
	d.nx = int(d.len % chunk)
	return nil
}

func consumeUint64(b []byte) ([]byte, uint64) {
	_ = b[7]
	x := uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	return b[8:], x
}

func consumeUint32(b []byte) ([]byte, uint32) {
	_ = b[3]
	x := uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	return b[4:], x
}

var (
	DigestCopy = digest{
		h: [8]uint32{
			init0, init1, init2, init3, init4, init5, init6, init7,
		},
		nx:  0,
		len: 0,
	}
)

// func (d *digest) Reset() {
// 	*d = DigestCopy
// }

func (d *digest) Reset() {
	d.h[0] = init0
	d.h[1] = init1
	d.h[2] = init2
	d.h[3] = init3
	d.h[4] = init4
	d.h[5] = init5
	d.h[6] = init6
	d.h[7] = init7
	d.nx = 0
	d.len = 0
}

// New returns a new hash.Hash computing the SHA256 checksum. The Hash
// also implements encoding.BinaryMarshaler and
// encoding.BinaryUnmarshaler to marshal and unmarshal the internal
// state of the hash.
func New() hash.Hash {
	if boring.Enabled {
		return boring.NewSHA256()
	}
	d := new(digest)
	d.Reset()
	return d
}

type MyHash interface {
	// Write (via the embedded io.Writer interface) adds more data to the running hash.
	// It never returns an error.
	Write2(p []byte)

	// Sum appends the current hash to b and returns the resulting slice.
	// It does not change the underlying hash state.
	Sum2(b []byte) []byte

	// Reset resets the Hash to its initial state.
	Reset()

	// Size returns the number of bytes Sum will return.
	Size() int

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if all writes
	// are a multiple of the block size.
	BlockSize() int
}

// func New2() MyHash {
// 	d := new(digest)
// 	d.Reset()
// 	return d
// }

func New2() *digest {
	d := new(digest)
	d.Reset()
	return d
}

// New224 returns a new hash.Hash computing the SHA224 checksum.
func New224() hash.Hash {
	if boring.Enabled {
		return boring.NewSHA224()
	}
	d := new(digest)
	d.is224 = true
	d.Reset()
	return d
}

func (d *digest) Size() int {
	if !d.is224 {
		return Size
	}
	return Size224
}

func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Write(p []byte) (nn int, err error) {
	//fmt.Println("---Write---")
	boring.Unreachable()
	nn = len(p)
	d.len += uint64(nn)
	//fmt.Println("d.nx:", d.nx)
	//fmt.Println("d.len:", d.len)
	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == chunk {
			Block(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}
	if len(p) >= chunk {
		n := len(p) &^ (chunk - 1)
		Block(d, p[:n])
		p = p[n:]
	}
	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}
	return
}

// p长度恒为64
func (d *digest) Write2(p []byte) {
	Block(d, p[:64])
}

// 输入恒为abc
func (d *digest) Write3() {
	Block2(d)
}

func (d *digest) Sum(in []byte) []byte {
	boring.Unreachable()
	// Make a copy of d so that caller can keep writing and summing.
	d0 := *d
	hash := d0.checkSum()
	if d0.is224 {
		return append(in, hash[:Size224]...)
	}
	return append(in, hash[:]...)
}

func (d *digest) Sum2(in []byte) []byte {
	hash := d.checkSum2()
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [Size]byte {
	//fmt.Println("---checkSum-----")
	len := d.len
	// Padding. Add a 1 bit and 0 bits until 56 bytes mod 64.
	var tmp [64 + 8]byte // padding + length buffer
	tmp[0] = 0x80
	var t uint64
	if len%64 < 56 {
		t = 56 - len%64
	} else {
		t = 64 + 56 - len%64
	}

	// Length in bits.
	len <<= 3
	padlen := tmp[:t+8]
	binary.BigEndian.PutUint64(padlen[t+0:], len)
	//fmt.Println("t:", t)
	//fmt.Println("len:", t+8, "  padlen:", padlen)
	d.Write(padlen)

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte

	binary.BigEndian.PutUint32(digest[0:], d.h[0])
	binary.BigEndian.PutUint32(digest[4:], d.h[1])
	binary.BigEndian.PutUint32(digest[8:], d.h[2])
	binary.BigEndian.PutUint32(digest[12:], d.h[3])
	binary.BigEndian.PutUint32(digest[16:], d.h[4])
	binary.BigEndian.PutUint32(digest[20:], d.h[5])
	binary.BigEndian.PutUint32(digest[24:], d.h[6])
	if !d.is224 {
		binary.BigEndian.PutUint32(digest[28:], d.h[7])
	}

	return digest
}

func (d *digest) checkSum2() [Size]byte {
	// padlen := abc
	d.Write3()
	// 汇编已经做了大小端转换
	var digest [Size]byte = *(*[Size]byte)(unsafe.Pointer(&d.h[0]))

	// binary.BigEndian.PutUint32(digest[0:], d.h[0])
	// binary.BigEndian.PutUint32(digest[4:], d.h[1])
	// binary.BigEndian.PutUint32(digest[8:], d.h[2])
	// binary.BigEndian.PutUint32(digest[12:], d.h[3])
	// binary.BigEndian.PutUint32(digest[16:], d.h[4])
	// binary.BigEndian.PutUint32(digest[20:], d.h[5])
	// binary.BigEndian.PutUint32(digest[24:], d.h[6])
	// if !d.is224 {
	// 	binary.BigEndian.PutUint32(digest[28:], d.h[7])
	// }
	return digest
	// a := *(*[32]byte)(unsafe.Pointer(&d.h))
	// return a
}

func (d *digest) GetH() *byte {
	// padlen := abc
	// 汇编已经做了大小端转换
	return (*byte)(unsafe.Pointer(&d.h[0]))

	// binary.BigEndian.PutUint32(digest[0:], d.h[0])
	// binary.BigEndian.PutUint32(digest[4:], d.h[1])
	// binary.BigEndian.PutUint32(digest[8:], d.h[2])
	// binary.BigEndian.PutUint32(digest[12:], d.h[3])
	// binary.BigEndian.PutUint32(digest[16:], d.h[4])
	// binary.BigEndian.PutUint32(digest[20:], d.h[5])
	// binary.BigEndian.PutUint32(digest[24:], d.h[6])
	// if !d.is224 {
	// 	binary.BigEndian.PutUint32(digest[28:], d.h[7])
	// }
	// a := *(*[32]byte)(unsafe.Pointer(&d.h))
	// return a
}

// Sum256 returns the SHA256 checksum of the data.
func Sum256(data []byte) [Size]byte {
	if boring.Enabled {
		return boring.SHA256(data)
	}
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}

// Sum224 returns the SHA224 checksum of the data.
func Sum224(data []byte) [Size224]byte {
	if boring.Enabled {
		return boring.SHA224(data)
	}
	var d digest
	d.is224 = true
	d.Reset()
	d.Write(data)
	sum := d.checkSum()
	ap := (*[Size224]byte)(sum[:])
	return *ap
}
