package substratetestruntime

import (
	"strconv"

	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Digest struct {
	logs []DigestItem
}

func (d *Digest) ParityDecode(pd paritycodec.Decoder) {
	pd.DecodeCollection(
		func(n int) { d.logs = make([]DigestItem, n) },
		func(i int) { d.logs[i] = DecodeDigestItem(pd) },
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

func DecodeDigestItem(pd paritycodec.Decoder) DigestItem {
	t := DigestItemType(pd.ReadOneByte())
	switch t {
	case DtOther:
		return OtherDigestItem(pd.DecodeByteSlice())
	case DtAuthoritiesChange:
		ac := AuthoritiesChange{}
		ac.ParityDecode(pd)
		return ac
	case DtChangesTrieRoot:
		ctr := ChangesTrieRoot{}
		(*H256)(&ctr).ParityDecode(pd)
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

func (a *AuthoritiesChange) ParityDecode(pd paritycodec.Decoder) {
	pd.DecodeCollection(
		func(n int) { *a = make([]AuthorityId, n) },
		func(i int) { (*H256)(&(*a)[i]).ParityDecode(pd) },
	)
}

func (di AuthoritiesChange) ImplementsDigestItem() DigestItemType { return DtAuthoritiesChange }

/// System digest item that contains the root of changes trie at given
/// block. It is created for every block iff runtime supports changes
/// trie creation.
type ChangesTrieRoot H256

func (di ChangesTrieRoot) ImplementsDigestItem() DigestItemType { return DtChangesTrieRoot }

type Signature H512

/// Put a Seal on it
type Seal struct {
	number    uint64 // ?????
	signature Signature
}

func (s *Seal) ParityDecode(pd paritycodec.Decoder) {
	s.number = pd.DecodeUint(8)
	(*H512)(&s.signature).ParityDecode(pd)
}

func (di Seal) ImplementsDigestItem() DigestItemType { return DtSeal }

/// Any 'non-system' digest item, opaque to the native code.
type OtherDigestItem []byte

func (di OtherDigestItem) ImplementsDigestItem() DigestItemType { return DtOther }
