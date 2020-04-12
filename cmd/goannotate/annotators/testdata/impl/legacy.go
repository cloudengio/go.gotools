package impl

import "cloudeng.io/go/cmd/goannotate/annotators/testdata/apilog"

type Legacy struct{}

func (i *Legacy) Write(buf []byte) error {
	defer apilog.LogCallfLegacy(nil, "buf=%v...", buf)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
	return nil
}

func APILegacy(n int) error {
	defer apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
	return nil
}

func APILegacyNonDefer(n int) error {
	apilog.LogCallfLegacy(nil, "n=%v...", n)(nil, "") // gologcop: DO NOT EDIT, MUST BE FIRST STATEMENT
	return nil
}
