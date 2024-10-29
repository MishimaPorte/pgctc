package meta

import (
	"fmt"
	"reflect"
)

func (g *generator) renderDriverValue(t reflect.Type) {
	switch t.Kind() {
	case reflect.Struct:
		g.fprintf(`
func (v *%s) Value() (t driver.Value, err error) {
	var b strings.Builder
`, t.Name())
		g.renderDriverValueForStruct(t)
	case reflect.Slice:
		g.fprintf(`
func (v %s) Value() (t driver.Value, err error) {
	var b strings.Builder
`, t.Name())
		g.fprintf(`
	{
		b.WriteByte('{')
		for i := range v {
`)
		g.renderSimpleToString(t.Elem(), "v[i]")
		g.fprintf(`
			b.WriteString(thing)
			if i != len(v)-1 {
				b.WriteByte(',')
			}
		}
		b.WriteByte('}')
	}
		`)

		// panic("TODO: do thing for slices")
	default:
		panic("unsupported type thing (" + t.Kind().String() + ") " + t.String())
	}
	g.fprintf(`
	return b.String(), nil
}
`)
}

func (g *generator) renderDriverValueForStruct(t reflect.Type) {
	g.fprintf(`
	b.WriteByte('(')
`)
	var nf = t.NumField()
	for i := range nf {
		var field = t.Field(i)
		g.renderStructFieldMarshaller(&field)
		if i != nf-1 {
			g.fprintf(`
	b.WriteByte(',')
`)
		}
	}
	g.fprintf(`
	b.WriteByte(')')
`)
}

func (g *generator) renderNullableStructFieldMarshaller(field *reflect.StructField, underlying reflect.Type) {
	panic("TODO: nullable type unmarshaller")
}

func (g *generator) renderSimpleToString(t reflect.Type, name string) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		g.fprintf(`
		var thing = fmt.Sprint(%s)
`, name)
	case reflect.Bool:
		g.fprintf(`
		var thing string
		if %s {
			thing = "t"
		} else {
			thing = "f"
		}
`, name)
	case reflect.String:
		g.fprintf(`
		var thing = strconv.Quote(%s)
`, name)
	case reflect.Ptr:
		panic("TODO: pointer type unmarshaller")
	case reflect.Slice:
		g.addImport("strconv")
		g.insertTypeForFurtherProcessing(t)
		var elem = t.Elem()
		g.fprintf(`
		var thing string
		var thingBuilder strings.Builder
		thingBuilder.WriteString("{")
		for i := range %s {
`, name)
		g.renderSimpleToString(elem, fmt.Sprintf("%s[i]", name))
		g.fprintf(`
			thingBuilder.WriteString(thing)
			if i != len(%s)-1 {
				thingBuilder.WriteString(",")
			}
		}
		thingBuilder.WriteString("}")
		thing = strconv.Quote(thingBuilder.String())
`, name)
	case reflect.Struct:
		g.addImport("strconv")
		g.insertTypeForFurtherProcessing(t)
		g.fprintf(`
		var thing string
		{
			var valued, e = %s.Value()
			if e != nil {
				return nil, e
			}
			switch val := valued.(type) {
			case string:
				thing = strconv.Quote(val)
			case nil:
				// do nothing since why
			default:
				panic("TODO: bad thing")
			}
		}
`, name)
	default:
		var msg = fmt.Sprintf("Unsupported type for unmasrhalling: kind - %q, type - %q on field %q", t.Kind(), t, name)
		panic(msg)

	}
}
func (g *generator) renderStructFieldMarshaller(field *reflect.StructField) {
	if !field.IsExported() {
		return
	}
	if nullable := isNullable(field.Type); nullable != nil {
		g.renderNullableStructFieldMarshaller(field, nullable)
		return
	}

	g.fprintf("\t{\n")
	g.renderSimpleToString(field.Type, fmt.Sprintf("v.%s", field.Name))
	g.fprintf(`
		b.WriteString(thing)
	}
	`)
}
