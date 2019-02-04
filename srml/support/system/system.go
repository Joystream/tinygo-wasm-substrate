package system

import (
	"bytes"

	"github.com/Joystream/tinygo-wasm-substrate/gohelpers"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srio"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	"github.com/Joystream/tinygo-wasm-substrate/srml/support"
	"github.com/Joystream/tinygo-wasm-substrate/srml/support/storage"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

type Module struct {
	support.BaseModule
	TypeParamsFactory   support.TypeParamsFactory
	BlockHashStore      storage.MapStorageValue
	AccountNonceStore   storage.MapStorageValue
	NumberStore         storage.SimpleStorageValue
	EventsStore         storage.SimpleStorageValue
	ExtrinsicDataStore  storage.MapStorageValue
	ExtrinsicCountStore storage.SimpleStorageValue
	RandomSeedStore     storage.SimpleStorageValue
	ParentHashStore     storage.SimpleStorageValue
	ExtrinsicsRootStore storage.SimpleStorageValue
	DigestStore         storage.SimpleStorageValue
}

func (m *Module) InitForRuntime(r support.TypeParamsFactory) {
	m.TypeParamsFactory = r

	m.BlockHashStore = storage.MapStorageValue{
		[]byte("System BlockHash"),
		"",
		func() storage.StoredValue { return m.TypeParamsFactory.NewHash(0) },
		func(pd codec.Decoder) storage.StoredValue {
			r := m.TypeParamsFactory.NewHash(0)
			r.ParityDecode(pd)
			return r
		},
	}
	m.NumberStore = storage.SimpleStorageValue{
		[]byte("System Number"),
		"",
		func() storage.StoredValue { return m.TypeParamsFactory.BlockNumber(1) },
		func(pd codec.Decoder) storage.StoredValue {
			r := m.TypeParamsFactory.BlockNumber(0)
			r.ParityDecode(pd)
			return r
		},
	}
	m.AccountNonceStore = storage.MapStorageValue{
		[]byte("System AccountNonce"),
		"",
		func() storage.StoredValue { return m.TypeParamsFactory.ZeroIndex() },
		func(pd codec.Decoder) storage.StoredValue {
			r := m.TypeParamsFactory.ZeroIndex()
			r.ParityDecode(pd)
			return r
		},
	}
	m.EventsStore = storage.SimpleStorageValue{
		[]byte("System Events"),
		"",
		func() storage.StoredValue { return []EventRecord{} },
		m.DecodeEventRecords,
	}
	m.ExtrinsicDataStore = storage.MapStorageValue{
		[]byte("System ExtrinsicData"),
		"",
		func() storage.StoredValue { return []byte{} },
		func(pd codec.Decoder) storage.StoredValue {
			return pd.DecodeByteSlice()
		},
	}
	m.ExtrinsicCountStore = storage.SimpleStorageValue{
		[]byte("System ExtrinsicCount"),
		"",
		func() storage.StoredValue { return nil },
		func(pd codec.Decoder) storage.StoredValue {
			b := pd.DecodeBool()
			if b {
				return pd.DecodeUint32()
			}
			return nil
		},
	}
	m.RandomSeedStore = storage.SimpleStorageValue{
		[]byte("System RandomSeed"),
		"",
		func() storage.StoredValue { return m.TypeParamsFactory.NewHash(0) },
		func(pd codec.Decoder) storage.StoredValue {
			r := m.TypeParamsFactory.NewHash(0)
			r.ParityDecode(pd)
			return r
		},
	}
	m.ParentHashStore = storage.SimpleStorageValue{
		[]byte("System ParentHash"),
		"",
		func() storage.StoredValue { return m.TypeParamsFactory.NewHash(69) },
		func(pd codec.Decoder) storage.StoredValue {
			r := m.TypeParamsFactory.NewHash(0)
			r.ParityDecode(pd)
			return r
		},
	}
	m.ExtrinsicsRootStore = storage.SimpleStorageValue{
		[]byte("System ExtrinsicsRoot"),
		"",
		func() storage.StoredValue { return m.TypeParamsFactory.NewHash(0) },
		func(pd codec.Decoder) storage.StoredValue {
			r := m.TypeParamsFactory.NewHash(0)
			r.ParityDecode(pd)
			return r
		},
	}
	m.DigestStore = storage.SimpleStorageValue{
		[]byte("System ExtrinsicsRoot"),
		"",
		func() storage.StoredValue { return srprimitives.Digest{} },
		func(pd codec.Decoder) storage.StoredValue {
			d := srprimitives.Digest{}
			pd.DecodeCollection(
				func(n int) { d.Logs = make([]srprimitives.DigestItem, n) },
				func(i int) { d.Logs[i] = m.TypeParamsFactory.DecodeDigestItem(pd) },
			)
			return d
		},
	}
}

func (m *Module) Initialise(number srprimitives.BlockNumber, parentHash srprimitives.Hash, txsRoot srprimitives.Hash) {
	srio.UnhashedPut(srio.EXTRINSIC_INDEX, codec.ToBytesCustom(func(pe codec.Encoder) { pe.EncodeUint32(0) }))
	m.NumberStore.Put(number)
	m.ParentHashStore.Put(parentHash)
	m.BlockHashStore.Insert(number.MinusOne(), parentHash)
	m.ExtrinsicsRootStore.Put(txsRoot)
	m.RandomSeedStore.Put(m.CalculateRandom())
	m.EventsStore.Kill()
}

/// Remove temporary "environment" entries in storage.
func (m *Module) Finalise() srprimitives.Header {

	m.RandomSeedStore.Kill()
	m.ExtrinsicCountStore.Kill()

	number := m.NumberStore.Take().(srprimitives.BlockNumber)
	parentHash := m.ParentHashStore.Take().(srprimitives.HashOutput)
	digest := m.DigestStore.Take().(srprimitives.Digest)
	extrinsicsRoot := m.ExtrinsicsRootStore.Take().(srprimitives.HashOutput)
	storageRoot := srio.StorageRoot()
	has, storageChangesRoot := srio.StorageChangesRoot(parentHash.AsBytes(), number.MinusOne().AsUint64())

	// we can't compute changes trie root earlier && put it to the Digest
	// because it will include all currently existing temporaries
	if has {
		item := srprimitives.ChangesTrieRoot(*storageChangesRoot)
		digest.Logs = append(digest.Logs, item)
	}

	// <Events<T>> stays to be inspected by the client.

	return srprimitives.Header{parentHash, number, storageRoot, extrinsicsRoot, digest}
}

/// Calculate the current block's random seed.
func (m *Module) CalculateRandom() srprimitives.HashOutput {
	blockNumber := m.NumberStore.Get().(srprimitives.BlockNumber)
	zeroBlockNumber := m.TypeParamsFactory.BlockNumber(0)
	gohelpers.Assert(blockNumber.GreaterThan(zeroBlockNumber), "Block number may never be zero")
	state := blockNumber.MinusOne()
	hashes := []srprimitives.HashOutput{}
	for i := 0; i < 81; i++ {
		if state.GreaterThan(zeroBlockNumber) {
			state = state.MinusOne()
		}
		hashes = append(hashes, m.BlockHashStore.Get(state).(srprimitives.HashOutput))
	}

	return srprimitives.TripletMix(hashes)
}

func (m *Module) ExtrinsicIndex() (bool, uint32) {
	ok, v := srio.UnhashedGet(srio.EXTRINSIC_INDEX)
	if ok {
		return true, codec.Decoder{bytes.NewBuffer(v)}.DecodeUint32()
	}
	return false, 0
}

/// Increment a particular account's nonce by 1.
func (m *Module) IncAccountNonce(sender srprimitives.AccountId) {
	val := m.AccountNonceStore.Get(sender).(srprimitives.Index)
	m.AccountNonceStore.Insert(sender, val.PlusOne())
}

/// Note what the extrinsic data of the current extrinsic index is. If this is called, then
/// ensure `derive_extrinsics` is also called before block-building is completed.
func (m *Module) NoteExtrinsic(encodedXt []byte) {
	_, index := m.ExtrinsicIndex()
	m.ExtrinsicDataStore.Insert(gohelpers.Uint32(index), gohelpers.ByteSlice(encodedXt))
}

/// To be called immediately after an extrinsic has been applied.
func (m *Module) NoteAppliedExtrinsic(maybeError error) {
	if maybeError != nil {
		m.DepositEventCall(EventExtrinsicSuccess{}).Dispatch()
	} else {
		m.DepositEventCall(EventExtrinsicFailed{}).Dispatch()
	}
	_, exInd := m.ExtrinsicIndex()
	nextExtrinsicIndex := exInd + 1
	srio.UnhashedPut(srio.EXTRINSIC_INDEX, codec.ToBytesCustom(func(pe codec.Encoder) { pe.EncodeUint32(nextExtrinsicIndex) }))
}

/// To be called immediately after `note_applied_extrinsic` of the last extrinsic of the block
/// has been called.
func (m *Module) NoteFinishedExtrinsics() {
	_, extrinsicIndex := srio.UnhashedGet(srio.EXTRINSIC_INDEX)
	srio.UnhashedKill(srio.EXTRINSIC_INDEX)
	m.ExtrinsicCountStore.Put(gohelpers.ByteSlice(extrinsicIndex))
}

/// Remove all extrinsics data and save the extrinsics trie root.
func (m *Module) DeriveExtrinsics() {
	extrinsicCountStored := m.ExtrinsicCountStore.Get()
	extrinsicCount := uint32(0)
	if extrinsicCountStored != nil {
		extrinsicCount = extrinsicCountStored.(uint32)
	}
	extrinsics := make([]codec.Encodeable, extrinsicCount)
	for i := range extrinsics {
		extrinsics[i] = m.ExtrinsicDataStore.Get(gohelpers.Uint32(i)).(srprimitives.Extrinsic).EncodeableEnum()
	}

	xtsRoot := primitives.H256(srio.EnumeratedTrieRootBlake256(extrinsics))
	m.ExtrinsicsRootStore.Put(&xtsRoot)
}

// Method IDs
const (
	DepositEventId byte = 0
)

type DepositEventCall struct {
	m     *Module
	event support.Event
}

func (d DepositEventCall) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{DepositEventId, d.event}
}

func (m *Module) DepositEventCall(event support.Event) DepositEventCall {
	return DepositEventCall{m, event}
}

func (c DepositEventCall) Dispatch(o srprimitives.Origin) error {

	ok, extrinsicIndex := c.m.ExtrinsicIndex()
	var phase Phase = PhaseFinalization{}
	if ok {
		phase = PhaseApplyExtrinsic(extrinsicIndex)
	}
	events := c.m.EventsStore.Get().([]EventRecord)
	events = append(events, EventRecord{phase, c.event})
	c.m.EventsStore.Put(EventRecords(events))
	return nil
}

func (m *Module) CallableBelongsToThisModule(c srprimitives.Callable) bool {
	switch c.(type) {
	case DepositEventCall:
		return true
	}
	return false
}

type ChainContext struct {
}

/// Origin for the system module.
type RawOrigin interface {
	srprimitives.Origin
	ImplementsRawOrigin()
}

/// The system itself ordained this dispatch to happen: this is the highest privilege level.
type RawOriginRoot struct{}

func (_ RawOriginRoot) ImplementsRawOrigin() {}

/// It is signed by some public key and we provide the AccountId.
type RawOriginAccountId struct {
	v srprimitives.AccountId
}

func (_ RawOriginAccountId) ImplementsRawOrigin() {}

/// It is signed by nobody but included and agreed upon by the validators anyway: it's "inherently" true.
type RawOriginInherent struct{}

func (_ RawOriginInherent) ImplementsRawOrigin() {}

func OptionAccountIdToOrigin(present bool, accountId srprimitives.AccountId) RawOrigin {
	if present {
		return RawOriginAccountId{accountId}
	} else {
		return RawOriginInherent{}
	}
}

type Phase interface {
	ImplementsPhase()
	EncodeableEnum() primitives.EncodeableEnum
}

type PhaseApplyExtrinsic uint32

func (_ PhaseApplyExtrinsic) ImplementsPhase() {}

func (p PhaseApplyExtrinsic) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{0, gohelpers.Uint32(p)}
}

type PhaseFinalization struct {
}

func (_ PhaseFinalization) ImplementsPhase() {}

func (p PhaseFinalization) EncodeableEnum() primitives.EncodeableEnum {
	return primitives.EncodeableEnum{1, primitives.NoPayload{}}
}

func DecodePhase(pd codec.Decoder) Phase {
	b := pd.DecodeByte()
	switch b {
	case 0:
		return PhaseApplyExtrinsic(pd.DecodeUint32())
	case 1:
		return PhaseFinalization{}
	}
	panic(primitives.InvalidEnum(b, "Phase"))
}

type EventRecord struct {
	/// The phase of the block it happened in.
	Phase Phase
	/// The event itself.
	Event support.Event
}

type EventRecords []EventRecord

func (e *EventRecord) ParityEncode(pe codec.Encoder) {
	e.Phase.EncodeableEnum().ParityEncode(pe)
	e.Event.ParityEncode(pe)
}

func (m *Module) DecodeEventRecord(pd codec.Decoder) EventRecord {
	var e EventRecord
	e.Phase = DecodePhase(pd)
	e.Event = m.TypeParamsFactory.DecodeEvent(pd)
	return e
}

func (m *Module) DecodeEventRecords(pd codec.Decoder) storage.StoredValue {
	var e *[]EventRecord
	pd.DecodeCollection(
		func(n int) { *e = make([]EventRecord, n) },
		func(i int) { (*e)[i] = m.DecodeEventRecord(pd) },
	)
	return *e
}

func (e EventRecords) ParityEncode(pe codec.Encoder) {
	pe.EncodeCollection(
		len(e),
		func(i int) { e[i].ParityEncode(pe) },
	)
}

type EventExtrinsicSuccess struct{}

func (e EventExtrinsicSuccess) ParityEncode(pe codec.Encoder) {
	pe.EncodeByte(0)
}

type EventExtrinsicFailed struct{}

func (e EventExtrinsicFailed) ParityEncode(pe codec.Encoder) {
	pe.EncodeByte(1)
}
