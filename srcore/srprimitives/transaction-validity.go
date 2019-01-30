package srprimitives

import (
	"math"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type TransactionValidity interface {
	primitives.Enum
	ImplementsTransactionValidity()
}

type TransactionValidityInvalid struct{}

func (_ TransactionValidityInvalid) ImplementsTransactionValidity()
func (_ TransactionValidityInvalid) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{0, primitives.NoPayload{}}
}

/// Priority for a transaction. Additive. Higher is better.
type TransactionPriority uint64

/// Minimum number of blocks a transaction will remain valid for.
/// `TransactionLongevity::max_value()` means "forever".
type TransactionLongevity uint64

var TransactionLongevityMaxValue TransactionLongevity = math.MaxUint64

/// Tag for a transaction. No two transactions with the same tag should be placed on-chain.
type TransactionTag []byte

type TransactionValidityValid struct {
	/// Priority of the transaction.
	///
	/// Priority determines the ordering of two transactions that have all
	/// their dependencies (required tags) satisfied.
	Priority TransactionPriority
	/// Transaction dependencies
	///
	/// A non-empty list signifies that some other transactions which provide
	/// given tags are required to be included before that one.
	Requires []TransactionTag
	/// Provided tags
	///
	/// A list of tags this transaction provides. Successfuly importing the transaction
	/// will enable other transactions that depend on (require) those tags to be included as well.
	/// Provided and requried tags allow Substrate to build a dependency graph of transactions
	/// and import them in the right (linear) order.
	Provides []TransactionTag
	/// Transaction longevity
	///
	/// Longevity describes minimum number of blocks the validity is correct.
	/// After this period transaction should be removed from the pool or revalidated.
	Longevity TransactionLongevity
}

func (_ TransactionValidityValid) ImplementsTransactionValidity()
func (t TransactionValidityValid) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{1, t}
}
func (t TransactionValidityValid) ParityEncode(pe codec.Encoder) {
	pe.EncodeUint64(uint64(t.Priority))
	pe.EncodeCollection(len(t.Requires), func(i int) { pe.EncodeByteSlice(t.Requires[i]) })
	pe.EncodeCollection(len(t.Provides), func(i int) { pe.EncodeByteSlice(t.Provides[i]) })
	pe.EncodeUint64(uint64(t.Longevity))
}

type TransactionValidityUnknown struct{}

func (_ TransactionValidityUnknown) ImplementsTransactionValidity()
func (_ TransactionValidityUnknown) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{2, primitives.NoPayload{}}
}
