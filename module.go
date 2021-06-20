package camogo

// Module can be used to register a set of services
type Module interface {
	Register(*Registrar) error
}
