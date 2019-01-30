package executive

import (
	"bytes"
	"strconv"

	"github.com/Joystream/tinygo-wasm-substrate/gohelpers"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srio"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	"github.com/Joystream/tinygo-wasm-substrate/srml/support/system"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Executive struct {
	SystemModule system.Module
	Payment      srprimitives.MakePayment
	Finalization srprimitives.OnFinalise
}

type ApplyError interface {
	primitives.Enum
}

type ApplyResultCode byte

const (
	OkSuccess       ApplyResultCode = iota
	OkFail                          // has message
	ErrBadSignature                 // has message
	ErrStale
	ErrFuture
	ErrCantPay
)

/// Start the execution of a particular block.
func (e *Executive) InitialiseBlock(header *srprimitives.Header) {
	e.SystemModule.Initialise(header.Number, header.ParentHash, header.ExtrinsicsRoot)
}

func (e *Executive) InitialChecks(block *srprimitives.Block) {
	// check parent_hash is correct.
	n := block.Header.Number
	if !(n.NonZero() && e.SystemModule.BlockHashStore.Get(n.MinusOne()) == block.Header.ParentHash) {
		panic("Parent hash should be valid.")
	}

	// check transaction trie root represents the transactions.
	// TODO: support generic hashers (Rust Substrate does not yet, as of January 2019)
	r := make([]codec.Encodeable, len(block.Extrinsics))
	for i, e := range block.Extrinsics {
		r[i] = e.EncodeableEnum()
	}
	xtsRoot := srio.EnumeratedTrieRootBlake256(r)
	if *(block.Header.ExtrinsicsRoot.(*primitives.H256)) != primitives.H256(xtsRoot) {
		panic("Transaction trie root must be valid.")
	}
}

/// Actually execute all transitioning for `block`.
func (e *Executive) ExecuteBlock(block *srprimitives.Block) {
	e.InitialiseBlock(&block.Header)

	// any initial checks
	e.InitialChecks(block)

	for _, ext := range block.Extrinsics {
		e.applyExtrinsicNoNote(ext)
	}

	// post-transactional book-keeping.
	e.SystemModule.NoteFinishedExtrinsics()
	e.Finalization.OnFinalise(block.Header.Number)

	// any final checks
	e.finalChecks(block.Header)
}

/// Finalise the block - it is up the caller to ensure that all header fields are valid
/// except state-root.
func (e *Executive) FinaliseBlock() srprimitives.Header {
	e.SystemModule.NoteFinishedExtrinsics()
	e.Finalization.OnFinalise(e.SystemModule.NumberStore.Get().(srprimitives.BlockNumber))

	// setup extrinsics
	e.SystemModule.DeriveExtrinsics()
	return e.SystemModule.Finalise()
}

/// Apply extrinsic outside of the block execution function.
/// This doesn't attempt to validate anything regarding the block, but it builds a list of uxt
/// hashes.
func (e *Executive) ApplyExtrinsic(uxt srprimitives.Extrinsic) srprimitives.ApplyResult {
	encoded := codec.ToBytes(uxt.EncodeableEnum())
	encodedLen := len(encoded)
	e.SystemModule.NoteExtrinsic(encoded)
	code, _ := e.applyExtrinsicNoNoteWithLen(uxt, uintptr(encodedLen))
	switch code {
	case OkSuccess:
		return srprimitives.ApplyOutcomeSuccess
	case OkFail:
		return srprimitives.ApplyOutcomeFail
	case ErrCantPay:
		return srprimitives.ApplyErrorCantPay
	case ErrBadSignature:
		return srprimitives.ApplyErrorBadSignature
	case ErrStale:
		return srprimitives.ApplyErrorBadSignature
	case ErrFuture:
		return srprimitives.ApplyErrorFuture
	}
	panic("Unknown code: " + strconv.Itoa(int(code)))
}

/// Apply an extrinsic inside the block execution function.
func (e *Executive) applyExtrinsicNoNote(uxt srprimitives.Extrinsic) {
	encoded := codec.ToBytes(uxt.EncodeableEnum())
	encodedLen := len(encoded)
	code, msg := e.applyExtrinsicNoNoteWithLen(uxt, uintptr(encodedLen))
	switch code {
	case OkSuccess:
		return
	case OkFail:
		srio.Print(msg)
	case ErrCantPay:
		panic("All extrinsics should have sender able to pay their fees")
	case ErrBadSignature:
		panic("All extrinsics should be properly signed")
	case ErrStale, ErrFuture:
		panic("All extrinsics should have the correct nonce")
	}
}

/// Actually apply an extrinsic given its `encodedLen`; this doesn't note its hash.
func (e *Executive) applyExtrinsicNoNoteWithLen(uxt srprimitives.Extrinsic, encodedLen uintptr) (code ApplyResultCode, maybeMessage string) {
	// Verify the signature is good.
	xt, err := uxt.(srprimitives.Checkable).Check(e.SystemModule.TypeParamsFactory.DefaultContext())
	if err != nil {
		return ErrBadSignature, err.Error()
	}

	if xt.SignatureAccountID != nil {
		sender := xt.SignatureAccountID
		index := xt.SignatureIndex
		// check index
		expectedIndex := e.SystemModule.AccountNonceStore.Get(sender).(srprimitives.Index)
		if index != expectedIndex {
			if index.LessThan(expectedIndex) {
				return ErrStale, ""
			} else {
				return ErrFuture, ""
			}
		}

		// pay any fees.
		if e.Payment.MakePayment(sender, encodedLen) != nil {
			return ErrCantPay, ""
		}

		// AUDIT: Under no circumstances may this function panic from here onwards.

		// increment nonce in storage
		e.SystemModule.IncAccountNonce(sender)
	}

	// decode parameters and dispatch
	call, accountID := xt.Deconstruct()
	err = call.Dispatch(accountID)
	e.SystemModule.NoteAppliedExtrinsic(err)

	if err == nil {
		return OkSuccess, ""
	} else {
		return OkFail, err.Error()
	}
}

func (e *Executive) finalChecks(header srprimitives.Header) {
	// remove temporaries.
	newHeader := e.SystemModule.Finalise()

	// check digest.
	if len(header.Digest.Logs) != len(newHeader.Digest.Logs) {
		panic("Number of digest items must match that calculated.")
	}

	for i, headerItem := range header.Digest.Logs {
		computedItem := newHeader.Digest.Logs[i]
		gohelpers.Assert(headerItem == computedItem, "Digest item must match that calculated.")
	}

	// check storage root.
	storageRoot := srio.StorageRoot()
	gohelpers.Assert(header.StateRoot == storageRoot, "Storage root must match that calculated.")
}

/// Check a given transaction for validity. This doesn't execute any
/// side-effects; it merely checks whether the transaction would panic if it were included or not.
///
/// Changes made to the storage should be discarded.
func (e *Executive) ValidateTransaction(uxt srprimitives.Extrinsic) srprimitives.TransactionValidity {
	encoded := codec.ToBytes(uxt.EncodeableEnum())
	encodedLen := len(encoded)

	xt, err := uxt.(srprimitives.Checkable).Check(e.SystemModule.TypeParamsFactory.DefaultContext())
	if err != nil {
		if err.Error() == "invalid account index" {
			// An unknown account index implies that the transaction may yet become valid.
			return srprimitives.TransactionValidityUnknown{}
		}
		// Technically a bad signature could also imply an out-of-date account index, but
		// that's more of an edge case.
		return srprimitives.TransactionValidityInvalid{}
	}

	if xt.SignatureAccountID != nil {

		sender := xt.SignatureAccountID
		index := xt.SignatureIndex

		// pay any fees.
		if e.Payment.MakePayment(sender, uintptr(encodedLen)) != nil {
			return srprimitives.TransactionValidityInvalid{}
		}

		// check index
		expectedIndex := e.SystemModule.AccountNonceStore.Get(sender).(srprimitives.Index)
		if index.LessThan(expectedIndex) {
			return srprimitives.TransactionValidityInvalid{}
		}
		if index.GreaterThan(expectedIndex.Plus(256)) {
			return srprimitives.TransactionValidityUnknown{}
		}

		deps := []srprimitives.TransactionTag{}
		for expectedIndex.LessThan(index) {

			buffer := bytes.Buffer{}
			sender.ParityEncode(codec.Encoder{&buffer})
			expectedIndex.ParityEncode(codec.Encoder{&buffer})

			deps = append(deps, srprimitives.TransactionTag(buffer.Bytes()))
			expectedIndex = expectedIndex.Plus(1)
		}

		buffer := bytes.Buffer{}
		sender.ParityEncode(codec.Encoder{&buffer})
		index.ParityEncode(codec.Encoder{&buffer})
		prov := srprimitives.TransactionTag(buffer.Bytes())

		return srprimitives.TransactionValidityValid{
			Priority:  srprimitives.TransactionPriority(encodedLen),
			Requires:  deps,
			Provides:  []srprimitives.TransactionTag{prov},
			Longevity: srprimitives.TransactionLongevityMaxValue,
		}
	} else {
		return srprimitives.TransactionValidityInvalid{}
	}
}
