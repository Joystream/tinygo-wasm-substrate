package main

import (
	"unsafe"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"

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

type State struct {
	counter uint32
}

func env_assert(_statePtr unsafe.Pointer, args primitives.TypedValues) primitives.ReturnValueOrHostError {
	if len(args) != 1 {
		return primitives.HostError{}
	}
	condition, ok := args[0].(primitives.I32)
	if !ok {
		return primitives.HostError{}
	}
	if condition.V != 0 {
		return primitives.Unit{}
	} else {
		return primitives.HostError{}
	}
}

func env_inc_counter(statePtr unsafe.Pointer, args primitives.TypedValues) primitives.ReturnValueOrHostError {
	state := ((*State)(statePtr))
	if len(args) != 1 {
		return primitives.HostError{}
	}
	inc_by, ok := args[0].(primitives.I32)
	if !ok {
		return primitives.HostError{}
	}
	state.counter += uint32(inc_by.V)
	return primitives.TypedReturnValue{primitives.I32{int32(state.counter)}}
}

func execute_sandboxed(code []byte, args primitives.TypedValues) primitives.ReturnValueOrHostError {

	state := State{}

	env_builder := srsandbox.EnvironmentDefinitionBuilder{}

	env_builder.AddHostFunc("env", "assert", env_assert)
	env_builder.AddHostFunc("env", "inc_counter", env_inc_counter)
	memory, err := srsandbox.NewMemory(1, 16)
	if err != srsandbox.NoError {
		return primitives.HostError{}
	}
	env_builder.AddMemory("env", "memory", memory)

	instance, err := srsandbox.NewInstance(code, env_builder, unsafe.Pointer(&state))
	if err != srsandbox.NoError {
		return primitives.HostError{}
	}
	return instance.Invoke("call", args, unsafe.Pointer(&state))
}

//go:export test_sandbox
func test_sandbox(codeOffstet *byte, codeLen uintptr) uint64 {
	code := wasmhelpers.Slice(codeOffstet, codeLen)
	res := execute_sandboxed(code, []primitives.TypedValue{})
	_, isErr := res.(primitives.HostError)
	var ok byte
	if !isErr {
		ok = 1
	}
	return wasmhelpers.PackedSlice(&ok, 1)
}

//go:export test_sandbox_args
func test_sandbox_args(codeOffstet *byte, codeLen uintptr) uint64 {
	code := wasmhelpers.Slice(codeOffstet, codeLen)
	res := execute_sandboxed(code, []primitives.TypedValue{primitives.I32{0x12345678}, primitives.I64{0x1234567887654321}})
	_, isErr := res.(primitives.HostError)
	var ok byte
	if !isErr {
		ok = 1
	}
	return wasmhelpers.PackedSlice(&ok, 1)
}

//go:export test_sandbox_return_val
func test_sandbox_return_val(codeOffstet *byte, codeLen uintptr) uint64 {
	code := wasmhelpers.Slice(codeOffstet, codeLen)
	res := execute_sandboxed(code, []primitives.TypedValue{primitives.I32{0x1336}})
	ret, notError := res.(primitives.TypedReturnValue)
	var ok byte
	if notError {
		v, notError := ret.Value.(primitives.I32)
		if notError && v.V == 0x1337 {
			ok = 1
		}
	}
	return wasmhelpers.PackedSlice(&ok, 1)
}
