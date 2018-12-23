package decompile

import "github.com/pkg/errors"

// Errorf formats according to a format specifier and returns the string as a
// value that satisfies error. The error is also passed to the error handler of
// the generator.
func (gen *Generator) Errorf(format string, a ...interface{}) error {
	err := errors.Errorf(format, a...)
	gen.eh(err)
	return err
}
