package pggen

import (
	"fmt"
	"strings"
)

// This type is madness
type ErrorSet struct {
	name   string
	errors []error
}

func (e ErrorSet) Error() string {
	// TODO: cache builders somehow
	var b strings.Builder
	for _, err := range e.errors {
		b.WriteString(err.Error())
		b.WriteByte('\n')
	}
	return b.String()
}

func (e *ErrorSet) AddError(err error) {
	e.errors = append(e.errors, err)
}

func (e *ErrorSet) Errorf(fmts string, datas ...any) {
	e.errors = append(e.errors, fmt.Errorf(fmts, datas...))
}

func (e *ErrorSet) AddErrors(err ...error) {
	e.errors = append(e.errors, err...)
}

func (e *ErrorSet) IsNil() bool {
	return len(e.errors) == 0
}
