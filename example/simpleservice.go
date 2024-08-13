package example

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repository PersonRepository
}

func (s *Service) CreateNewPerson(ctx context.Context, person Person) (Person, error) {
	return s.repository.Save(ctx, person)
}

func (s *Service) FindAllPersons(ctx context.Context, person Person) ([]Person, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) DeletePerson(ctx context.Context, id primitive.ObjectID) error {
	return s.repository.DeleteById(ctx, id)
}
