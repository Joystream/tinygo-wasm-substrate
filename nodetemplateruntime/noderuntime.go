package main

import (
	"github.com/Joystream/tinygo-wasm-substrate/srcore/inherents"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srprimitives"
	"github.com/Joystream/tinygo-wasm-substrate/srcore/srversion"
	executivemodule "github.com/Joystream/tinygo-wasm-substrate/srml/executive"
	metadatamodule "github.com/Joystream/tinygo-wasm-substrate/srml/metadata"
	"github.com/Joystream/tinygo-wasm-substrate/srml/support"
	runtimemodule "github.com/Joystream/tinygo-wasm-substrate/srml/support/runtime"
	codec "github.com/kyegupov/parity-codec-go/noreflect"
)

/// Alias to Ed25519 pubkey that identifies an account on the chain.
type AccountId primitives.H256

/// A hash of some data used by the chain.
type Hash primitives.H256

/// Index of a block number in the chain.
type BlockNumber uint64

/// Index of an account's extrinsic in the chain.
type Nonce uint64

/// Opaque types. These are used by the CLI to instantiate machinery that don't need to know
/// the specifics of the runtime. They can then be made to be agnostic over specific formats
/// of data like extrinsics, allowing for them to continue syncing the network through upgrades
/// to even the core datastructures.

// type UncheckedExtrinsic []byte

// func (*UncheckedExtrinsic) IsSigned() (bool, bool) {
// 	return false, false
// }

func authorityIdFactory() srprimitives.AuthorityId { return &primitives.H256{} }

func decodeDigestItem(pd codec.Decoder) srprimitives.DigestItem {
	return srprimitives.DecodeDigestItem(pd, authorityIdFactory)
}

// func DecodeExtrinsic(pd codec.Decoder) UncheckedExtrinsic {
// 	return UncheckedExtrinsic(pd.DecodeByteSlice())
// }

type SessionKey primitives.Ed25519AuthorityId

// End opaque types

/// This runtime version.
var VERSION srversion.RuntimeVersion = srversion.RuntimeVersion{
	SpecName:         "template-node",
	ImplName:         "template-node",
	AuthoringVersion: 3,
	SpecVersion:      3,
	ImplVersion:      0,
	ApiVersions:      []srversion.ApiVersion{},
}

// impl system.Trait for Runtime {
// 	/// The identifier used to distinguish between accounts.
// 	type AccountId = AccountId;
// 	/// The lookup mechanism to get account ID from whatever is passed in dispatchers.
// 	type Lookup = Indices;
// 	/// The index type for storing how many extrinsics an account has signed.
// 	type Index = Nonce;
// 	/// The index type for blocks.
// 	type BlockNumber = BlockNumber;
// 	/// The type for hashing blocks and tries.
// 	type Hash = Hash;
// 	/// The hashing algorithm used.
// 	type Hashing = BlakeTwo256;
// 	/// The header digest type.
// 	type Digest = generic.Digest<Log>;
// 	/// The header type.
// 	type Header = generic.Header<BlockNumber, BlakeTwo256, Log>;
// 	/// The ubiquitous event type.
// 	type Event = Event;
// 	/// The ubiquitous log type.
// 	type Log = Log;
// 	/// The ubiquitous origin type.
// 	type Origin = Origin;
// }

// impl aura.Trait for Runtime {
// 	type HandleReport = ();
// }

// impl consensus.Trait for Runtime {
// 	/// The position in the block's extrinsics that the note-offline inherent must be placed.
// 	const NOTE_OFFLINE_POSITION: u32 = 1;
// 	/// The identifier we use to refer to authorities.
// 	type SessionKey = Ed25519AuthorityId;
// 	// The aura module handles offline-reports internally
// 	// rather than using an explicit report system.
// 	type InherentOfflineReport = ();
// 	/// The ubiquitous log type.
// 	type Log = Log;
// }

// impl indices.Trait for Runtime {
// 	/// The type for recording indexing into the account enumeration. If this ever overflows, there
// 	/// will be problems!
// 	type AccountIndex = u32;
// 	/// Use the standard means of resolving an index hint from an id.
// 	type ResolveHint = indices.SimpleResolveHint<Self.AccountId, Self.AccountIndex>;
// 	/// Determine whether an account is dead.
// 	type IsDeadAccount = Balances;
// 	/// The uniquitous event type.
// 	type Event = Event;
// }

// impl timestamp.Trait for Runtime {
// 	/// The position in the block's extrinsics that the timestamp-set inherent must be placed.
// 	const TIMESTAMP_SET_POSITION: u32 = 0;
// 	/// A timestamp: seconds since the unix epoch.
// 	type Moment = uint64;
// 	type OnTimestampSet = Aura;
// }

// impl balances.Trait for Runtime {
// 	/// The type for recording an account's balance.
// 	type Balance = u128;
// 	/// What to do if an account's free balance gets zeroed.
// 	type OnFreeBalanceZero = ();
// 	/// What to do if a new account is created.
// 	type OnNewAccount = Indices;
// 	/// Restrict whether an account can transfer funds. We don't place any further restrictions.
// 	type EnsureAccountLiquid = ();
// 	/// The uniquitous event type.
// 	type Event = Event;
// }

// impl sudo.Trait for Runtime {
// 	/// The uniquitous event type.
// 	type Event = Event;
// 	type Proposal = Call;
// }

type Runtime runtimemodule.Runtime

type TypeParams struct{}

// TODO: implement
func (_ TypeParams) NewHash(b byte) srprimitives.HashOutput                    { return nil }
func (_ TypeParams) BlockNumber(uint64) srprimitives.BlockNumber               { return nil }
func (_ TypeParams) DecodeDigestItem(pd codec.Decoder) srprimitives.DigestItem { return nil }
func (_ TypeParams) ZeroIndex() srprimitives.Index                             { return nil }
func (_ TypeParams) DecodeEvent(pd codec.Decoder) support.Event                { return nil }
func (_ TypeParams) EmptyHash() srprimitives.HashOutput                        { return nil }
func (_ TypeParams) DefaultContext() interface{}                               { return nil }

var runtime = Runtime{TypeParams: TypeParams{}}

var executive = executivemodule.Executive{
	runtime.System,
	nil, // Todo: payment
	nil, // Todo: onfinalise
}

// construct_runtime!(
// 	pub enum Runtime with Log(InternalLog: DigestItem<Hash, Ed25519AuthorityId>) where
// 		Block = Block,
// 		NodeBlock = opaque::Block,
// 		UncheckedExtrinsic = UncheckedExtrinsic
// 	{
// 		System: system::{default, Log(ChangesTrieRoot)},
// 		Timestamp: timestamp::{Module, Call, Storage, Config<T>, Inherent},
// 		Consensus: consensus::{Module, Call, Storage, Config<T>, Log(AuthoritiesChange), Inherent},
// 		Aura: aura::{Module},
// 		Indices: indices,
// 		Balances: balances,
// 		Sudo: sudo,
// 	}
// );

// /// The type used as a helper for interpreting the sender of transactions.
// type Context = system.ChainContext<Runtime>;
// /// The address format for describing accounts.
// type Address = <Indices as StaticLookup>.Source;
// /// Block header type as expected by this runtime.
// type Header = generic.Header<BlockNumber, BlakeTwo256, Log>;
// /// Block type as expected by this runtime.
// type Block = generic.Block<Header, UncheckedExtrinsic>;
// /// BlockId type as expected by this runtime.
// type BlockId = generic.BlockId<Block>;
// /// Unchecked extrinsic type as expected by this runtime.
// type UncheckedExtrinsic = generic.UncheckedMortalCompactExtrinsic<Address, Nonce, Call, Ed25519Signature>;
// /// Extrinsic type that has already been checked.
// type CheckedExtrinsic = generic.CheckedExtrinsic<AccountId, Nonce, Call>;
// /// Executive: handles dispatch to the various modules.
// type Executive = executive.Executive<Runtime, Block, Context, Balances, AllModules>;

// Implement our runtime API endpoints. This is just a bunch of proxying.

//go:export "Core_version"
func version() srversion.RuntimeVersion {
	return VERSION
}

// TODO: implement consensus
// //go:export "Core_authorities"
// func authorities() []primitives.Ed25519AuthorityId {
// 	return Consensus.authorities()
// }

//go:export "Core_execute_block"
func execute_block(block srprimitives.Block) {
	executive.ExecuteBlock(&block)
}

//go:export "Core_initialise_block"
func initialise_block(header srprimitives.Header) {
	executive.InitialiseBlock(&header)
}

//go:export "Metadata_metadata"
func metadata() metadatamodule.RuntimeMetadata {
	return (*runtimemodule.Runtime)(&runtime).GetMetadata()
}

//go:export "BlockBuilder_apply_extrinsic"
func apply_extrinsic(extrinsic srprimitives.UncheckedExtrinsic) srprimitives.ApplyResult {
	return executive.ApplyExtrinsic(&extrinsic)
}

//go:export "BlockBuilder_finalise_block"
func finalise_block() srprimitives.Header {
	return executive.FinaliseBlock()
}

//go:export "BlockBuilder_inherent_extrinsics"
func inherent_extrinsics(data inherents.InherentData) []srprimitives.Extrinsic {
	return data.CreateExtrinsics((*runtimemodule.Runtime)(&runtime))
}

//go:export "BlockBuilder_check_inherents"
func check_inherents(block srprimitives.Block, data inherents.InherentData) inherents.CheckInherentsResult {
	return data.CheckExtrinsics((*runtimemodule.Runtime)(&runtime), block)
}

//go:export "BlockBuilder_random_seed"
func random_seed() srprimitives.HashOutput {
	return runtime.System.RandomSeedStore.Get().(srprimitives.HashOutput)
}

//go:export "TaggedTransactionQueue_validate_transaction"
func validate_transaction(tx srprimitives.Extrinsic) srprimitives.TransactionValidity {
	return executive.ValidateTransaction(tx)
}

// TODO: implement aura
// //go:export "AuraApi_x"
// func slot_duration() uint64 {
// 	return Aura.SlotDuration()
// }

func main() {}
