package types

import (
	"fmt"
	"unsafe"
	"strings"
	"strconv"
	"database/sql/driver"
	"bytes"
	"time"
	"encoding/hex"
)

func (v *Things) Scan(thing any) (err error) {
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
func (v *Option) Scan(thing any) (err error) {
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
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		v.Type = next
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else if answer, e := strconv.ParseInt(next, 10, 64); e != nil {
		return e
	} else {
		cur = cur + n
		v.A = int(answer)
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else if answer, e := strconv.ParseInt(next, 10, 64); e != nil {
		return e
	} else {
		cur = cur + n
		v.B = int16(answer)
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else if answer, e := strconv.ParseInt(next, 10, 64); e != nil {
		return e
	} else {
		cur = cur + n
		v.C = uint(answer)
	}
	{
		v.D = new(*****uintptr)
		{
			*v.D = new(****uintptr)
			{
				**v.D = new(***uintptr)
				{
					***v.D = new(**uintptr)
					{
						****v.D = new(*uintptr)
						{
							*****v.D = new(uintptr)
							if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
								return e
							} else if answer, e := strconv.ParseInt(next, 10, 64); e != nil {
								return e
							} else {
								cur = cur + n
								******v.D = uintptr(answer)
							}
						}
					}
				}
			}
		}
	}
	{
		v.Keks = make([]bool, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				v.Keks = append(v.Keks, false)
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); e != nil {
					return e
				} else {
					cur = cur + n
					switch next {
					case "t":
						v.Keks[len(v.Keks)-1] = true
					case "f":
						v.Keks[len(v.Keks)-1] = false
					default:
						panic("bad bool string from postgres: " + next)
					}
				}
				cur = cur + 1
			}
		}
	}
	{
		v.Things2 = make([]string, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				v.Things2 = append(v.Things2, "")
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); e != nil {
					return e
				} else {
					cur = cur + n
					v.Things2[len(v.Things2)-1] = next
				}
				cur = cur + 1
			}
		}
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		switch next {
		case "t":
			v.Required = true
		case "f":
			v.Required = false
		default:
			panic("bad bool string from postgres: " + next)
		}
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); n == 0 {
	} else if e != nil {
		return e
	} else {
		cur = cur + n
		if e := v.Lable.Scan(next); e != nil {
			return e
		}
	}
	{
		v.Name = make([]kek, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				v.Name = append(v.Name, kek(struct {
					Kek string ""
				}{}))
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); n == 0 {
				} else if e != nil {
					return e
				} else {
					cur = cur + n
					if e := v.Name[len(v.Name)-1].Scan(next); e != nil {
						return e
					}
				}
				cur = cur + 1
			}
		}
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); n == 0 {
	} else if e != nil {
		return e
	} else {
		cur = cur + n
		if e := v.Auf.Scan(next); e != nil {
			return e
		}
	}
	{
		var sliceAsArray []bool
		{
			sliceAsArray = make([]bool, 0)
			if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
				return e
			} else {
				var sourceLen = len(next)
				cur = cur + n + 1
				for cur := 1; cur < sourceLen-1; {
					sliceAsArray = append(sliceAsArray, false)
					if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); e != nil {
						return e
					} else {
						cur = cur + n
						switch next {
						case "t":
							sliceAsArray[len(sliceAsArray)-1] = true
						case "f":
							sliceAsArray[len(sliceAsArray)-1] = false
						default:
							panic("bad bool string from postgres: " + next)
						}
					}
					cur = cur + 1
				}
			}
		}
		if len(sliceAsArray) != 10 {
			return fmt.Errorf("bad parsed array element count: got %d, expected 10", len(sliceAsArray))
		}
		copy(v.Array[0:10], sliceAsArray)
	}
	{
		v.Slice = make([]bool, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				v.Slice = append(v.Slice, false)
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); e != nil {
					return e
				} else {
					cur = cur + n
					switch next {
					case "t":
						v.Slice[len(v.Slice)-1] = true
					case "f":
						v.Slice[len(v.Slice)-1] = false
					default:
						panic("bad bool string from postgres: " + next)
					}
				}
				cur = cur + 1
			}
		}
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); n == 0 {
	} else if e != nil {
		return e
	} else {
		cur = cur + n
		if e := __Scan_Option_BadThing(&v.BadThing, next); e != nil {
			return e
		}
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); n == 0 {
	} else if e != nil {
		return e
	} else {
		cur = cur + n
		if e := v.Hash.Scan(next); e != nil {
			return e
		}
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); n == 0 {
	} else if e != nil {
		return e
	} else {
		cur = cur + n
		if e := v.Extern.Scan(next); e != nil {
			return e
		}
	}
	_ = cur
	return nil
}
func (v *Option) Value() (t driver.Value, err error) {
	var b strings.Builder
	b.WriteByte('(')
	var value string
	value = strconv.Quote(v.Type)
	b.WriteString(value)
	b.WriteByte(',')
	value = strconv.FormatInt(int64(v.A), 10)
	b.WriteString(value)
	b.WriteByte(',')
	value = strconv.FormatInt(int64(v.B), 10)
	b.WriteString(value)
	b.WriteByte(',')
	value = strconv.FormatUint(uint64(v.C), 10)
	b.WriteString(value)
	b.WriteByte(',')
	if v.D == nil {
		value = ""
	} else if *v.D == nil {
		value = ""
	} else if **v.D == nil {
		value = ""
	} else if ***v.D == nil {
		value = ""
	} else if ****v.D == nil {
		value = ""
	} else if *****v.D == nil {
		value = ""
	} else {
		value = strconv.FormatUint(uint64(******v.D), 10)
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value2 string
		var value2Sb strings.Builder
		value2Sb.WriteByte('{')
		for i, val := range v.Keks {
			if val {
				value2 = "t"
			} else {
				value2 = "f"
			}
			value2Sb.WriteString(value2)
			if i != len(v.Keks)-1 {
				value2Sb.WriteByte(',')
			}
		}
		value2Sb.WriteByte('}')
		value = value2Sb.String()
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value2 string
		var value2Sb strings.Builder
		value2Sb.WriteByte('{')
		for i, val := range v.Things2 {
			value2 = strconv.Quote(val)
			value2Sb.WriteString(value2)
			if i != len(v.Things2)-1 {
				value2Sb.WriteByte(',')
			}
		}
		value2Sb.WriteByte('}')
		value = value2Sb.String()
	}
	b.WriteString(value)
	b.WriteByte(',')
	if v.Required {
		value = "t"
	} else {
		value = "f"
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value, err = v.Lable.Value()
		if err != nil {
			return nil, err
		}
		value = __intrinsic_computeStringFromDriverValuer(value)
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value2 string
		var value2Sb strings.Builder
		value2Sb.WriteByte('{')
		for i, val := range v.Name {
			{
				var value, err = val.Value()
				if err != nil {
					return nil, err
				}
				value2 = __intrinsic_computeStringFromDriverValuer(value)
			}
			value2Sb.WriteString(value2)
			if i != len(v.Name)-1 {
				value2Sb.WriteByte(',')
			}
		}
		value2Sb.WriteByte('}')
		value = value2Sb.String()
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value, err = v.Auf.Value()
		if err != nil {
			return nil, err
		}
		value = __intrinsic_computeStringFromDriverValuer(value)
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value2 string
		var value2Sb strings.Builder
		value2Sb.WriteByte('{')
		for i, val := range v.Array {
			if val {
				value2 = "t"
			} else {
				value2 = "f"
			}
			value2Sb.WriteString(value2)
			if i != len(v.Array)-1 {
				value2Sb.WriteByte(',')
			}
		}
		value2Sb.WriteByte('}')
		value = value2Sb.String()
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value2 string
		var value2Sb strings.Builder
		value2Sb.WriteByte('{')
		for i, val := range v.Slice {
			if val {
				value2 = "t"
			} else {
				value2 = "f"
			}
			value2Sb.WriteString(value2)
			if i != len(v.Slice)-1 {
				value2Sb.WriteByte(',')
			}
		}
		value2Sb.WriteByte('}')
		value = value2Sb.String()
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var Option_BadThing_value, err = __Value_Option_BadThing(&v.BadThing)
		if err != nil {
			return nil, err
		}
		value = __intrinsic_computeStringFromDriverValuer(Option_BadThing_value)
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value, err = v.Hash.Value()
		if err != nil {
			return nil, err
		}
		value = __intrinsic_computeStringFromDriverValuer(value)
	}
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value, err = v.Extern.Value()
		if err != nil {
			return nil, err
		}
		value = __intrinsic_computeStringFromDriverValuer(value)
	}
	b.WriteString(value)
	b.WriteByte(')')
	return b.String(), nil
}
func (v *lol) Scan(thing any) (err error) {
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
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		v.Kek = next
	}
	_ = cur
	return nil
}
func (v *lol) Value() (t driver.Value, err error) {
	var b strings.Builder
	b.WriteByte('(')
	var value string
	value = strconv.Quote(v.Kek)
	b.WriteString(value)
	b.WriteByte(')')
	return b.String(), nil
}
func (v *kek) Scan(thing any) (err error) {
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
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		v.Kek = next
	}
	_ = cur
	return nil
}
func (v *kek) Value() (t driver.Value, err error) {
	var b strings.Builder
	b.WriteByte('(')
	var value string
	value = strconv.Quote(v.Kek)
	b.WriteString(value)
	b.WriteByte(')')
	return b.String(), nil
}
func (v *A) Scan(thing any) (err error) {
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
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		v.Kek = next
	}
	_ = cur
	return nil
}
func (v *A) Value() (t driver.Value, err error) {
	var b strings.Builder
	b.WriteByte('(')
	var value string
	value = strconv.Quote(v.Kek)
	b.WriteString(value)
	b.WriteByte(')')
	return b.String(), nil
}
func __Scan_Option_BadThing(place *struct {
	Kek	string	""
	Lol	[]struct {
		Kek	string	""
		Lol	bool	""
	}	""
}, thing any) (err error) {
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
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		place.Kek = next
	}
	{
		place.Lol = make([]struct {
			Kek	string	""
			Lol	bool	""
		}, 0)
		if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
			return e
		} else {
			var sourceLen = len(next)
			cur = cur + n + 1
			for cur := 1; cur < sourceLen-1; {
				place.Lol = append(place.Lol, struct {
					Kek	string	""
					Lol	bool	""
				}{})
				if next, n, e := __intrinsic_readString(unsafe.Slice(unsafe.StringData(next), len(next))[cur:], ')'); n == 0 {
				} else if e != nil {
					return e
				} else {
					cur = cur + n
					if e := __Scan_Option_BadThing_Lol(&place.Lol[len(place.Lol)-1], next); e != nil {
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
func __Scan_Option_BadThing_Lol(place *struct {
	Kek	string	""
	Lol	bool	""
}, thing any) (err error) {
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
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		place.Kek = next
	}
	if next, n, e := __intrinsic_readString(sourceBytes[cur:], ')'); e != nil {
		return e
	} else {
		cur = cur + n
		switch next {
		case "t":
			place.Lol = true
		case "f":
			place.Lol = false
		default:
			panic("bad bool string from postgres: " + next)
		}
	}
	_ = cur
	return nil
}
func __Value_Option_BadThing(place *struct {
	Kek	string	""
	Lol	[]struct {
		Kek	string	""
		Lol	bool	""
	}	""
}) (t driver.Value, err error) {
	var b strings.Builder
	b.WriteByte('(')
	var value string
	value = strconv.Quote(place.Kek)
	b.WriteString(value)
	b.WriteByte(',')
	{
		var value2 string
		var value2Sb strings.Builder
		value2Sb.WriteByte('{')
		for i, val := range place.Lol {
			{
				var Option_BadThing_Lol_value, err = __Value_Option_BadThing_Lol(&val)
				if err != nil {
					return nil, err
				}
				value2 = __intrinsic_computeStringFromDriverValuer(Option_BadThing_Lol_value)
			}
			value2Sb.WriteString(value2)
			if i != len(place.Lol)-1 {
				value2Sb.WriteByte(',')
			}
		}
		value2Sb.WriteByte('}')
		value = value2Sb.String()
	}
	b.WriteString(value)
	b.WriteByte(')')
	return b.String(), nil
}
func __Value_Option_BadThing_Lol(place *struct {
	Kek	string	""
	Lol	bool	""
}) (t driver.Value, err error) {
	var b strings.Builder
	b.WriteByte('(')
	var value string
	value = strconv.Quote(place.Kek)
	b.WriteString(value)
	b.WriteByte(',')
	if place.Lol {
		value = "t"
	} else {
		value = "f"
	}
	b.WriteString(value)
	b.WriteByte(')')
	return b.String(), nil
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
