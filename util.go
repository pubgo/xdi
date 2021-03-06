package dix

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

func isError(t reflect.Type) bool {
	return t.Implements(errType)
}

func isDoublePtr(tpy reflect.Type) bool {
	if isElem(tpy) {
		return isElem(tpy.Elem())
	}
	return false
}

func isElem(tye reflect.Type) bool {
	switch tye.Kind() {
	case reflect.Chan, reflect.Map, reflect.Ptr, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func getIndirectType(tye reflect.Type) reflect.Type {
	for isElem(tye) {
		tye = tye.Elem()
	}
	return tye
}

func fPrintln(w io.Writer, a ...interface{}) {
	_, _ = fmt.Fprintln(w, a...)
}

func callerWithFunc(fn reflect.Value) string {
	var _e = runtime.FuncForPC(fn.Pointer())
	var file, line = _e.FileLine(fn.Pointer())

	var buf = &strings.Builder{}
	defer buf.Reset()

	files := strings.Split(file, "/")
	if len(files) > 2 {
		file = strings.Join(files[len(files)-2:], "/")
	}

	buf.WriteString(file)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(line))
	buf.WriteString(" ")

	buf.WriteString(fn.String())
	return buf.String()
}

func equal(x, y []reflect.Value) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if x[i].IsNil() || y[i].IsNil() {
			return false
		}

		if x[i].Pointer() != y[i].Pointer() {
			return false
		}
	}
	return true
}
