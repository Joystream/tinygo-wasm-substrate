package substratetestruntime

import (
	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

type H256 [32]byte

func (h *H256) ParityDecode(pd paritycodec.Decoder) {
	pd.Read(h[:])
}

func (h H256) ParityEncode(pe paritycodec.Encoder) {
	pe.Write(h[:])
}

type H512 [64]byte

func (h *H512) ParityDecode(pd paritycodec.Decoder) {
	pd.Read(h[:])
}

func (h H512) ParityEncode(pe paritycodec.Encoder) {
	pe.Write(h[:])
}

type AuthorityId [32]byte

type ApiID [8]byte

type ApiVersion struct {
	id      ApiID
	version uint32
}

type RuntimeVersion struct {
	/// Identifies the different Substrate runtimes. There'll be at least polkadot and node.
	/// A different on-chain spec_name to that of the native runtime would normally result
	/// in node not attempting to sync or author blocks.
	SpecName string

	/// Name of the implementation of the spec. This is of little consequence for the node
	/// and serves only to differentiate code of different implementation teams. For this
	/// codebase, it will be parity-polkadot. If there were a non-Rust implementation of the
	/// Polkadot runtime (e.g. C++), then it would identify itself with an accordingly different
	/// `impl_name`.
	ImplName string

	/// `authoring_version` is the version of the authorship interface. An authoring node
	/// will not attempt to author blocks unless this is equal to its native runtime.
	AuthoringVersion uint32

	/// Version of the runtime specification. A full-node will not attempt to use its native
	/// runtime in substitute for the on-chain Wasm runtime unless all of `spec_name`,
	/// `spec_version` and `authoring_version` are the same between Wasm and native.
	SpecVersion uint32

	/// Version of the implementation of the specification. Nodes are free to ignore this; it
	/// serves only as an indication that the code is different; as long as the other two versions
	/// are the same then while the actual code may be different, it is nonetheless required to
	/// do the same thing.
	/// Non-consensus-breaking optimisations are about the only changes that could be made which
	/// would result in only the `impl_version` changing.
	ImplVersion uint32

	/// List of supported API "features" along with their versions.
	ApiVersions []ApiVersion
}

// See generate_runtime_api_id in Rust implementation
func GenerateRuntimeApiId(name string) ApiID {
	// hash := blake2b.Sum512([]byte(name))
	var res [8]byte
	// copy(res[:], hash[:8])
	return ApiID(res)
}

func (v RuntimeVersion) ParityEncode(pe paritycodec.Encoder) {
	pe.EncodeString(v.SpecName)
	pe.EncodeString(v.ImplName)
	pe.EncodeUint(uint64(v.AuthoringVersion), 4)
	pe.EncodeUint(uint64(v.SpecVersion), 4)
	pe.EncodeUint(uint64(v.ImplVersion), 4)
}
