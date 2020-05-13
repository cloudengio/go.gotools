package main

func main() {
}

func InMain() error {
	return nil
}

type rcvr struct{}

func (r *rcvr) InMain() {}
