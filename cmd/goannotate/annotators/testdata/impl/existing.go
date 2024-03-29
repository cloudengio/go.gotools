package impl

import "cloudeng.io/go/cmd/goannotate/annotators/testdata/apilog"

type Existing struct{}

func (i *Existing) Write(buf []byte) error {
	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.Write", "impl/existing.go:5", "buf[:%d]=...", len(buf))(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
	return nil
}

func APIExisting(n int) error {
	defer apilog.LogCallf(nil, "cloudeng.io/go/cmd/goannotate/annotators/testdata/impl.APIExisting", "impl/existing.go:9", "n=%d", n)(nil, "_=?") // DO NOT EDIT, AUTO GENERATED BY cloudeng.io/go/cmd/goannotate/annotators.AddLogCall#add
	return nil
}

func APINew(n int) error {
	return nil
}

func APIEmptyFunc(n int) {
}

func APINoLogCall(n int) error {
	// nologcall:
	return nil
}
