package wasmhelpers

import (
	"io"
	"math"
	"unsafe"
)

// Helper functions to translate between Go functions / structures
// and WebAssembly / Substrate calling conventions

func GetOffset(b []byte) *byte {
	if len(b) == 0 {
		return nil
	}
	return &b[0]
}

func GetLen(b []byte) uintptr {
	return uintptr(len(b))
}

func PackedSlice(offset *byte, len uintptr) uint64 {
	return uint64(len)<<32 + uint64(uintptr(unsafe.Pointer(offset)))
}

func ReturnSlice(b []byte) uint64 {
	len := uintptr(len(b))
	if len == 0 {
		return 0
	}
	return PackedSlice(GetOffset(b), len)
}

// TODO: unsafe convertor asbytes(string)

type MemReader struct {
	curPtr uintptr
	end    uintptr
}

func NewMemReader(offset *byte, length uintptr) MemReader {
	offsPtr := uintptr(unsafe.Pointer(offset))
	return MemReader{offsPtr, offsPtr + length}
}

func (r *MemReader) Read(p []byte) (n int, err error) {
	for i := range p {
		if r.curPtr >= r.end {
			return i, io.EOF
		}
		p[i] = *((*byte)(unsafe.Pointer(r.curPtr)))
		r.curPtr++
	}
	return len(p), nil
}

func Slice(offset *byte, length uintptr) []byte {
	arrayZeroPtr := (*[math.MaxInt32]byte)(unsafe.Pointer(uintptr(0)))
	uo := uintptr(unsafe.Pointer(offset))
	ul := uintptr(length)
	return (*arrayZeroPtr)[uo : uo+ul]
}

func ConcatByteSlices(a []byte, b []byte) []byte {
	r := make([]byte, len(a)+len(b))
	copy(r[:len(a)], a)
	copy(r[len(a):], b)
	return r
}

// TODO: why do we need this, as opposed to runtime.memset
//go:export memset
func memset(ptr unsafe.Pointer, c byte, size uintptr) unsafe.Pointer {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
	return ptr
}

// // TODO: why do we need this, as opposed to runtime.memmove
// //go:export memmove
// func memmove(dst, src unsafe.Pointer, size uintptr) {
// 	if uintptr(dst) < uintptr(src) {
// 		// Copy forwards.
// 		memcpy(dst, src, size)
// 		return
// 	}
// 	// Copy backwards.
// 	for i := size; i != 0; {
// 		i--
// 		*(*uint8)(unsafe.Pointer(uintptr(dst) + i)) = *(*uint8)(unsafe.Pointer(uintptr(src) + i))
// 	}
// }
