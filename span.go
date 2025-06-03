package pggen

import (
	"unsafe"
)

type anyint interface {
	~int | ~uint | ~uintptr |
		~int8 | ~int16 | ~int32 | ~int64 |
		~uint8 | ~uint16 | ~uint32 | ~uint64
}

// This type is like a builtin string, but for any type.
type SizedSpan[T any, I anyint] struct {
	Items  *T
	Length I
}
type Span[T any] = SizedSpan[T, int]

func (s *SizedSpan[T, I]) FromSlice(datas []T) SizedSpan[T, I] {
	var ss = SizedSpan[T, I]{unsafe.SliceData(datas), I(len(datas))}
	*s = ss
	return ss
}
func NewSpanDefault[T any](length int) Span[T] {
	return Span[T]{
		Length: length,
		Items:  unsafe.SliceData(make([]T, length)),
	}
}
func NewSpan[T any, I anyint](length I) SizedSpan[T, I] {
	return SizedSpan[T, I]{
		Length: length,
		Items:  unsafe.SliceData(make([]T, length)),
	}
}
func SpanFromSlice[I anyint, T any, Slice ~[]T](datas Slice) SizedSpan[T, I] {
	var ss = SizedSpan[T, I]{unsafe.SliceData(datas), I(len(datas))}
	return ss
}
func SpanContains[T comparable, I anyint](haystack SizedSpan[T, I], needle T) bool {
	for i := range uint64(haystack.Length) {
		if haystack.At(I(i)) == needle {
			return true
		}
	}
	return false
}
func (s SizedSpan[T, I]) Slice() []T {
	return unsafe.Slice(s.Items, s.Length)
}
func (s SizedSpan[T, I]) At(i I) T {
	return *(*T)(unsafe.Add(unsafe.Pointer(s.Items), unsafe.Sizeof(*s.Items)*uintptr(i)))
}
func (s SizedSpan[T, I]) RefAt(i I) *T {
	return (*T)(unsafe.Add(unsafe.Pointer(s.Items), unsafe.Sizeof(*s.Items)*uintptr(i)))
}
