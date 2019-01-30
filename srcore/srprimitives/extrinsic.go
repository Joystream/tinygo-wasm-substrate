package srprimitives

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	"github.com/Joystream/tinygo-wasm-substrate/srml/indices"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Extrinsic interface {
	primitives.Enum
	IsSigned() (bool, bool)

	GetFunction() Callable // All UncheckedExtrinsic
}

/// Index of a transaction in the chain.
type Index interface {
	codec.Encodeable
	codec.Decodeable
	PlusOne() Index
	Plus(i int) Index
	LessThan(o Index) bool
	GreaterThan(o Index) bool
}

type SignatureContent struct {
	signed    indices.Address
	signature Signature
	index     Index
}

// Default implementation
// TODO: also mortal ones
type UncheckedExtrinsic struct {
	/// The signature, address and number of extrinsics have come before from
	/// the same signer, if this is a signed extrinsic.
	HasSignature bool
	Signature    SignatureContent
	/// The function that should be called.
	Function Callable
}

func (e *UncheckedExtrinsic) IsSigned() (bool, bool) {
	// TODO: proper impl
	return e.HasSignature, e.Signature.signed != nil
}

// TODO: encoding
func (e *UncheckedExtrinsic) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{}
}

func (e *UncheckedExtrinsic) GetFunction() Callable {
	return e.Function
}

type CheckedExtrinsic struct {
	SignatureAccountID AccountId // nil if unsigned
	SignatureIndex     Index     // nil if unsigned
	Function           Callable
}

// Implements Applyable

func (c *CheckedExtrinsic) Index() Index {
	return c.SignatureIndex
}
func (c *CheckedExtrinsic) Sender() AccountId {
	return c.SignatureAccountID
}
func (c *CheckedExtrinsic) Deconstruct() (Callable, AccountId) {
	return c.Function, c.SignatureAccountID
}
