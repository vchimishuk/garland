package source

type Source interface {
	Name() string
	Value() (float64, bool, error)
}
