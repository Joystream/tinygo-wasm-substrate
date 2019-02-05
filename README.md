# WebAssembly modules for Substrate platform in TinyGo.

## What?

Substrate blockchain environment (https://github.com/paritytech/substrate) runs user-defined code that
is compiled as WebAssembly (https://developer.mozilla.org/en-US/docs/WebAssembly) modules.

Substrate developers provide libraries (Substrate Core and SRML) to write the code using the Rust language.
See more at https://www.parity.io/substrate-has-arrived/.

This is an experimental implementation of Substrate-compatible WASM modules in Tinygo
(https://github.com/aykevl/tinygo/), a subset of Go language that is used for low-level
targets.

This implementation includes **SRML and core modules** that are required to build runtimes
and **three example module implementations** to be verified against the Substrate platform.

SRML is rougly 30% implemented (see below).

(Unfortunately, using mainstream implementation of Go is currently infeasible, since
WebAssembly support in Go is quite limited and the resulting code is encumbered by the
huge Go runtime).

## Status: examples

Currently we compile example modules that pass all the "black box" tests included in Substrade:

* `executortestmodule`: for `core/executor` tests for basic IO and sandbox support
* `testruntime`: for `core/test-runtime` tests, a basic `execute_block` implementation

There is also an *incomplete* port of https://github.com/paritytech/substrate-node-template/
in `nodetemplateruntime`.

## Status: SRML

SRML consists of various modules. Some (primitives, io, support, system, executive) are essential
to running a module; some provide application-level functionality.

Not all of SRML is ported to Go. The current status is approximately this:

<table>
<tr><th>Module</th><th>Percent done</th><th>Missing bits and general notes</th></tr>
<tr><td>sr-api</td><td>0</td><td>Should not be converted as is (Rust macros that transform API definitions)</td></tr>
<tr><td>sr-io</td><td>80</td><td>Missing: child storage, new hash functions (keccak, secp256k1), tests</td></tr>
<tr><td>sr-primitives</td><td>35</td><td>Missing: permill/perbill, log macro, traits, era, uncheckeds</td></tr>
<tr><td>sr-sandbox</td><td>95</td><td></td></tr>
<tr><td>sr-version</td><td>50</td><td>Helper methods, serialization?</td></tr>
<tr><td>state-machine</td><td>0</td><td>(maybe not needed at all for the runtime)?</td></tr>
<tr><td>srml-assets</td><td>0</td><td></td></tr>
<tr><td>srml-balances</td><td>0</td><td></td></tr>
<tr><td>srml-consensus</td><td>0</td><td></td></tr>
<tr><td>srml-contract</td><td>0</td><td></td></tr>
<tr><td>srml-council</td><td>0</td><td></td></tr>
<tr><td>srml-democracy</td><td>0</td><td></td></tr>
<tr><td>srml-example,</td><td>0</td><td>Mostly valuable because of example code and comments</td></tr>
<tr><td>srml-executive</td><td>65</td><td>Tests, latest changes</td></tr>
<tr><td>srml-grandpa</td><td>0</td><td></td></tr>
<tr><td>srml-indices</td><td>30</td><td>Address encoding, resolvehint, module+storage, tests</td></tr>
<tr><td>srml-metadata</td><td>40</td><td>Additional types and serialization</td></tr>
<tr><td>srml-session</td><td>0</td><td></td></tr>
<tr><td>srml-staking</td><td>0</td><td></td></tr>
<tr><td>srml-sudo</td><td>0</td><td></td></tr>
<tr><td>srml-support/procedural/storage</td><td>80</td><td>(hard to judge, rust macros were converted to go runtime storage definitions)</td></tr>
<tr><td>srml-support/src/dispatch</td><td>70</td><td>(hard to judge, rust macros were converted to go runtime module definitions)</td></tr>
<tr><td>srml-support/src/double_map</td><td>0</td><td></td></tr>
<tr><td>srml-support/src/event</td><td>0</td><td></td></tr>
<tr><td>srml-support/src/hashable</td><td>0</td><td></td></tr>
<tr><td>srml-support/src/inherent</td><td>60</td><td>(hard to judge, rust macros were converted to go runtime module definitions)</td></tr>
<tr><td>srml-support/src/metadata</td><td>0</td><td></td></tr>
<tr><td>srml-support/src/origin</td><td>0</td><td></td></tr>
<tr><td>srml-support/src/runtime</td><td>70</td><td>(hard to judge, rust macros were converted to go runtime module definitions)</td></tr>
<tr><td>srml-system</td><td>40</td><td>is_account, RawLog, add_extra_genesis, ensure_, externalities, set_, ChainContext, tests</td></tr>
<tr><td>srml-timestamp</td><td>0</td><td></td></tr>
<tr><td>srml-treasury</td><td>0</td><td></td></tr>
<tr><td>srml-upgrade-key</td><td>0</td><td></td></tr>
<tr><td>keyring</td><td>0</td><td></td></tr>
<tr><td>client</td><td>0</td><td></td></tr>
<tr><td>primitives</td><td>60</td><td>missing tests</td></tr>
<tr><td>consensus-aura-primitives</td><td>0</td><td></td></tr>
</table>

## Differences between Rust and Go implementations

The Rust implementation heavily uses macros to provide adaptors for functions and data structures,
generics to define customizable data structures, enums to define variant types.
Unfortunately, Go does not support any of those.

Instead of macros, we are using direct implementations of the structures and transformations.

Instead of generic parameters, either a single "compromise" concrete type or an `interface` is used. 
In the latter case, one needs to perform runtime type checks manually, and also supply
things like "New()" methods for the types.

Instead of enums, we are using corresponding idiomatic constructs in Go: interfaces, consts, 
multi-value returns. A preference is given to interfaces when possible, since they tend
to be more convenient and provide for safer type-checking.

The package naming follows the go convention of joinedlowercasewords, therefore "-" and "_" from
corresponging Rust packages is dropped. We try to keep renaming imports at minimum, for the code
to remain readable.

## Notes on TinyGo

TinyGo implementation lacks some features of mainline Go; most notably, the support of maps and
garbage collection is inadequate at the moment. We plan to address this, collaborating with the 
TinyGo project. See more at: https://github.com/aykevl/tinygo/issues/115 
https://github.com/aykevl/tinygo/issues/46#issuecomment-452642874

## Notes on WebAssembly integration

WebAssembly environment is a low-level one, and offers only very primitive data types (`i32`, `i64`, etc).
It requires some low-level expertise to understand how TinyGo and Rust calling conventions
are mapped to those. In particular:

* If a data structure (string, record, array) needs to be passed in the parameters, it is usually 
  encoded as a byte array, stored in the module's memory and represented as a pair of (pointer, length),
  or just a pointer (in case of a fixed size structure)

* Regarding encoding, see https://github.com/Joystream/parity-codec-go

* WebAssembly functions can only retun one value. If a function needs to return a data structure (byte array), 
  the pointer and length of that can be encoded as a single `i64` value. If a function needs to return more
  than one value, the function's incoming parameters might supply a memory location (pointer) to write the
  additional value(s) to.

* In WebAssembly function references are not memory pointers, but rather indices into the module's "table".
  Since Substrate's Sandbox feature publishes a function reference, the module needs to be compiled with 
  the `-ldflags="--export-table"` flag to export the table. A patch for TinyGo is used to solve some 
  additional function reference difficulties: https://github.com/aykevl/tinygo/pull/135

## How to run executor test module

Executor test module is a very simple module that is used to test
wasm-executor by exporting functions that test "ext_" imported functions.

Original Rust source: https://github.com/paritytech/substrate/tree/master/core/executor/wasm

(Instructions tested on Ubuntu Linux)

Ensure you have Go installed (see https://golang.org/dl/)

Use a patched Tinygo compiler:

    go get -u github.com/kyegupov/tinygo

*(TODO: push the compiler patches upstream)*

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

All tests shall pass.


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
