package annotators

import (
	"runtime"

	"cloudeng.io/go/locate"
)

func concurrencyOpt(val int) locate.Option {
	if val == 0 {
		val = runtime.NumCPU()
	}
	return locate.Concurrency(val)
}
