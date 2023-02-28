package uuidgen

import "github.com/google/uuid"

var global Generator = NewGoogleUUID()

//go:generate mockery --name Generator --with-expecter
type Generator interface {
	NewString() string
}

type GoogleUUID struct{}

func NewGoogleUUID() GoogleUUID {
	return GoogleUUID{}
}

func (GoogleUUID) NewString() string {
	return uuid.NewString()
}

func SetGlobal(gen Generator) {
	global = gen
}

func NewString() string {
	return global.NewString()
}
