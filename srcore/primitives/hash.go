package primitives

import (
	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

type H256 [32]byte

func (h *H256) ParityDecode(pd paritycodec.Decoder) {
	pd.Read(h[:])
}

func (h *H256) AsBytes() []byte {
	return h[:]
}

func (h *H256) ParityEncode(pe paritycodec.Encoder) {
	pe.Write(h[:])
}

type H512 [64]byte

func (h *H512) ParityDecode(pd paritycodec.Decoder) {
	pd.Read(h[:])
}

func (h *H512) ParityEncode(pe paritycodec.Encoder) {
	pe.Write(h[:])
}
