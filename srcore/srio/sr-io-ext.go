package srio

import "github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"

// External functions, provided by the host

// See also ext_adapters.go

// The code assumes that *byte == uintptr == uint32, which is respected
// by Tinygo compiler, as of early 2019

// TODO: all these should be unexported, export only high-level wrappers

//go:export ext_clear_prefix
func ext_clear_prefix(prefix_data *byte, prefix_len uint32)

//go:export ext_print_utf8
func ext_print_utf8(utf8_data *byte, utf8_len uint32)

//go:export ext_set_storage
func ext_set_storage(key_data *byte, key_len uint32, value_data *byte, value_len uint32)

//go:export ext_get_allocated_storage
func ext_get_allocated_storage(key_data *byte, key_len uint32, value_len_ptr *uint32) *byte

//go:export ext_clear_storage
func ext_clear_storage(key_data *byte, key_len uint32)

//go:export ext_blake2_256
func ext_blake2_256(data *byte, len uint32, out *byte)

//go:export ext_twox_128
func ext_twox_128(data *byte, len uint32, out *byte)

//go:export ext_twox_256
func ext_twox_256(data *byte, len uint32, out *byte)

//go:export ext_ed25519_verify
func Ext_ed25519_verify(msg_data *byte, msg_len uint32, sig_data *byte, pubkey_data *byte) uint32

//go:export ext_blake2_256_enumerated_trie_root
func ext_blake2_256_enumerated_trie_root(values_data *byte, lens_data_addr *uint32, lens_len uint32, resultPtr *byte)

//go:export ext_storage_root
func ext_storage_root(resultPtr *primitives.H256)

//go:export ext_storage_changes_root
func ext_storage_changes_root(parent_hash_data *byte, parent_hash_len uint32, parent_num uint64, result *primitives.H256) uint32

//go:export ext_clear_prefix
func Ext_clear_prefix(prefix_data *byte, prefix_len uint32)
