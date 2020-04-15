package impl

type Empty struct{}

func (i *Empty) Write(buf []byte) error {
	return nil
}

func APIEmpty(n int) error {
	return nil
}
