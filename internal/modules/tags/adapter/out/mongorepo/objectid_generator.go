package mongorepo

import "go.mongodb.org/mongo-driver/bson/primitive"

type ObjectIDGenerator struct{}

func (g ObjectIDGenerator) Generate() string {
	return primitive.NewObjectID().Hex()
}
