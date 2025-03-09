package meta

import (
	"reflect"
	"strings"
)

func isNullable(t reflect.Type) reflect.Type {

	if t.Kind() != reflect.Struct {
		return nil
	}

	if !strings.HasPrefix(t.Name(), "Nullable") {
		return nil
	}

	var ok bool
	if _, ok = t.FieldByName("Valid"); !ok {
		return nil
	}
	var validField reflect.StructField
	if validField, ok = t.FieldByName("Item"); !ok {
		return nil
	}

	var neededT reflect.Type
	if validField.Type.Kind() == reflect.Pointer {
		neededT = validField.Type.Elem()
	} else {
		neededT = validField.Type
	}
	if neededT.Name() != strings.TrimPrefix(t.Name(), "Nullable") {
		return nil
	}
	return neededT
}

func (g *generator) renderNullable(t reflect.Type) (ok bool) {

	var neededT reflect.Type
	if neededT = isNullable(t); neededT == nil {
		return false
	}
	g.insertTypeForFurtherProcessing(neededT)

	g.fprintf(`
func (v *%s) Scan(thing any) (err error) {
	switch thval := thing.(type) {
	case nil:
		v.Valid = false
		return
	case string:
		if len(thval) == 0 {
			v.Valid = false
			return
		}
		v.Valid = true
		v.Item = new(%s)
		return v.Item.Scan(thval)
	default:
		panic("bad type from posthres: " + fmt.Sprint(thval))
	}
}

func (v *%s) Value() (t driver.Value, err error) {
	if v.Valid {
		return v.Item.Value()
	} else {
		return nil, nil
	}
}
`, t.Name(), neededT.Name(), t.Name())
	return true
}
