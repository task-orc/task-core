package task

import "github.com/google/uuid"

type Identity struct {
	ID          string
	Name        string
	Description string
}

func (i Identity) GetInfo() Identity {
	return i
}

func (i Identity) GenerateID(prefix string) Identity {
	return Identity{
		ID:          prefix + uuid.New().String(),
		Name:        i.Name,
		Description: i.Description,
	}
}

type IdentityInfo interface {
	GetInfo() Identity
}
