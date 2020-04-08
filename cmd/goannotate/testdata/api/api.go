package api

type API interface {
	Write([]byte) error
}
