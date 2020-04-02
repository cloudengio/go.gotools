package data

func private(a int) error {
	return nil
}

func Fn1() error {
	return nil
}

type rcv struct{}

func (r *rcv) Fn1() {

}
