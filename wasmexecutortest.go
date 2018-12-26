package wasmexecutortest

import (
	"strings"
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

// Actually usable exported functions

//go:export test_data_in
func test_data_in(offset *byte, len uint32) uint64 {
	print("set_storage")

	key := []byte("input")
	ext_set_storage(getOffset(key), getLen(key), offset, len)

	print("storage")
	key = []byte("foo")
	var fooLen uint32
	fooOffset := ext_get_allocated_storage(getOffset(key), getLen(key), &fooLen)

	print("set_storage")
	key = []byte("baz")
	ext_set_storage(getOffset(key), getLen(key), fooOffset, fooLen)

	print("finished!")
	return returnSlice([]byte("all ok!"))
}

//go:export test_clear_prefix
func test_clear_prefix(offset *byte, len uint32) uint64 {
	ext_clear_prefix(offset, len)
	return returnSlice([]byte("all ok!"))
}

//go:export test_empty_return
func test_empty_return(_ *byte, _ uint32) uint64 {
	return returnSlice([]byte{})
}

//go:export test_panic
func test_panic() uint64 {
	panic("test panic")
	return 0
}

//go:export test_conditional_panic
func test_conditional_panic(offset *byte, len uint32) uint64 {
	if len > 0 {
		panic("test panic")
	}
	return packedSlice(offset, len)
}

//go:export test_blake2_256
func test_blake2_256(offset *byte, len uint32) uint64 {
	result := make([]byte, 32)
	resultPtr := getOffset(result)
	ext_blake2_256(offset, len, resultPtr)
	return packedSlice(resultPtr, 32)
}

//go:export test_twox_256
func test_twox_256(offset *byte, len uint32) uint64 {
	result := make([]byte, 32)
	resultPtr := getOffset(result)
	ext_twox_256(offset, len, resultPtr)
	return packedSlice(resultPtr, 32)
}

//go:export test_twox_128
func test_twox_128(offset *byte, len uint32) uint64 {
	result := make([]byte, 16)
	resultPtr := getOffset(result)
	ext_twox_128(offset, len, resultPtr)
	return packedSlice(resultPtr, 16)
}

//go:export test_ed25519_verify
func test_ed25519_verify(offset *byte, len uint32) uint64 {
	pubkeyPtr := offset
	sigPtr := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(offset)) + 32))
	msg := []byte("all ok!")
	res := ext_ed25519_verify(getOffset(msg), getLen(msg), sigPtr, pubkeyPtr) == 0
	return packedSlice((*byte)(unsafe.Pointer(&res)), 1)
}

//go:export test_enumerated_trie_root
func test_enumerated_trie_root(_ *byte, _ uint32) uint64 {
	lengths := make([]uint32, 3)
	values := []string{"zero", "one", "two"}
	for i, v := range values {
		lengths[i] = getLen([]byte(v))
	}
	joined := strings.Join(values, "")
	result := make([]byte, 32)
	resultPtr := getOffset(result)

	ext_blake2_256_enumerated_trie_root(
		getOffset([]byte(joined)),
		&lengths[0],
		uint32(len(lengths)),
		resultPtr,
	)
	return packedSlice(resultPtr, 32)
}

func main() {
}
