package substratetestruntime

type Ed25519Signature H512

func (s Ed25519Signature) Verify(message []byte, signer H256) {
	errCode := ext_ed25519_verify(getOffset(message), getLen(message), &s[0], &signer[0])
	if errCode != 0 {
		panic("Invalid signature")
	}
}
