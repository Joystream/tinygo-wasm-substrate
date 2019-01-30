package metadata

import "github.com/Joystream/tinygo-wasm-substrate/srcore/primitives"

/// All the metadata about a module.
type ModuleMetadata struct {
	Name string
	Call CallMetadata
}

/// All the metadata about a call.
type CallMetadata struct {
	Name      string
	Functions []FunctionMetadata
}

/// All the metadata about a function.
type FunctionMetadata struct {
	Id            uint16
	Name          string
	Arguments     []FunctionArgumentMetadata
	Documentation []string
}

/// All the metadata about a function argument.
type FunctionArgumentMetadata struct {
	Name string
	Ty   string
}

type EventData struct {
	Name     string
	Metadata []EventMetadata
}

/// All the metadata about an outer event.
type OuterEventMetadata struct {
	Name   string
	Events []EventData
}

/// All the metadata about a event.
type EventMetadata struct {
	Name          string
	Arguments     []string
	Documentation []string
}

/// All the metadata about a storage.
type StorageMetadata struct {
	Prefix    string
	Functions []StorageFunctionMetadata
}

/// All the metadata about a storage function.
type StorageFunctionMetadata struct {
	Name          string
	Modifier      StorageFunctionModifier
	Ty            StorageFunctionType
	Default       []byte
	Documentation []string
}

/// A storage function type.
type StorageFunctionType interface {
	primitives.Enum
	ImplementsStorageFunctionType()
}

type StorageFunctionTypePlain struct {
	V string
}

func (_ StorageFunctionTypePlain) ImplementsStorageFunctionType() {}

type StorageFunctionTypeMap struct {
	Key   string
	Value string
}

func (_ StorageFunctionTypeMap) ImplementsStorageFunctionType() {}

/// A storage function modifier.
type StorageFunctionModifier byte

const (
	StorageFunctionModifierOptional StorageFunctionModifier = 0
	StorageFunctionModifierDefault  StorageFunctionModifier = 1
)

/// All metadata about the outer dispatch.
type OuterDispatchMetadata struct {
	name  string
	calls []OuterDispatchCall
}

/// A Call from the outer dispatch.
type OuterDispatchCall struct {
	Name   string
	Prefix string
	Index  uint16
}

/// All metadata about an runtime module.
type RuntimeModuleMetadata struct {
	Prefix     string
	Module     ModuleMetadata
	HasStorage bool
	Storage    StorageMetadata
}

/// The metadata of a runtime.
type RuntimeMetadata struct {
	OuterEvent    OuterEventMetadata
	Modules       []RuntimeModuleMetadata
	OuterDispatch OuterDispatchMetadata
}
