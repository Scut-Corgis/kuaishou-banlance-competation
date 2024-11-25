//go:build amd64 && !gccgo && !appengine
// +build amd64,!gccgo,!appengine

package hex

var (
	Lower = []byte("0123456789abcdef0123456789abcdef")
)

//go:generate go run asm_gen.go

// This function is implemented in hex_encode_amd64.s
//
//go:noescape
func EncodeAVX(dst *byte, src *byte, len uint64, alpha *byte)
