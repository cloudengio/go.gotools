package impl

type Impl struct{}

func (i *Impl) Write(buf []byte) error {
	return nil
}

func APICall(n int) error {
	return nil
}
