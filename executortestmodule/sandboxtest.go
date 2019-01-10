package main

import (
	"unsafe"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/srsandbox"
	"github.com/Joystream/tinygo-wasm-substrate/wasmhelpers"
)

type Unit struct {
}

//go:export test_sandbox_instantiate
func test_sandbox_instantiate(codeOffstet *byte, codeLen uintptr) uint64 {
	code := wasmhelpers.Slice(codeOffstet, codeLen)
	env_builder := srsandbox.EnvironmentDefinitionBuilder{}
	_, err := srsandbox.NewInstance(code, env_builder, unsafe.Pointer(nil))

	var result byte
	switch err {
	case srsandbox.NoError:
		result = 0
	case srsandbox.ErrModule:
		result = 1
	case srsandbox.ErrExecution:
		result = 2
	case srsandbox.ErrOutOfBounds:
		result = 3
	}
	return wasmhelpers.PackedSlice(&result, 1)
}
