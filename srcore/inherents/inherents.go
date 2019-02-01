package inherents

import (
	"bytes"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	runtimemodule "github.com/Joystream/tinygo-wasm-substrate/srml/support/runtime"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

/// An identifier for an inherent.
type InherentIdentifier [8]byte

/// Inherent data to include in a block.
type InherentData struct {
	/// All inherent data encoded with parity-codec and an identifier.
	data map[InherentIdentifier][]byte
}

/// Put data for an inherent into the internal storage.
///
/// # Return
///
/// Returns `Ok(())` if the data could be inserted an no data for an inherent with the same
/// identifier existed, otherwise an error is returned.
///
/// Inherent identifiers need to be unique, otherwise decoding of these values will not work!
func (i *InherentData) PutData(
	identifier InherentIdentifier,
	inherent codec.Encodeable,
) {
	_, has := i.data[identifier]
	if has {
		panic("Inherent with same identifier already exists!")
	}
	i.data[identifier] = codec.ToBytes(inherent)
}

/// Replace the data for an inherent.
///
/// If it does not exist, the data is just inserted.
func (i *InherentData) ReplaceData(
	identifier InherentIdentifier,
	inherent codec.Encodeable,
) {
	i.data[identifier] = codec.ToBytes(inherent)
}

/// Returns the data for the requested inherent.
///
func (i *InherentData) GetData(
	identifier InherentIdentifier,
	decoder func(codec.Decoder) interface{},
) (bool, interface{}) {
	valBytes, ok := i.data[identifier]
	if ok {
		return true, decoder(codec.Decoder{bytes.NewBuffer(valBytes)})
	}
	return false, nil
}

/// Did we encounter a fatal error while checking an inherent?
///
/// A fatal error is everything that fails while checking an inherent error, e.g. the inherent
/// was not found, could not be decoded etc.
/// Then there are cases where you not want the inherent check to fail, but report that there is an
/// action required. For example a timestamp of a block is in the future, the timestamp is still
/// correct, but it is required to verify the block at a later time again and then the inherent
/// check will succeed.
type IsFatalError interface {
	/// Is this a fatal error?
	IsFatalError() bool
}

type CheckInherentError interface {
	IsFatalError
	codec.Encodeable
}

type ProvideInherent interface {
	InherentIdentifier() InherentIdentifier
	CreateInherent(*InherentData) srprimitives.Callable                    // nillable
	CheckInherent(srprimitives.Callable, *InherentData) CheckInherentError // nillable
}

type CheckInherentsResult struct {
	/// Did the check succeed?
	Okay bool
	/// Did we encounter a fatal error?
	FatalError bool
	/// We use the `InherentData` to store our errors.
	Errors InherentData
}

func NewCheckInherentsResult() CheckInherentsResult {
	return CheckInherentsResult{Okay: true, Errors: InherentData{data: make(map[InherentIdentifier][]byte)}}
}

/// Put an error into the result.
///
/// This makes this result resolve to `ok() == false`.
///
/// # Parameters
///
/// - identifier - The identifier of the inherent that generated the error.
/// - error - The error that will be encoded.
func (c *CheckInherentsResult) PutError(identifier InherentIdentifier, err CheckInherentError) {
	// Don't accept any other error
	if c.FatalError {
		panic("No other errors are accepted after an hard error!")
	}

	if err.IsFatalError() {
		c.Errors.data = make(map[InherentIdentifier][]byte)
	}

	c.Errors.PutData(identifier, err)

	c.Okay = false
	c.FatalError = err.IsFatalError()
}

func (i *InherentData) CreateExtrinsics(runtime *runtimemodule.Runtime) []srprimitives.Extrinsic {
	inherents := []srprimitives.Extrinsic{}
	for _, m := range runtime.Modules {
		mi, ok := m.Module.(ProvideInherent)
		if ok {
			inherent := mi.CreateInherent(i)
			if inherent != nil {
				// TODO: figure out how wrapping into a Call works
				inherents = append(inherents, &srprimitives.UncheckedExtrinsic{Function: inherent})
			}
		}
	}
	return inherents
}

func (i *InherentData) CheckExtrinsics(runtime *runtimemodule.Runtime, block srprimitives.Block) CheckInherentsResult {

	result := NewCheckInherentsResult()
	for _, xt := range block.Extrinsics {
		has, signed := xt.IsSigned()
		if has && signed {
			break
		}

		function := xt.GetFunction().(runtimemodule.RuntimeCall)
		m := runtime.ModuleForCall(function)
		mi, ok := m.(ProvideInherent)
		if ok {
			err := mi.CheckInherent(function, i)
			if err != nil {
				result.PutError(mi.InherentIdentifier(), err) // Panic if more than one fatal error
			}
			if err.IsFatalError() {
				return result
			}
		}
	}
	return result
}
