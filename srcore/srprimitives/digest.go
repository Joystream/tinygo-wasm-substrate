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
	Logs       []DigestItem
	AidFactory func() AuthorityId // pseudo-generic
}

func (d *Digest) ParityDecode(pd paritycodec.Decoder) {
	pd.DecodeCollection(
		func(n int) { d.Logs = make([]DigestItem, n) },
		func(i int) { d.Logs[i] = DecodeDigestItem(pd, d.AidFactory) },
	)
}

type DigestItemType int

const (
	DtOther             DigestItemType = 0
	DtAuthoritiesChange DigestItemType = 1
	DtChangesTrieRoot   DigestItemType = 2
	DtSeal              DigestItemType = 3
)

type DigestItem interface {
	ImplementsDigestItem() DigestItemType
}

func DecodeDigestItem(pd paritycodec.Decoder, aidFactory func() AuthorityId) DigestItem {
	t := DigestItemType(pd.ReadOneByte())
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

func (di AuthoritiesChange) ImplementsDigestItem() DigestItemType { return DtAuthoritiesChange }

/// System digest item that contains the root of changes trie at given
/// block. It is created for every block iff runtime supports changes
/// trie creation.
type ChangesTrieRoot primitives.H256

func (di ChangesTrieRoot) ImplementsDigestItem() DigestItemType { return DtChangesTrieRoot }

type Signature primitives.H512

/// Put a Seal on it
type Seal struct {
	number    uint64 // ?????
	signature Signature
}

func (s *Seal) ParityDecode(pd paritycodec.Decoder) {
	s.number = pd.DecodeUint(8)
	(*primitives.H512)(&s.signature).ParityDecode(pd)
}

func (di Seal) ImplementsDigestItem() DigestItemType { return DtSeal }

/// Any 'non-system' digest item, opaque to the native code.
type OtherDigestItem []byte

func (di OtherDigestItem) ImplementsDigestItem() DigestItemType { return DtOther }
