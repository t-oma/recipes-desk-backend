package ports

type IDGenerator interface {
	Generate() string
	Validate(id string) error
}
