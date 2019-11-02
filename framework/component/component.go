package component

type Component interface {
	Start() error
	Stop() error
}
