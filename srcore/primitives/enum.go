package primitives

import (
	"strconv"

	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

// Rust enum
type Enum interface {
	EncodeableEnum() EncodeableEnum
}

func InvalidEnum(b byte, typ string) string {
	return "Invalid enum value for type " + typ + ": " + strconv.Itoa(int(b))
}

type EncodeableEnum struct {
	Kind    byte
	Payload codec.Encodeable
}

func (e EncodeableEnum) ParityEncode(pe codec.Encoder) {
	pe.EncodeByte(e.Kind)
	e.Payload.ParityEncode(pe)
}

type NoPayload struct{}

func (_ NoPayload) ParityEncode(codec.Encoder) {}
