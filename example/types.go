package example

import "net"

type POD_1 struct {
	Kek string
}

type SomeData struct {
	someData                 []byte
	someExternalLibraryStaff net.Conn
}

func (s *SomeData) Scan(i int) (err error) {
	return
}

// Then the valuer wrapper gets generated:
//
//	func (s *SomeData) Value() (i *SomeOtherType, err error) {
//	  return &SomeOtherType{"kek", "lol"}, nil
//	}
//
//	func __SomeData_ActualValuer(v *SomeData) (driver.Value, error){
//	  var val, err = v.Value()
//	  // converting the SomeOtherType instance into driver.Valuer in a type-safe way
//	  ...conversion code
//	}
//
// The idea is that for unconvertable types a user can provide a method to convert
// the value into a POD type that we can essentially handle.
// This makes user free of writing code that deals with buffers and other
// things like that.
func (s *SomeData) ToDriverValue() (i POD_1) {
	return POD_1{Kek: s.someExternalLibraryStaff.LocalAddr().String()}
}
