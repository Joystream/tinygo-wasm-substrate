package substratetestruntime

import (
	"strconv"

	paritycodec "github.com/kyegupov/parity-codec-go/noreflect"
)

// TODO:
// (export "Core_version" (func $Core_version))
// (export "Core_authorities" (func $Core_authorities))
// (export "Core_execute_block" (func $Core_execute_block))
// (export "Core_initialise_block" (func $Core_initialise_block))
// (export "Metadata_metadata" (func $Metadata_metadata))
// (export "TaggedTransactionQueue_validate_transaction" (func $TaggedTransactionQueue_validate_transaction))
// (export "BlockBuilder_apply_extrinsic" (func $BlockBuilder_apply_extrinsic))
// (export "BlockBuilder_finalise_block" (func $BlockBuilder_finalise_block))
// (export "BlockBuilder_inherent_extrinsics" (func $BlockBuilder_inherent_extrinsics))
// (export "BlockBuilder_check_inherents" (func $BlockBuilder_check_inherents))
// (export "BlockBuilder_random_seed" (func $BlockBuilder_random_seed))
// (export "TestAPI_balance_of" (func $TestAPI_balance_of))
// (export "AuraApi_slot_duration" (func $AuraApi_slot_duration))

//go:export Core_version
func coreVersion(_ *byte, _ uint32) uint64 {
	return returnSlice(paritycodec.Encode(
		RuntimeVersion{
			"test",
			"parity-test",
			1,
			1,
			1,
			[]ApiVersion{},
		}))
}

type AccountId H256

type Transfer struct {
	from   AccountId
	to     AccountId
	amount uint64
	nonce  uint64
}

func (t *Transfer) ParityDecode(pd paritycodec.Decoder) {
	print("Transfer")
	(*H256)(&t.from).ParityDecode(pd)
	print("Transfer to")
	(*H256)(&t.to).ParityDecode(pd)
	print("Transfer amo")
	t.amount = pd.DecodeUint(8)
	t.nonce = pd.DecodeUint(8)
}

func (t *Transfer) ParityEncode(pe paritycodec.Encoder) {
	(H256)(t.from).ParityEncode(pe)
	(H256)(t.to).ParityEncode(pe)
	pe.EncodeUint(t.amount, 8)
	pe.EncodeUint(t.nonce, 8)
}

type Ed25519Signature H512

type Extrinsic struct {
	transfer  Transfer
	signature Ed25519Signature
}

func (e *Extrinsic) ParityDecode(pd paritycodec.Decoder) {
	print("Extrinsic")
	e.transfer.ParityDecode(pd)
	print("Extrinsic sig")
	(*H512)(&e.signature).ParityDecode(pd)
}

func (e Extrinsic) ParityEncode(pe paritycodec.Encoder) {
	e.transfer.ParityEncode(pe)
	(H512)(e.signature).ParityEncode(pe)
}

type BlockNumber uint64

type HashOutput H256 // BlakeTwo256::Output

type Header struct {
	/// The parent hash.
	parentHash HashOutput
	/// The block number.
	number BlockNumber
	/// The state trie merkle root
	stateRoot HashOutput
	/// The merkle root of the extrinsics.
	extrinsicsRoot HashOutput
	/// A chain-specific digest of data useful for light clients or referencing auxiliary data.
	digest Digest
}

func (h *Header) ParityDecode(pd paritycodec.Decoder) {
	print("Header 1")
	(*H256)(&h.parentHash).ParityDecode(pd)
	print("Header 2")
	h.number = BlockNumber(pd.DecodeUintCompact())
	print("Header 3")
	(*H256)(&h.stateRoot).ParityDecode(pd)
	print("Header 4")
	(*H256)(&h.extrinsicsRoot).ParityDecode(pd)
	print("Header 5")
	(&h.digest).ParityDecode(pd)
}

type Extrinsics []Extrinsic

func (e *Extrinsics) ParityDecode(pd paritycodec.Decoder) {
	print("Extrinsics")
	pd.DecodeCollection(
		func(n int) { *e = make([]Extrinsic, n) },
		func(i int) { (&(*e)[i]).ParityDecode(pd) },
	)
}

func (e Extrinsics) ParityEncode(pe paritycodec.Encoder) {
	pe.EncodeCollection(
		len(e),
		func(i int) { e[i].ParityEncode(pe) },
	)
}

type Block struct {
	header     Header
	extrinsics Extrinsics
}

func (b *Block) ParityDecode(pd paritycodec.Decoder) {
	print("Block")
	b.header.ParityDecode(pd)
	b.extrinsics.ParityDecode(pd)
}

//go:export Core_execute_block
func executeBlock(offset *byte, length uintptr) uint64 {
	print(strconv.Itoa(int(length)))
	block := Block{}
	mr := NewMemReader(offset, length)
	pd := paritycodec.Decoder{&mr}
	block.ParityDecode(pd)
	// header := block.header

	// check transaction trie root represents the transactions.
	// txs := make([][]byte, len(block.extrinsics))
	// for i, e := range block.extrinsics {
	// 	txs[i] = paritycodec.Encode(e)
	// }

	// txsRoot := enumeratedTrieRootBlake256ForByteSlices(txs)
	// if txsRoot != header.extrinsicsRoot {
	// 	panic("Transaction trie root must be valid.")
	// }
	// info_expect_equal_hash(&txs_root, &header.extrinsics_root)
	// assert!(txs_root == header.extrinsics_root, )

	// // execute transactions
	// block.extrinsics.iter().enumerate().for_each(|(i, e)| {
	// 	storage::unhashed::put(well_known_keys::EXTRINSIC_INDEX, &(i as u32))
	// 	execute_transaction_backend(e).map_err(|_| ()).expect("Extrinsic error")
	// 	storage::unhashed::kill(well_known_keys::EXTRINSIC_INDEX)
	// })

	// // check storage root.
	// let storage_root = storage_root().into()
	// info_expect_equal_hash(&storage_root, &header.state_root)
	// assert!(storage_root == header.state_root, "Storage root must match that calculated.")

	// // check digest
	// let mut digest = Digest::default()
	// if let Some(storage_changes_root) = storage_changes_root(header.parent_hash.into(), header.number - 1) {
	// 	digest.push(generic::DigestItem::ChangesTrieRoot::<Hash, u64>(storage_changes_root.into()))
	// }
	// assert!(digest == header.digest, "Header digest items must match that calculated.")
	return 0
}

func main() {

}
