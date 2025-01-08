package functions

import (
	"bytes"
	"os"
)

func Empty() {
}

func HasCall() {
	// nologcall:

	// Comment before.
	os.ReadFile("x") // Comment on the same line.
	// Comment after.
}

func HasDefer() {
	// Comment before.
	defer os.ReadFile("x") // Comment on the same line.
	// Comment after.
	// nologcall:
}

func HasOther() {
	bytes.NewBuffer()
}

func HasOtherDefer() {
	defer bytes.NewBuffer()
}

func Expressions() {
	x := os.ReadFile
	x("x")

	y := func() func() {
		return func() {}
	}

	y()()
}
