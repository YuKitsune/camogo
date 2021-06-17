package container

type Module interface{
	Register(*Registrar) error
}