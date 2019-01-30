package srprimitives

// Call is a universal dispatcher object that can be used to store any function of any module.
// See macro_rules! decl_module in the Rust version
type Call func(interface{}) interface{}
