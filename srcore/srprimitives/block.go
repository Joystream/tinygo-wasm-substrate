package srprimitives

import codec "github.com/kyegupov/parity-codec-go/noreflect"

type Block struct {
	Header     Header
	Extrinsics []Extrinsic
}

type BlockTypeParamsFactory interface {
	HeaderTypeParamsFactory
	DecodeExtrinsic(pd codec.Decoder) Extrinsic
}

func (b *Block) ParityDecode(pd codec.Decoder, types BlockTypeParamsFactory) {
	b.Header.ParityDecode(pd, types)
	pd.DecodeCollection(
		func(n int) { b.Extrinsics = make([]Extrinsic, n) },
		func(i int) { b.Extrinsics[i] = types.DecodeExtrinsic(pd) },
	)
}

type BlockId interface {
	ImplementsBlockId()
}

// type HashBlockId struct {
// 	hash HashOutput
// }

// func (bid *HashBlockId) ImplementsBlockId() {}

// type NumberBlockId struct {
// 	number Number
// }

// func (bid *NumberBlockId) ImplementsBlockId() {}
