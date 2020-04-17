package functions

import (
	"bytes"
	"io/ioutil"
)

func Empty() {
}

func HasCall() {
	// nologcall:

	// Comment before.
	ioutil.ReadFile("x") // Comment on the same line.
	// Comment after.
}

func HasDefer() {
	// Comment before.
	defer ioutil.ReadFile("x") // Comment on the same line.
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
	x := ioutil.ReadFile
	x("x")

	y := func() func() {
		return func() {}
	}

	y()()
}
