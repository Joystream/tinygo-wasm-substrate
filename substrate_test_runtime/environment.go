package substratetestruntime

import (
	"bytes"
	"io"
	"strconv"
	"unsafe"
)

// The code assumes that *byte == uint32

// Provided by the host

//go:export ext_clear_prefix
func ext_clear_prefix(prefix_data *byte, prefix_len uint32)

//go:export ext_print_utf8
func ext_print_utf8(utf8_data *byte, utf8_len uint32)

//go:export ext_set_storage
func ext_set_storage(key_data *byte, key_len uint32, value_data *byte, value_len uint32)

//go:export ext_get_allocated_storage
func ext_get_allocated_storage(key_data *byte, key_len uint32, written_out_addr *uint32) *byte

//go:export ext_blake2_256
func ext_blake2_256(data *byte, len uint32, out *byte)

//go:export ext_twox_128
func ext_twox_128(data *byte, len uint32, out *byte)

//go:export ext_twox_256
func ext_twox_256(data *byte, len uint32, out *byte)

//go:export ext_ed25519_verify
func ext_ed25519_verify(msg_data *byte, msg_len uint32, sig_data *byte, pubkey_data *byte) uint32

//go:export ext_blake2_256_enumerated_trie_root
func ext_blake2_256_enumerated_trie_root(values_data *byte, lens_data_addr *uint32, lens_len uint32, resultPtr *byte)

// Helper functions

func getOffset(b []byte) *byte {
	return &b[0]
}

func getLen(b []byte) uint32 {
	return uint32(len(b))
}

func packedSlice(offset *byte, len uint32) uint64 {
	return uint64(len)<<32 + uint64(uintptr(unsafe.Pointer(offset)))
}

func returnSlice(b []byte) uint64 {
	len := uint32(len(b))
	if len == 0 {
		return 0
	}
	return packedSlice(getOffset(b), len)
}

// TODO: unsafe convertor asbytes(string)

func print(s string) {
	b := []byte(s)
	ext_print_utf8(&b[0], uint32(len(b)))
}

//go:export io_get_stdout
func io_get_stdout() int32 {
	return 0 // Ignored in resource_write
}

//go:export resource_write
func resource_write(id int32, ptr *byte, len int32) int32 {
	// Note that ext_print_utf8 appends newline, which resource_write is not supposed to do...
	ext_print_utf8(ptr, uint32(len))
	return len
}

type MemReader struct {
	curPtr uintptr
	end    uintptr
}

func NewMemReader(offset *byte, length uintptr) MemReader {
	offsPtr := uintptr(unsafe.Pointer(offset))
	return MemReader{offsPtr, offsPtr + length}
}

func (r *MemReader) Read(p []byte) (n int, err error) {
	print(strconv.Itoa(int(r.curPtr)))
	for i := range p {
		if r.curPtr >= r.end {
			return i, io.EOF
		}
		p[i] = *((*byte)(unsafe.Pointer(r.curPtr)))
		r.curPtr++
	}
	return len(p), nil
}

func enumeratedTrieRootBlake256ForByteSlices(values [][]byte) [32]byte {
	lengths := make([]uint32, len(values))
	for i, v := range values {
		lengths[i] = getLen([]byte(v))
	}
	joined := bytes.Join(values, []byte{})
	var result [32]byte
	resultPtr := &result[0]

	ext_blake2_256_enumerated_trie_root(
		getOffset([]byte(joined)),
		&lengths[0],
		uint32(len(lengths)),
		resultPtr,
	)
	return result
}

// func expectCollectionSize(expected int, pd paritycodec.Decoder) {
// 	n := int(pd.DecodeUintCompact())
// 	if n != expected {
// 		panic("Expected a collection of " + strconv.Itoa(expected) + " elements, got " + strconv.Itoa(n))
// 	}
// }
