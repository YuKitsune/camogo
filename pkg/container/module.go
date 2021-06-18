package container

type RegistrationFunc func (*Registrar) error
type Module interface{
	Register(*Registrar) error
}