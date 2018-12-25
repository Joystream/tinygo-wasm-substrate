package wasmexecutortest

import "unsafe"

// The code assumes that uintptr == uint32

//go:export ext_clear_prefix
func ext_clear_prefix(prefix_data uintptr, prefix_len uint32)

//go:export ext_print_utf8
func ext_print_utf8(utf8_data uintptr, utf8_len uint32)

func returnSlice(b []byte) uint64 {
	offset := uintptr(unsafe.Pointer(&b[0]))
	return uint64(len(b))<<32 + uint64(offset)
}

//go:export test_clear_prefix
func test_clear_prefix(offset uintptr, len uint32) uint64 {
	ext_clear_prefix(offset, len)
	return returnSlice([]byte("all ok!"))
}

//go:export io_get_stdout
func io_get_stdout() int32 {
	return 0 // Ignored in resource_write
}

//go:export resource_write
func resource_write(id int32, ptr *uint8, len int32) int32 {
	ext_print_utf8(uintptr(unsafe.Pointer(ptr)), uint32(len))
	return len
}

func main() {
}
