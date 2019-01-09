package srprimitives

import "github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
import "github.com/Joystream/tinygo-wasm-substrate/srcore/srio"
import . "github.com/Joystream/tinygo-wasm-substrate/wasmhelpers"

type Ed25519Signature primitives.H512

func (s Ed25519Signature) Verify(message []byte, signer primitives.H256) {
	errCode := srio.Ext_ed25519_verify(GetOffset(message), GetLen(message), &s[0], &signer[0])
	if errCode != 0 {
		panic("Invalid signature")
	}
}
