package main

import (
	"bytes"
	"strconv"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srio"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srversion"
	. "github.com/Joystream/tinygo-wasm-substrate/wasmhelpers"
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
	return ReturnSlice(paritycodec.Encode(
		srversion.RuntimeVersion{
			"test",
			"parity-test",
			1,
			1,
			1,
			[]srversion.ApiVersion{},
		}))
}

type AuthorityId [32]byte

func (t *AuthorityId) ParityDecode(pd paritycodec.Decoder) {
	(*primitives.H256)(t).ParityDecode(pd)
}

type AccountId primitives.H256

type Transfer struct {
	from   AccountId
	to     AccountId
	amount uint64
	nonce  uint64
}

func (t *Transfer) ParityDecode(pd paritycodec.Decoder) {
	(*primitives.H256)(&t.from).ParityDecode(pd)
	(*primitives.H256)(&t.to).ParityDecode(pd)
	t.amount = pd.DecodeUint(8)
	t.nonce = pd.DecodeUint(8)
}

func (t Transfer) ParityEncode(pe paritycodec.Encoder) {
	(primitives.H256)(t.from).ParityEncode(pe)
	(primitives.H256)(t.to).ParityEncode(pe)
	pe.EncodeUint(t.amount, 8)
	pe.EncodeUint(t.nonce, 8)
}

type Extrinsic struct {
	transfer  Transfer
	signature srprimitives.Ed25519Signature
}

func (e *Extrinsic) ParityDecode(pd paritycodec.Decoder) {
	e.transfer.ParityDecode(pd)
	(*primitives.H512)(&e.signature).ParityDecode(pd)
}

func (e Extrinsic) ParityEncode(pe paritycodec.Encoder) {
	e.transfer.ParityEncode(pe)
	(primitives.H512)(e.signature).ParityEncode(pe)
}

type BlockNumber uint64

type HashOutput primitives.H256 // BlakeTwo256::Output

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
	digest srprimitives.Digest
}

func (h *Header) ParityDecode(pd paritycodec.Decoder) {
	(*primitives.H256)(&h.parentHash).ParityDecode(pd)
	h.number = BlockNumber(pd.DecodeUintCompact())
	(*primitives.H256)(&h.stateRoot).ParityDecode(pd)
	(*primitives.H256)(&h.extrinsicsRoot).ParityDecode(pd)
	(&h.digest).ParityDecode(pd)
}

type Extrinsics []Extrinsic

func (e *Extrinsics) ParityDecode(pd paritycodec.Decoder) {
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
	b.header.ParityDecode(pd)
	b.extrinsics.ParityDecode(pd)
}

var EXTRINSIC_INDEX = []byte(":extrinsic_index")

type Result struct {
	isError       bool
	okOrErrorCode byte
}

type ApplyError byte

type ApplyOutcome byte

func Ok(v ApplyOutcome) Result {
	return Result{false, byte(v)}
}

func Err(v ApplyError) Result {
	return Result{true, byte(v)}
}

const (
	/// Bad signature.
	BadSignature ApplyError = 0
	/// Nonce too low.
	Stale ApplyError = 1
	/// Nonce too high.
	Future ApplyError = 2
	/// Sending account had too low a balance.
	CantPay ApplyError = 3
	/// Successful application (extrinsic reported no issue).
	Success ApplyOutcome = 0
	/// Failed application (extrinsic was probably a no-op other than fees).
	Fail ApplyOutcome = 1
)

var NONCE_OF = []byte("nonce:")
var BALANCE_OF = []byte("balance:")

func executeTransactionBackend(utx Extrinsic) Result {
	// check signature
	utx.signature.Verify(paritycodec.Encode(utx.transfer), primitives.H256(utx.transfer.from))

	// check nonce
	nonce_key := ConcatByteSlices(NONCE_OF, paritycodec.Encode(primitives.H256(utx.transfer.from)))
	expected_nonce := srio.StorageGetUint64Or(nonce_key, 0)
	if utx.transfer.nonce != expected_nonce {
		return Err(Stale)
	}

	// increment nonce in storage
	srio.StoragePutUint64(nonce_key, expected_nonce+1)

	// check sender balance
	from_balance_key := ConcatByteSlices(BALANCE_OF, paritycodec.Encode(primitives.H256(utx.transfer.from)))
	from_balance := srio.StorageGetUint64Or(from_balance_key, 0)

	// enact transfer
	if utx.transfer.amount > from_balance {
		return Err(CantPay)
	}
	to_balance_key := ConcatByteSlices(BALANCE_OF, paritycodec.Encode(primitives.H256(utx.transfer.to)))
	to_balance := srio.StorageGetUint64Or(to_balance_key, 0)
	srio.StoragePutUint64(from_balance_key, from_balance-utx.transfer.amount)
	srio.StoragePutUint64(to_balance_key, to_balance+utx.transfer.amount)
	return Ok(Success)
}

func digestEqual(d1 srprimitives.Digest, d2 srprimitives.Digest) bool {
	if len(d1.Logs) != len(d2.Logs) {
		return false
	}
	for i := range d1.Logs {
		if d1.Logs[i] != d2.Logs[i] {
			return false
		}
	}
	return true
}

//go:export Core_execute_block
func executeBlock(offset *byte, length uintptr) uint64 {
	block := Block{}
	mr := NewMemReader(offset, length)
	pd := paritycodec.Decoder{&mr}
	block.ParityDecode(pd)

	// check transaction trie root represents the transactions.
	txs := make([][]byte, len(block.extrinsics))
	for i, e := range block.extrinsics {
		txs[i] = paritycodec.Encode(e)
	}

	txsRoot := srio.EnumeratedTrieRootBlake256ForByteSlices(txs)
	if txsRoot != block.header.extrinsicsRoot {
		panic("Transaction trie root must be valid.")
	}

	// execute transactions
	for i, e := range block.extrinsics {
		var buffer = bytes.Buffer{}
		paritycodec.Encoder{&buffer}.EncodeUint(uint64(i), 4)
		srio.StoragePut(EXTRINSIC_INDEX, buffer.Bytes())
		res := executeTransactionBackend(e)
		srio.StorageKill(EXTRINSIC_INDEX)
		if res.isError {
			panic("Extrinsic error " + strconv.Itoa(int(res.okOrErrorCode)))
		}
	}

	sr := srio.StorageRoot()
	if *sr != primitives.H256(block.header.stateRoot) {
		panic("Storage root must match that calculated.")
	}

	// check digest
	digest := srprimitives.Digest{[]srprimitives.DigestItem{}, func() srprimitives.AuthorityId { return &AuthorityId{} }}
	if len(digest.Logs) > 0 {
		panic("whoa")
	}
	phb := block.header.parentHash[:]
	ok, scr := srio.StorageChangesRoot(phb, uint64(block.header.number)-1)
	if ok {
		digest.Logs = append(digest.Logs, srprimitives.ChangesTrieRoot(*scr))
	}
	if !digestEqual(digest, block.header.digest) {
		panic("Header digest items must match that calculated.")
	}
	return 0
}

// TODO: learn to build WASM modules in TinyGo without main()
func main() {

}
