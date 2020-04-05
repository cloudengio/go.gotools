package examples

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"strings"
	"unsafe"
)

func Noop() {}

func NumericBasic(a int, b uint8, c uint16, d float32, e float64, f float32, g complex128, h complex64) (ar int, br uint8, cr uint16, dr float32, er float64, fr float32, gr complex128, hr complex64) {
	return
}

func OtherBasic(a bool, b string, c rune, d int32, e byte) (ar bool, br string, cr rune, dr int32, er byte) {
	return
}

func Error(a error) (err error) {
	return nil
}

type myStruct struct{}

func Pointers(a *int, b *myStruct, c unsafe.Pointer) {}

func Unknown(a rand.Rand, b io.Reader) (int, error)

func Stringer(a bytes.Buffer, b *strings.Builder) {}

func Anon(_ int, _ string)

type NamedInt int
type NamedString string

func Named(a NamedInt, b NamedString) {}

func SlicesMaps(a []int, b []NamedInt, c map[string]NamedString) {}

func Variadic(a int, b ...int) {}

func Group(a, b int) (ar, br string) {
	return
}

func WithContextOnly(ctx context.Context) {}

func WithContext(ctx context.Context, a int) {}

func WithContextAnon(ctx context.Context, _ int, c bool) {}
