package srsandbox

import (
	"bytes"
	"math"
	"strconv"
	"unsafe"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	"github.com/Joystream/tinygo-wasm-substrate/wasmhelpers"
	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Func func(state unsafe.Pointer, args primitives.TypedValues) primitives.ReturnValueOrHostError

type TinyGoClosure struct {
	// See https://github.com/aykevl/tinygo/blob/107fccb288f8ba6258d417c5e14921d4d97f3e64/compiler/compiler.go#L536
	// and http://fitzgeraldnick.com/2018/04/26/how-does-dynamic-dispatch-work-in-wasm.html

	// TODO: add tests to verify this structure is correct?..

	closureContext unsafe.Pointer
	wasmTableIndex uint32
}

func DispatchThunk(
	serialized_args_ptr *byte,
	serialized_args_len uintptr,
	state *byte,
	hostFuncId uint32,
) uint64 {
	serializedArgs := wasmhelpers.Slice(serialized_args_ptr, serialized_args_len)
	args := primitives.TypedValues{}
	paritycodec.FromBytes(&args, serializedArgs)

	closure := TinyGoClosure{unsafe.Pointer(nil), hostFuncId}
	funcObj := *(*Func)(unsafe.Pointer(&closure))

	// Pass control flow to the designated function.
	resultOrErr := funcObj(unsafe.Pointer(state), args)

	// TODO: dealloc result

	return wasmhelpers.ReturnSlice(paritycodec.ToBytesCustom(func(pe paritycodec.Encoder) {
		resultOrErr.ReturnValueEncode(pe)
	}))
}

func NewMemory(pages uint32, maxPages uint32) (primitives.ExternMemory, Error) {
	result := ext_sandbox_memory_new(pages, maxPages)
	switch result {
	case ERR_MODULE:
		return primitives.ExternMemory{}, ErrModule
	}
	return primitives.ExternMemory{result}, NoError
}

type Instance struct {
	instance_idx       uint32
	_retained_memories []primitives.ExternMemory
}

type EnvironmentDefinitionBuilder struct {
	env_def           primitives.EnvironmentDefinition
	retained_memories []primitives.ExternMemory
}

func (e EnvironmentDefinitionBuilder) AddHostFunc(module string, name string, function Func) {
	funcClosure := *(*TinyGoClosure)(unsafe.Pointer(&function))

	e.env_def.Entries = append(e.env_def.Entries,
		primitives.Entry{module, name, primitives.ExternFunction{uint32(funcClosure.wasmTableIndex)}})
}

func (e EnvironmentDefinitionBuilder) AddMemory(module string, name string, memory primitives.ExternMemory) {
	e.env_def.Entries = append(e.env_def.Entries,
		primitives.Entry{module, name, memory})
}

type Error int

const (
	NoError Error = iota

	/// Module is not valid, couldn't be instantiated or it's `start` function trapped
	/// when executed.
	ErrModule

	/// Access to a memory or table was made with an address or an index which is out of bounds.
	///
	/// Note that if wasm module makes an out-of-bounds access then trap will occur.
	ErrOutOfBounds

	/// Failed to invoke an exported function for some reason.
	ErrExecution
)

/// No error happened.
///
/// For FFI purposes.
const ERR_OK uint32 = 0

/// Validation or instantiation error occured when creating new
/// sandboxed module instance.
///
/// For FFI purposes.
const ERR_MODULE uint32 = math.MaxUint32

/// Out-of-bounds access attempted with memory or table.
///
/// For FFI purposes.
const ERR_OUT_OF_BOUNDS uint32 = math.MaxUint32 - 1

/// Execution error occurred (typically trap).
///
/// For FFI purposes.
const ERR_EXECUTION uint32 = math.MaxUint32 - 2

func NewInstance(code []byte, env_def_builder EnvironmentDefinitionBuilder, state unsafe.Pointer) (Instance, Error) {
	serialized_env_def := paritycodec.ToBytesCustom(func(pe paritycodec.Encoder) {
		pe.EncodeCollection(len(env_def_builder.env_def.Entries), func(i int) {
			env_def_builder.env_def.Entries[i].Entity.ExternEntityEncode(pe)
		})
	})

	dthunk := DispatchThunk
	dthunkClosure := *(*TinyGoClosure)(unsafe.Pointer(&dthunk))

	instanceIdx := ext_sandbox_instantiate(
		dthunkClosure.wasmTableIndex,
		wasmhelpers.GetOffset(code),
		wasmhelpers.GetLen(code),
		wasmhelpers.GetOffset(serialized_env_def),
		wasmhelpers.GetLen(serialized_env_def),
		state,
	)
	switch instanceIdx {
	case ERR_MODULE:
		return Instance{}, ErrModule
	case ERR_EXECUTION:
		return Instance{}, ErrExecution
	}

	// We need to retain memories to keep them alive while the Instance is alive.
	retainedMemories := env_def_builder.retained_memories // TODO: .clone()
	return Instance{instanceIdx, retainedMemories}, NoError
}

// TODO: return proper error type
func (i Instance) Invoke(
	name string,
	args primitives.TypedValues,
	state unsafe.Pointer,
) primitives.ReturnValueOrHostError {

	serializedArgs := paritycodec.ToBytes(args)
	returnValBuf := make([]byte, primitives.ENCODED_MAX_SIZE_RETURN_VALUE_OR_ERROR)
	nameBytes := ([]byte)(name)
	result := ext_sandbox_invoke(
		i.instance_idx,
		wasmhelpers.GetOffset(nameBytes), wasmhelpers.GetLen(nameBytes),
		wasmhelpers.GetOffset(serializedArgs), wasmhelpers.GetLen(serializedArgs),
		wasmhelpers.GetOffset(returnValBuf), wasmhelpers.GetLen(returnValBuf),
		state,
	)

	switch result {
	case ERR_OK:
		pd := paritycodec.Decoder{bytes.NewBuffer(returnValBuf)}
		return primitives.ReturnValueDecode(pd)
	case ERR_EXECUTION:
		return primitives.HostError{}
	default:
		panic("Invalid ext_sandbox_invoke result: " + strconv.Itoa(int(result)))
	}
}

// 		match result {
// 			sandbox_primitives::ERR_OK => {
// 				return_val = sandbox_primitives::ReturnValue::decode(&mut &return_val[..])
// 					.ok_or(Error::Execution)?;
// 				Ok(return_val)
// 			}
// 			sandbox_primitives::ERR_EXECUTION => Err(Error::Execution),
// 			_ => unreachable!(),
// 		}
// 	}
// }
