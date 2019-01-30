package gohelpers

import (
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Uint32 uint32

func (v Uint32) ParityEncode(pe codec.Encoder) {
	pe.EncodeUint32(uint32(v))
}

type ByteSlice []byte

func (b ByteSlice) ParityEncode(pe codec.Encoder) {
	pe.EncodeByteSlice(b)
}
