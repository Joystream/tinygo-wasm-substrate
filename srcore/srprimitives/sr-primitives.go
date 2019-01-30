package srprimitives

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
)

type ApplyResult interface {
	primitives.Result
	ImplementsApplyResult()
}

type ApplyOutcome byte

func (_ ApplyOutcome) ImplementsApplyResult() {}
func (_ ApplyOutcome) IsError() bool          { return false }

type ApplyError byte

func (_ ApplyError) ImplementsApplyResult() {}
func (_ ApplyError) IsError() bool          { return true }

const (
	/// Successful application (extrinsic reported no issue).
	ApplyOutcomeSuccess ApplyOutcome = 0
	/// Failed application (extrinsic was probably a no-op other than fees).
	ApplyOutcomeFail ApplyOutcome = 1
	/// Bad signature.
	ApplyErrorBadSignature ApplyError = 0
	/// Nonce too low.
	ApplyErrorStale ApplyError = 1
	/// Nonce too high.
	ApplyErrorFuture ApplyError = 2
	/// Sending account had too low a balance.
	ApplyErrorCantPay ApplyError = 3
)
