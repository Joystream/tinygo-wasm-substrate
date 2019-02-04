package srprimitives

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Origin interface{}

// See support/dispatch
type Callable interface {
	primitives.Enum
	Dispatch(Origin) error
}

// Effectively, some fixed-size array.
// Current substrate implementation seems to only use H256
// (for BlakeTwo256 hash)
type HashOutput interface {
	codec.Encodeable
	codec.Decodeable
	AsBytes() []byte
}

// TODO: properly implement https://docs.rs/safe-mix/1.0.0/safe_mix/
func TripletMix(hashes []HashOutput) HashOutput {
	return hashes[0]
}

// We do not have actual Hash trait or implementation defined here,
// since the hashing is not to be performed in the module

type Checked interface {
}

type Checkable interface {
	/// Check self, given an instance of Context.
	Check(context interface{}) (CheckedExtrinsic, error)
}

type OnFinalise interface {
	/// The block is being finalised. Implement to have something happen.
	OnFinalise(n BlockNumber)
}

type AccountId interface {
	codec.Encodeable
}

type MakePayment interface {
	/// Make some sort of payment concerning `who` for an extrinsic (transaction) of encoded length
	/// `encoded_len` bytes. Return true iff the payment was successful.
	MakePayment(who AccountId, encodedLen uintptr) error
}

/// An "executable" piece of information, used by the standard Substrate Executive in order to
/// enact a piece of extrinsic information by marshalling and dispatching to a named functioon
/// call.
///
/// Also provides information on to whom this information is attributable and an index that allows
/// each piece of attributable information to be disambiguated.
type Applyable interface {
	Index() Index                   // can be nil
	Sender() AccountId              // can be nil
	Deconstruct() (Call, AccountId) // can both be nil
}
