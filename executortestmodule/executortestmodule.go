package main

import (
	"unsafe"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/srio"
	. "github.com/Joystream/tinygo-wasm-substrate/wasmhelpers"
)

//go:export test_data_in
func test_data_in(offset *byte, len uintptr) uint64 {
	print("set_storage")

	key := []byte("input")
	srio.UnhashedPut(key, Slice(offset, len))

	print("storage")
	key = []byte("foo")
	ok, foo := srio.UnhashedGet(key)
	if !ok {
		panic("No value for this key")
	}

	print("set_storage")
	key = []byte("baz")
	srio.UnhashedPut(key, foo)

	print("finished!")
	return ReturnSlice([]byte("all ok!"))
}

//go:export test_clear_prefix
func test_clear_prefix(offset *byte, len uintptr) uint64 {
	srio.Ext_clear_prefix(offset, len)
	return ReturnSlice([]byte("all ok!"))
}

//go:export test_empty_return
func test_empty_return(_ *byte, _ uintptr) uint64 {
	return ReturnSlice([]byte{})
}

//go:export test_panic
func test_panic() uint64 {
	panic("test panic")
	return 0
}

//go:export test_conditional_panic
func test_conditional_panic(offset *byte, len uintptr) uint64 {
	if len > 0 {
		panic("test panic")
	}
	return PackedSlice(offset, len)
}

//go:export test_blake2_256
func test_blake2_256(offset *byte, len uintptr) uint64 {
	return ReturnSlice(srio.Blake256(Slice(offset, len)))
}

//go:export test_twox_256
func test_twox_256(offset *byte, len uintptr) uint64 {
	return ReturnSlice(srio.Twox256(Slice(offset, len)))
}

//go:export test_twox_128
func test_twox_128(offset *byte, len uintptr) uint64 {
	return ReturnSlice(srio.Twox128(Slice(offset, len)))
}

//go:export test_ed25519_verify
func test_ed25519_verify(offset *byte, len uintptr) uint64 {
	pubkeyPtr := offset
	sigPtr := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(offset)) + 32))
	msg := []byte("all ok!")
	res := srio.Ext_ed25519_verify(GetOffset(msg), GetLen(msg), sigPtr, pubkeyPtr) == 0
	return PackedSlice((*byte)(unsafe.Pointer(&res)), 1)
}

//go:export test_enumerated_trie_root
func test_enumerated_trie_root(_ *byte, _ uintptr) uint64 {
	values := [][]byte{
		[]byte("zero"),
		[]byte("one"),
		[]byte("two")}
	a := srio.EnumeratedTrieRootBlake256ForByteSlices(values)
	return ReturnSlice(a[:])
}

func main() {
}
