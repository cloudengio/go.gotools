package impl

type Impl1 struct{}

func (i *Impl1) M1() {

}

func (i *Impl1) M2(string) {

}

type impl2 struct{}

func (i *impl2) M3(int) error {
	return nil

}

type Impl12 struct{}

func (i *Impl12) M1() {

}

func (i *Impl12) M2(string) {

}

func (i *Impl12) M3(int) error {
	return nil

}

type Other struct{}

func (o *Other) M1() {

}

type WithInterfaceField struct {
	A int
	B interface {
		M1()
		M2(string)
	}
}
