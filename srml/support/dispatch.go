package support

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type BaseModule struct {

	// Order is important, because it determines the encoding of the "module"
	methods []srprimitives.Callable
}

type Module interface {
	InitForRuntime(TypeParamsFactory)
}

func (m *BaseModule) AddMethod(c srprimitives.Callable) {
	m.methods = append(m.methods, c)
}

// Corresponds to "Call" enum generated for Rust modules
type ModuleCall struct {
	methodIndex int
	method      srprimitives.Callable
}

func (c ModuleCall) Dispatch(o srprimitives.Origin) error {
	return c.method.Dispatch(o)
}

func (c ModuleCall) ParityEncode(pe codec.Encoder) {
	pe.EncodeByte(byte(c.methodIndex))
	c.method.ParityEncode(pe)
}

// Every Runtime should implement this
type TypeParamsFactory interface {
	NewHash(byte) srprimitives.HashOutput
	BlockNumber(uint64) srprimitives.BlockNumber
	DecodeDigestItem(pd codec.Decoder) srprimitives.DigestItem
	ZeroIndex() srprimitives.Index
	DecodeEvent(pd codec.Decoder) Event
	DefaultContext() interface{}
}

// Since implementations of type parameters might differ from runtime to runtime,
// a TypeParamsFactory is needed to instantiate values of these types.
// Thus, storage requires the stored types to implement this interface instead of
// codec.Decodeable
type DecodeableWithParams interface {
	ParityDecode(codec.Decoder, TypeParamsFactory)
}
