package data

type Ifc1 interface {
	M1()
	M2(string)
}

type hidden interface {
	M1(int) error
}

type Ifc2 interface {
	M3(int) error
}

type Ifc3 interface {
	Ifc1
	Ifc2
}

type StructExampleIgnored struct {
	Field Ifc3
}

var IgnoredVariable Ifc2
