package srprimitives

import (
	"strconv"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

type AuthorityId interface {
	ParityDecode(decoder paritycodec.Decoder)
}

type Digest struct {
	Logs []DigestItem
}

func (d Digest) ParityEncode(pe paritycodec.Encoder) {
	pe.EncodeCollection(
		len(d.Logs),
		func(i int) {
			pe.EncodeByte(byte(d.Logs[i].DigestItemType()))
			panic("Digest item encoding not implemented")
			// TODO: d.Logs[i].ParityEncode(pe)
		},
	)
}

func (d *Digest) ParityDecodeDigest(pd paritycodec.Decoder, itemDecoder func(paritycodec.Decoder) DigestItem) {
	pd.DecodeCollection(
		func(n int) { d.Logs = make([]DigestItem, n) },
		func(i int) { d.Logs[i] = itemDecoder(pd) },
	)
}

type DigestItemType byte

const (
	DtOther             DigestItemType = 0
	DtAuthoritiesChange DigestItemType = 1
	DtChangesTrieRoot   DigestItemType = 2
	DtSeal              DigestItemType = 3
)

type DigestItem interface {
	// paritycodec.Encodeable
	DigestItemType() DigestItemType
}

// Below are subtypes of of "enum DigestItem", the default digest item type

/// Digest item that is able to encode/decode 'system' digest items and
/// provide opaque access to other items.

func DigestItemFactory(aidFactory func() AuthorityId) func(paritycodec.Decoder) DigestItem {
	return func(pd paritycodec.Decoder) DigestItem { return DecodeDigestItem(pd, aidFactory) }
}

func DecodeDigestItem(pd paritycodec.Decoder, aidFactory func() AuthorityId) DigestItem {
	t := DigestItemType(pd.DecodeByte())
	switch t {
	case DtOther:
		return OtherDigestItem(pd.DecodeByteSlice())
	case DtAuthoritiesChange:
		ac := AuthoritiesChange{}
		ac.ParityDecode(pd, aidFactory)
		return ac
	case DtChangesTrieRoot:
		ctr := ChangesTrieRoot{}
		(*primitives.H256)(&ctr).ParityDecode(pd)
		return ctr
	case DtSeal:
		seal := Seal{}
		seal.ParityDecode(pd)
		return seal
	default:
		panic("Unknown digest item type: " + strconv.Itoa(int(t)))
	}
}

/// System digest item announcing that authorities set has been changed
/// in the block. Contains the new set of authorities.
type AuthoritiesChange []AuthorityId

func (a *AuthoritiesChange) ParityDecode(pd paritycodec.Decoder, aidFactory func() AuthorityId) {
	pd.DecodeCollection(
		func(n int) { *a = make([]AuthorityId, n) },
		func(i int) { (*a)[i] = aidFactory(); (*a)[i].ParityDecode(pd) },
	)
}

func (di AuthoritiesChange) DigestItemType() DigestItemType { return DtAuthoritiesChange }

/// System digest item that contains the root of changes trie at given
/// block. It is created for every block iff runtime supports changes
/// trie creation.
/// TODO: *H256
type ChangesTrieRoot primitives.H256

func (di ChangesTrieRoot) DigestItemType() DigestItemType { return DtChangesTrieRoot }

// end "enum DigestItem"

type Signature primitives.H512

/// Put a Seal on it
type Seal struct {
	number    uint64 // ?????
	signature Signature
}

func (s *Seal) ParityDecode(pd paritycodec.Decoder) {
	s.number = pd.DecodeUint64()
	(*primitives.H512)(&s.signature).ParityDecode(pd)
}

func (di Seal) DigestItemType() DigestItemType { return DtSeal }

/// Any 'non-system' digest item, opaque to the native code.
type OtherDigestItem []byte

func (di OtherDigestItem) DigestItemType() DigestItemType { return DtOther }
