package main

import (
	"os"
	"reflect"

	"git.apsolutions.ru/aps/Internal/streaming-platform/source-code/libs/pg-composite-parser-gen.git/example/types"
	"git.apsolutions.ru/aps/Internal/streaming-platform/source-code/libs/pg-composite-parser-gen.git/meta"
)

func main() {
	if f, err := os.Create(os.Args[1]); err != nil {
		panic(err.Error())
	} else {
		meta.GenerateFor(
			"types",
			f,
			reflect.TypeOf(types.Containers{}),
			reflect.TypeOf(types.Props{}),
			reflect.TypeOf(types.NullableHTTPParams{}),
		)
	}
}
