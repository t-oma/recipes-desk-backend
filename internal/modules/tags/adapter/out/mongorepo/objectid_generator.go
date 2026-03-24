package mongorepo

import "go.mongodb.org/mongo-driver/bson/primitive"

type ObjectIDGenerator struct{}

func (g ObjectIDGenerator) Generate() string {
	return primitive.NewObjectID().Hex()
}

func (g ObjectIDGenerator) Validate(id string) error {
	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return nil
}
