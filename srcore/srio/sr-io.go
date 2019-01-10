package srio

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	. "github.com/Joystream/tinygo-wasm-substrate/wasmhelpers"
)

// Adapters for external functions, provided by the host
// (Matches "Externalities" in Substrate)

// Debug output

func Print(s string) {
	b := []byte(s)
	ext_print_utf8(&b[0], uintptr(len(b)))
}

// Debug printing of a byte array. ASCII characters are printed as is
// TODO: bytes as decimal ints, hexadecimal

func SprintBytes(bs []byte) string {
	s1 := []string{}
	for _, b := range bs {
		if b >= 32 && b < 128 {
			s1 = append(s1, string(rune(b)))
		} else {
			ss := strconv.FormatUint(uint64(b), 16)
			if len(ss) < 2 {
				ss = "\\x0" + ss
			} else {
				ss = "\\x" + ss
			}
		}
	}
	return strings.Join(s1, "")
}

//go:export io_get_stdout
func io_get_stdout() int32 {
	return 0 // Ignored in resource_write
}

var printBuffer = []byte{}

//go:export resource_write
func resource_write(id int32, ptr *byte, length uintptr) uintptr {
	// Implementation similar to resource_write in wasm_exec.js
	for i := uintptr(0); i < length; i++ {
		ptr2 := uintptr(unsafe.Pointer(ptr)) + i
		c := *(*byte)(unsafe.Pointer(ptr2))
		if c == 10 {
			ext_print_utf8(GetOffset(printBuffer), GetLen(printBuffer))
			printBuffer = printBuffer[:0]
		} else if c == 13 {
			// ignore
		} else {
			printBuffer = append(printBuffer, c)
		}
	}
	return length
}

func EnumeratedTrieRootBlake256ForByteSlices(values [][]byte) [32]byte {
	lengths := make([]uintptr, len(values))
	for i, v := range values {
		lengths[i] = GetLen([]byte(v))
	}
	joined := bytes.Join(values, []byte{})
	var result [32]byte
	resultPtr := &result[0]

	ptrLengths := (*uintptr)(nil)
	if len(lengths) > 0 {
		ptrLengths = &lengths[0]
	}
	ext_blake2_256_enumerated_trie_root(
		GetOffset([]byte(joined)),
		ptrLengths,
		uintptr(len(lengths)),
		resultPtr,
	)
	return result
}

func StoragePut(key []byte, value []byte) {
	ext_set_storage(GetOffset(key), GetLen(key), GetOffset(value), GetLen(value))
}

func StorageGet(key []byte) (bool, []byte) {
	var valueLen uintptr
	valuePtr := ext_get_allocated_storage(GetOffset(key), GetLen(key), &valueLen)
	if valueLen == math.MaxUint32 {
		return false, []byte{}
	}
	return true, Slice(valuePtr, valueLen)
}

func StorageKill(key []byte) {
	ext_clear_storage(GetOffset(key), GetLen(key))
}

func StorageRoot() *primitives.H256 {
	var res primitives.H256
	ext_storage_root(&res)
	return &res
}

func StorageChangesRoot(parentHash []byte, parentNum uint64) (bool, *primitives.H256) {
	var res primitives.H256
	ok := ext_storage_changes_root(GetOffset(parentHash), GetLen(parentHash), parentNum, &res) > 0
	return ok, &res
}

func Twox128(v []byte) []byte {
	var res [16]byte
	ext_twox_128(GetOffset(v), GetLen(v), &res[0])
	return res[:]
}

func Twox256(v []byte) []byte {
	var res [32]byte
	ext_twox_256(GetOffset(v), GetLen(v), &res[0])
	return res[:]
}

func Blake256(v []byte) []byte {
	var res [32]byte
	ext_blake2_256(GetOffset(v), GetLen(v), &res[0])
	return res[:]
}
