package srprimitives

import (
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type BlockNumber interface {
	codec.Encodeable
	codec.Decodeable
	// TODO: support 128+ bits
	AsUint64() uint64
	MinusOne() BlockNumber
	NonZero() bool
	GreaterThan(BlockNumber) bool
}

// TODO: unify with HashOutput
type Hash interface {
	codec.Encodeable
	codec.Decodeable
}

type HeaderTypeParamsFactory interface {
	NewHashOutput() HashOutput
	BlockNumber(uint64) BlockNumber
	DecodeDigestItem(pd codec.Decoder) DigestItem
}

type Header struct {
	/// The parent hash.
	ParentHash HashOutput
	/// The block number.
	Number BlockNumber
	/// The state trie merkle root
	StateRoot HashOutput
	/// The merkle root of the extrinsics.
	ExtrinsicsRoot HashOutput
	/// A chain-specific digest of data useful for light clients or referencing auxiliary data.
	Digest Digest
}

func (h *Header) ParityEncode(pe codec.Encoder) {
	h.ParentHash.ParityEncode(pe)
	pe.EncodeUintCompact(h.Number.AsUint64())
	h.StateRoot.ParityEncode(pe)
	h.ExtrinsicsRoot.ParityEncode(pe)
	h.Digest.ParityEncode(pe)
}

func (h *Header) ParityDecode(pd codec.Decoder, types HeaderTypeParamsFactory) {
	h.ParentHash = types.NewHashOutput()
	h.ParentHash.ParityDecode(pd)
	h.Number = types.BlockNumber(pd.DecodeUintCompact())
	h.StateRoot = types.NewHashOutput()
	h.StateRoot.ParityDecode(pd)
	h.ExtrinsicsRoot = types.NewHashOutput()
	h.ExtrinsicsRoot.ParityDecode(pd)
	h.Digest.ParityDecodeDigest(pd, types.DecodeDigestItem)
}
