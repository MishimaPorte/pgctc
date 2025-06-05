package types

import (
	"fmt"
	"unsafe"
	"database/sql/driver"
	"bytes"
	"time"
	"encoding/hex"
	"strconv"
)

func (v *Keks) Scan(thing any) (err error) {
	var source, ok = thing.(string)
	if !ok {
		return fmt.Errorf("incompatible type: %+v", thing)
	}
	if source[0] == '"' {
		if source, _, err = __intrinsic_readString(unsafe.Slice(unsafe.StringData(source), len(source))[1:], 0); err != nil {
			return err
		}
	}
	var sourceBytes = unsafe.Slice(unsafe.StringData(source), len(source))
	if sourceBytes[0] != '(' || sourceBytes[len(source)-1] != ')' {
		return fmt.Errorf("bad source string: %q", source)
	}
	var cur = 1
	{
		*v = make([]Kek, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				*v = append(*v, Kek(struct {
					NamedSlice NamedSlice ""
				}{}))
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); n == 0 {
				} else if e != nil {
					return e
				} else {
					cur = cur + n
					if e := (*v)[len(*v)-1].Scan(next); e != nil {
						return e
					}
				}
				cur = cur + 1
			}
		}
	}
	_ = cur
	return nil
}
func (v *Kek) Scan(thing any) (err error) {
	var source, ok = thing.(string)
	if !ok {
		return fmt.Errorf("incompatible type: %+v", thing)
	}
	if source[0] == '"' {
		if source, _, err = __intrinsic_readString(unsafe.Slice(unsafe.StringData(source), len(source))[1:], 0); err != nil {
			return err
		}
	}
	var sourceBytes = unsafe.Slice(unsafe.StringData(source), len(source))
	if sourceBytes[0] != '(' || sourceBytes[len(source)-1] != ')' {
		return fmt.Errorf("bad source string: %q", source)
	}
	var cur = 1
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); n == 0 {
	} else if e != nil {
		return e
	} else {
		cur = cur + n
		if e := v.NamedSlice.Scan(next); e != nil {
			return e
		}
	}
	_ = cur
	return nil
}
func (v *NamedSlice) Scan(thing any) (err error) {
	var source, ok = thing.(string)
	if !ok {
		return fmt.Errorf("incompatible type: %+v", thing)
	}
	if source[0] == '"' {
		if source, _, err = __intrinsic_readString(unsafe.Slice(unsafe.StringData(source), len(source))[1:], 0); err != nil {
			return err
		}
	}
	var sourceBytes = unsafe.Slice(unsafe.StringData(source), len(source))
	if sourceBytes[0] != '(' || sourceBytes[len(source)-1] != ')' {
		return fmt.Errorf("bad source string: %q", source)
	}
	var cur = 1
	{
		*v = make([]string, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				*v = append(*v, "")
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); e != nil {
					return e
				} else {
					cur = cur + n
					(*v)[len(*v)-1] = next
				}
				cur = cur + 1
			}
		}
	}
	_ = cur
	return nil
}

func __intrinsic_readString(b []byte, delim byte) (res string, read int, err error) {
	switch b[0] {
	case '"':
		var buf bytes.Buffer
		for i := 1; i < len(b); i++ {
			var n = b[i]
			switch n {
			case '"':
				i++
				if i == len(b) {
					return buf.String(), i, nil
				} else if n = b[i]; n == '"' {
					buf.WriteByte('"')
				} else {
					return buf.String(), i, nil
				}
			case '\\':
				i++
				if i == len(b) {
					return "", 0, fmt.Errorf("bad escape sequence: %q", string(b))
				}
				n = b[i]
				switch n {
				case '"':
					buf.WriteByte('"')
				case '\\':
					buf.WriteByte('\\')
				default:
					return "", 0, fmt.Errorf("bad escape sequence: %q", string(b))
				}
			default:
				buf.WriteByte(n)
			}
		}
		return "", 0, fmt.Errorf("bad escaped string literal: %s", string(b))
	default:
		for i := range b {
			if b[i] == delim || b[i] == ',' {
				return string(b[:i]), i, nil
			}
		}
		return "", 0, fmt.Errorf(`readString should be called only for '"', '{' or '(' - delimited strings or records or array literals`)
	}
}

func __intrinsic_computeStringFromDriverValuer(value driver.Value) string {
	switch v := value.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "t"
		} else {
			return "f"
		}
	case []byte:
		return strconv.Quote("\\x" + hex.EncodeToString(v))
	case string:
		return strconv.Quote(v)
	case time.Time:
		return strconv.Quote(v.String())
	default:
		panic("all is bad: bad value happened inside '__intrinsic_computeStringFromDriverValuer' intrinsic: " + fmt.Sprintf("%#v", value))
	}
}
