package srsandbox

import "unsafe"

//go:export ext_sandbox_instantiate
func ext_sandbox_instantiate(
	dispatch_thunk_table_id uint32,
	wasm_ptr *byte,
	wasm_len uintptr,
	imports_ptr *byte,
	imports_len uintptr,
	state unsafe.Pointer,
) uint32

//go:export ext_sandbox_invoke
func ext_sandbox_invoke(
	instance_idx uint32,
	export_ptr *byte,
	export_len uintptr,
	args_ptr *byte,
	args_len uintptr,
	return_val_ptr *byte,
	return_val_len uintptr,
	state uintptr,
) uint32

//go:export ext_sandbox_memory_new
func ext_sandbox_memory_new(initial uint32, maximum uint32) uint32

//go:export ext_sandbox_memory_get
func ext_sandbox_memory_get(
	memory_idx uint32,
	offset uint32,
	buf_ptr *byte,
	buf_len uintptr,
) uint32

//go:export ext_sandbox_memory_set
func ext_sandbox_memory_set(
	memory_idx uint32,
	offset uint32,
	val_ptr *byte,
	val_len uintptr,
) uint32

//go:export ext_sandbox_memory_teardown
func ext_sandbox_memory_teardown(
	memory_idx uint32,
)

//go:export ext_sandbox_instance_teardown
func ext_sandbox_instance_teardown(
	instance_idx uint32,
)
