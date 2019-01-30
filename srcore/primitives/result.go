package primitives

// Rust result
type Result interface {
	IsError() bool
}
