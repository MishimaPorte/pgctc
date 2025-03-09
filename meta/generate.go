package meta

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"slices"
)

type generator struct {
	label            int
	out              io.Writer
	importsNeeded    []string
	needMoreTypes    []reflect.Type
	seenAlreadyTypes []reflect.Type
}

func (g *generator) getLabel() string {
	g.label++
	return fmt.Sprintf("label%d", g.label)
}

func (g *generator) fprintf(format string, a ...any) (n int, err error) {
	return fmt.Fprintf(g.out, format, a...)
}

func (g *generator) addImport(imp ...string) {
	for _, item := range imp {
		if !slices.Contains(g.importsNeeded, item) {
			g.importsNeeded = append(g.importsNeeded, item)
		}
	}
}

func GenerateFor(packageName string, w io.WriteCloser, typ ...reflect.Type) (err error) {
	var b bytes.Buffer
	var gen = generator{
		out: &b,
	}
	gen.addImport(
		"unsafe",
		"database/sql/driver",
		"fmt",
	)

	for _, i := range typ {
		gen.seenAlreadyTypes = append(gen.seenAlreadyTypes, i)
		if ok := gen.renderNullable(i); ok {
			continue
		}
		gen.renderDriverScan(i)
		gen.renderDriverValue(i)
	}
	gen.renderTails()

	gen.out = w
	gen.renderFileHeader(packageName)
	io.Copy(gen.out, &b)
	return nil
}

func (g *generator) renderTails() {
	var l = len(g.needMoreTypes)
	for _, tp := range g.needMoreTypes {
		g.renderDriverScan(tp)
		g.renderDriverValue(tp)
	}
	if len(g.needMoreTypes) != l {
		g.needMoreTypes = g.needMoreTypes[l:]
		g.renderTails()
	}
}

func (g *generator) insertTypeForFurtherProcessing(t reflect.Type) {
	if !slices.Contains(g.seenAlreadyTypes, t) {
		if t.Kind() == reflect.Slice && t.Name() == "" {
			// do not need to emit code for unnamed slices, since they cannot be receivers
			// for names slices, though, we need.
			return
		}
		g.seenAlreadyTypes = append(g.seenAlreadyTypes, t)
		g.needMoreTypes = append(g.needMoreTypes, t)
	}
}
