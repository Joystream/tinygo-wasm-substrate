package substratetestruntime

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strconv"
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
func ext_get_allocated_storage(key_data *byte, key_len uint32, value_len_ptr *uint32) *byte

//go:export ext_clear_storage
func ext_clear_storage(key_data *byte, key_len uint32)

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

//go:export ext_storage_root
func ext_storage_root(resultPtr *H256)

//go:export ext_storage_changes_root
func ext_storage_changes_root(parent_hash_data *byte, parent_hash_len uint32, parent_num uint64, result *H256) uint32

// Helper functions

func getOffset(b []byte) *byte {
	if len(b) == 0 {
		return nil
	}
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

func sprintBytes(bs []byte) string {
	s1 := []string{}
	s2 := []string{}
	for _, b := range bs {
		ss := strconv.FormatUint(uint64(b), 10)
		if len(ss) < 2 {
			ss = "0" + ss
		}
		s2 = append(s2, ss)
		if b >= 32 && b < 128 {
			s1 = append(s1, string(rune(b)))
		} else {
			s1 = append(s1, "?")
		}
	}
	return strings.Join(s1, "") + " / " + strings.Join(s2, " ")
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
	print("for")
	for i, v := range values {
		lengths[i] = getLen([]byte(v))
	}
	print("joining")
	joined := bytes.Join(values, []byte{})
	var result [32]byte
	resultPtr := &result[0]

	ptrLengths := (*uint32)(nil)
	if len(lengths) > 0 {
		ptrLengths = &lengths[0]
	}
	print("call ext")
	ext_blake2_256_enumerated_trie_root(
		getOffset([]byte(joined)),
		ptrLengths,
		uint32(len(lengths)),
		resultPtr,
	)
	print("ext done")
	return result
}

// func expectCollectionSize(expected int, pd paritycodec.Decoder) {
// 	n := int(pd.DecodeUintCompact())
// 	if n != expected {
// 		panic("Expected a collection of " + strconv.Itoa(expected) + " elements, got " + strconv.Itoa(n))
// 	}
// }

func storagePut(key []byte, value []byte) {
	key = hashStorageKey(key)
	ext_set_storage(getOffset(key), getLen(key), getOffset(value), getLen(value))
}

func storagePutUint64(key []byte, value uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	storagePut(key, buf)
}

func slice(offset *byte, length uint32) []byte {
	arrayZeroPtr := (*[math.MaxInt32]byte)(unsafe.Pointer(uintptr(0)))
	uo := uintptr(unsafe.Pointer(offset))
	ul := uintptr(length)
	return (*arrayZeroPtr)[uo : uo+ul]
}

func hashStorageKey(key []byte) []byte {
	result := make([]byte, 16)
	resultPtr := getOffset(result)
	ext_twox_128(getOffset(key), getLen(key), resultPtr)
	return result
}

func storageGet(key []byte) (bool, []byte) {
	key = hashStorageKey(key)
	var valueLen uint32
	print("###############")
	print(sprintBytes(key))
	valuePtr := ext_get_allocated_storage(getOffset(key), getLen(key), &valueLen)
	print(strconv.Itoa(int(valueLen)))
	if valueLen == math.MaxUint32 {
		return false, []byte{}
	}
	return true, slice(valuePtr, valueLen)
}

func storageGetUint64Or(key []byte, deflt uint64) uint64 {
	ok, buf := storageGet(key)
	if ok {
		return binary.LittleEndian.Uint64(buf)
	}
	return deflt
}

func storage_kill(key []byte) {
	key = hashStorageKey(key)
	ext_clear_storage(getOffset(key), getLen(key))
}

// TODO: avoid copy?
func storage_root() *H256 {
	var res H256
	ext_storage_root(&res)
	return &res
}

type Error struct {
	message string
}

func (e Error) Error() string {
	return e.message
}

func concatByteSlices(a []byte, b []byte) []byte {
	r := make([]byte, len(a)+len(b))
	copy(r[:len(a)], a)
	copy(r[len(a):], b)
	return r
}

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

func storage_changes_root(parentHash []byte, parentNum uint64) (bool, *H256) {
	var res H256
	ok := ext_storage_changes_root(getOffset(parentHash), getLen(parentHash), parentNum, &res) > 0
	return ok, &res
}
