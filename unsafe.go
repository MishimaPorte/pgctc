package pggen

import "unsafe"

func String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Transforms a slice of things of one kind to a slice of things of another kind.
// SAFETY: T and Z should have equal gc shapes, please
func SliceTransmute[Z, T any](items []T) []Z {
	return unsafe.Slice(((*Z)(unsafe.Pointer(unsafe.SliceData(items)))), len(items))
}

// Transforms a pointer to a one thing to a pointer to another thing.
//
// SAFETY: do not do bad shit (the T and Z should have
// equal gc shapes to not cause The Bad Thing)
func Transmute[Z, T any](item *T) *Z {
	return (*Z)(unsafe.Pointer(item))
}

// Transforms a pointer to an item into a slice of length one.
func AsSlice[T any](item *T) []T {
	return unsafe.Slice(item, 1)
}
