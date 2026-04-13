package generator_test

//nolint:gocheckcompilerdirectives // go:fix is a valid Go 1.24+ pragma for inlinable wrappers
//go:fix inline
func int64Ptr(v int64) *int64 { return new(v) }
