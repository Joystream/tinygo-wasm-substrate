package gohelpers

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
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

type Uint64 uint64

func (v *Uint64) ParityEncode(pe codec.Encoder) {
	pe.EncodeUint64(uint64(*v))
}

func (v *Uint64) ParityDecode(pd codec.Decoder) {
	*v = Uint64(pd.DecodeUint64())
}

func (v *Uint64) AsUint64() uint64 {
	return uint64(*v)
}

func (u *Uint64) MinusOne() srprimitives.BlockNumber {
	v := Uint64(uint64(*u) - 1)
	return &v
}

func (u *Uint64) NonZero() bool {
	return uint64(*u) > 0
}

func (u *Uint64) GreaterThan(o srprimitives.BlockNumber) bool {
	return uint64(*u) > uint64(*o.(*Uint64))
}
