package storage

import (
	"encoding/binary"

	"github.com/Joystream/tinygo-wasm-substrate/srcore/srio"
)

// Unlike plain srio, for some reason, this package hashes keys.
// TODO: find out why

func hashStorageKey(key []byte) []byte {
	return srio.Twox128(key)
}

func Put(key []byte, value []byte) {
	srio.StoragePut(hashStorageKey(key), value)
}

func PutUint64(key []byte, value uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	Put(key, buf)
}

func Get(key []byte) (bool, []byte) {
	return srio.StorageGet(hashStorageKey(key))
}

func GetUint64Or(key []byte, deflt uint64) uint64 {
	ok, buf := Get(key)
	if ok {
		return binary.LittleEndian.Uint64(buf)
	}
	return deflt
}

func Kill(key []byte) {
	srio.StorageKill(hashStorageKey(key))
}
