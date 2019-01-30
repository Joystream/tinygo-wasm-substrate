package primitives

import (
	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Ed25519AuthorityId [32]byte

func (a *Ed25519AuthorityId) ParityDecode(pd paritycodec.Decoder) {
	pd.Read(a[:])
}
