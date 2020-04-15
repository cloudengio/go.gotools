// Copyright 2020 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

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
