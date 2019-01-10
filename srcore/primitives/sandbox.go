package primitives

import paritycodec "github.com/kyegupov/parity-codec-go/noreflect"

// Matches wasm basic types. Used for arguments in sandbox function calls.
type TypedValue interface {
	TypedValueEncode(paritycodec.Encoder)
}

type I32 struct{ v int32 }
type I64 struct{ v int64 }
type F32 struct{ v int32 } // TODO: do we really need float
type F64 struct{ v int64 }

func (x I32) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(1)
	pe.EncodeInt32(x.v)
}
func (x I32) ImplementsReturnValue() {}

func (x I64) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(2)
	pe.EncodeInt64(x.v)
}
func (x I64) ImplementsReturnValue() {}

func (x F32) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(3)
	pe.EncodeInt32(x.v)
}
func (x F32) ImplementsReturnValue() {}

func (x F64) TypedValueEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(4)
	pe.EncodeInt64(x.v)
}
func (x F64) ImplementsReturnValue() {}

type TypedValues []TypedValue

func (t TypedValues) ParityEncode(pe paritycodec.Encoder) {
	pe.EncodeCollection(len(t), func(i int) { t[i].TypedValueEncode(pe) })
}

func (t *TypedValues) ParityDecode(pd paritycodec.Decoder) {
	pd.DecodeCollection(
		func(l int) { *t = make([]TypedValue, l) },
		func(i int) {
			switch pd.DecodeByte() {
			case 1:
				(*t)[i] = I32{pd.DecodeInt32()}
			case 2:
				(*t)[i] = I64{pd.DecodeInt64()}
			case 3:
				(*t)[i] = F32{pd.DecodeInt32()}
			case 4:
				(*t)[i] = F64{pd.DecodeInt64()}
			}
		},
	)
}

type ReturnValue struct {
	hasValue   bool
	typedValue TypedValue
}

func (r ReturnValue) ParityEncode(pe paritycodec.Encoder) {
	pe.EncodeBool(r.hasValue)
	if r.hasValue {
		r.typedValue.TypedValueEncode(pe)
	}
}

type HostError struct {
}

/// Describes an entity to define or import into the environment.
type ExternEntity interface {
	ExternEntityEncode(pe paritycodec.Encoder)
}

/// Function that is specified by an index in a default table of
/// a module that creates the sandbox.
type Function struct {
	Id uint32
}

func (e Function) ExternEntityEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(1)
	pe.EncodeUint32(e.Id)
}

/// Linear memory that is specified by some identifier returned by sandbox
/// module upon creation new sandboxed memory.
type Memory struct {
	Id uint32
}

func (e Memory) ExternEntityEncode(pe paritycodec.Encoder) {
	pe.EncodeByte(2)
	pe.EncodeUint32(e.Id)
}

type Entry struct {
	/// Module name of which corresponding entity being defined.
	ModuleName []byte
	/// Field name in which corresponding entity being defined.
	FieldName []byte
	/// External entity being defined.
	Entity ExternEntity
}

type EnvironmentDefinition struct {
	/// Vector of all entries in the environment definition.
	Entries []Entry
}
