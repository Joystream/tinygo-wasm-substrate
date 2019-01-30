package support

import (
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Event RawEvent

type RawEvent interface {
	codec.Encodeable
}
