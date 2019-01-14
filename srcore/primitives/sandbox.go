package primitives

import (
	"strconv"

	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

// Matches wasm basic types. Used for arguments in sandbox function calls.
type TypedValue interface {
	TypedValueEncode(paritycodec.Encoder)
}

type I32 struct{ V int32 }
type I64 struct{ V int64 }
type F32 struct{ V int32 } // TODO: do we really need float
type F64 struct{ V int64 }

func (x I32) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(1)
	pe.EncodeInt32(x.V)
}

func (x I64) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(2)
	pe.EncodeInt64(x.V)
}
func (x I64) ImplementsReturnValue() {}

func (x F32) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(3)
	pe.EncodeInt32(x.V)
}
func (x F32) ImplementsReturnValue() {}

func (x F64) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(4)
	pe.EncodeInt64(x.V)
}
func (x F64) ImplementsReturnValue() {}

type TypedValues []TypedValue

func (t TypedValues) ParityEncode(pe paritycodec.Encoder) {
	pe.EncodeCollection(len(t), func(i int) { t[i].TypedValueEncode(pe) })
}

func TypedValueDecode(pd paritycodec.Decoder) TypedValue {
	kind := pd.DecodeByte()
	switch kind {
	case 1:
		return I32{pd.DecodeInt32()}
	case 2:
		return I64{pd.DecodeInt64()}
	case 3:
		return F32{pd.DecodeInt32()}
	case 4:
		return F64{pd.DecodeInt64()}
	default:
		panic("Invalid TypedValue code: " + strconv.Itoa(int(kind)))
	}
}

func (t *TypedValues) ParityDecode(pd paritycodec.Decoder) {
	pd.DecodeCollection(
		func(l int) { *t = make([]TypedValue, l) },
		func(i int) {
			(*t)[i] = TypedValueDecode(pd)
		},
	)
}

type ReturnValueOrHostError interface {
	ReturnValueEncode(pe paritycodec.Encoder)
}

type Unit struct {
}

type HostError struct {
}

type TypedReturnValue struct {
	Value TypedValue
}

func (x Unit) ReturnValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(0) // Result::Value
	pe.EncodeByte(0) // ReturnValue::Unit
}

func (x HostError) ReturnValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(1) // Result::Err
}

func (x TypedReturnValue) ReturnValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(0) // Result::Value
	pe.EncodeByte(1) // ReturnValue::Value
	x.Value.TypedValueEncode(pe)
}

func ReturnValueDecode(pd paritycodec.Decoder) ReturnValueOrHostError {
	// TODO: return ReturnValue
	isValue := pd.DecodeBool()
	if isValue {
		return TypedReturnValue{TypedValueDecode(pd)}
	} else {
		return Unit{}
	}
}

const ENCODED_MAX_SIZE_RETURN_VALUE_OR_ERROR = 11

/// Describes an entity to define or import into the environment.
type ExternEntity interface {
	ExternEntityEncode(pe paritycodec.Encoder)
}

/// Function that is specified by an index in a default table of
/// a module that creates the sandbox.
type ExternFunction struct {
	Id uint32
}

func (e ExternFunction) ExternEntityEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(1)
	pe.EncodeUint32(e.Id)
}

/// Linear memory that is specified by some identifier returned by sandbox
/// module upon creation new sandboxed memory.
type ExternMemory struct {
	Id uint32
}

func (e ExternMemory) ExternEntityEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(2)
	pe.EncodeUint32(e.Id)
}

type Entry struct {
	/// Module name of which corresponding entity being defined.
	ModuleName string
	/// Field name in which corresponding entity being defined.
	FieldName string
	/// External entity being defined.
	Entity ExternEntity
}

type EnvironmentDefinition struct {
	/// Vector of all entries in the environment definition.
	Entries []Entry
}
