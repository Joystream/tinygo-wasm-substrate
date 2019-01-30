package support

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Module interface {
	InitForRuntime(TypeParamsFactory)
	CallableBelongsToThisModule(srprimitives.Callable) bool
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
