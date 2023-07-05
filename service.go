package discovery

type Service interface {
	Name() string
	Addr() string
}
