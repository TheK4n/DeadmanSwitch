package switcher

type Switcher interface {
	Execute() ([]byte, error)
}