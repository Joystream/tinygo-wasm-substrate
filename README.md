# WebAssembly modules for Substrate platform in TinyGo.

## What?

Substrate blockchain environment (https://github.com/paritytech/substrate) runs user-defined code that
is compiled as WebAssembly (https://developer.mozilla.org/en-US/docs/WebAssembly) modules.

Substrate developers provide libraries to write the code using Rust language.

This is an experimental implementation of Substrate-compatible WASM modules in Tinygo
(https://github.com/aykevl/tinygo/), a subset of Go language that is used for low-level
targets.

(Unfortunately, using mainstream implementation of Go is currently infeasible, since
WebAssembly support in Go is quite limited and the resulting code is encumbered by the
huge Go runtime).

## How to run executor test module

Executor test module is a very simple module that is used to test
wasm-executor by exporting functions that test "ext_" imported functions.

Original Rust source: https://github.com/paritytech/substrate/tree/master/core/executor/wasm

(Instructions tested on Ubuntu Linux)

Ensure you have Go installed (see https://golang.org/dl/)

Use a patched Tinygo compiler:

    go get -u github.com/kyegupov/tinygo

*(TODO: push the compiler patch upstream)*

Build the module:

    tinygo build -wasm-abi=generic -ldflags="--export-table" -o wasmexecutortest.wasm ./executortestmodule
    export TEST_SUBSTRATE_MODULE_PATH=`readlink -f wasmexecutortest.wasm`

Ensure you have Rust installed (see https://rustup.rs/)

Get the custom version of Substrate:

    cd .. # or whatever your directory for projects is
    git clone git@github.com:kyegupov/substrate.git
    git checkout run_wasmexec_tests_against_custom_module

*(TODO: push the substrate tests patch upstream)*

Run the tests:

    cd core/executor
    cargo test wasm_executor

All tests shall pass.

    cargo test sandbox

(Note: sandbox tests are still a work in progress).

## How to run test-runtime module

Executor test module is a small runtime module that implements a simple
transfer transaction.

Original Rust source: https://github.com/paritytech/substrate/blob/master/core/test-runtime/src/system.rs

All the instructions ar as per above, with the following changes:

Build the module:

    tinygo build -wasm-abi=generic -o str.wasm ./testruntime/
    export TEST_SUBSTRATE_RUNTIME_PATH=`readlink -f str.wasm`

Run the tests:

    cd core/test-runtime
    cargo test _wasm

All tests shall pass.
