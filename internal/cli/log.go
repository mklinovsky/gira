package cli

import (
	"fmt"
	"io"
)

type env struct {
	stdout io.Writer
	stderr io.Writer
}

func (e *env) success(format string, args ...any) {
	fmt.Fprintf(e.stdout, "✅ %s\n", fmt.Sprintf(format, args...))
}

func (e *env) info(format string, args ...any) {
	fmt.Fprintf(e.stdout, "👍 %s\n", fmt.Sprintf(format, args...))
}

func (e *env) failure(err error) {
	fmt.Fprintf(e.stderr, "❌ %v\n", err)
}

func (e *env) failuref(format string, args ...any) {
	fmt.Fprintf(e.stderr, "❌ %s\n", fmt.Sprintf(format, args...))
}
