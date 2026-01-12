package interfaces

// скорее всего названия методов  - поменяются !!!!!
type Job interface {
	GetID() string
	Complete(data interface{}, err error)
}
