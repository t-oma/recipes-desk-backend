package mongorepo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/tags/domain/entity"
	"recipes-desk/internal/modules/tags/domain/valueobject"
)

type tagModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` //nolint:tagliatelle // mongo _id field
	Name      string             `bson:"name"`
	Slug      string             `bson:"slug"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

func (m *tagModel) setTimestamps() {
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}

func (m *tagModel) setID() {
	m.ID = primitive.NewObjectID()
}

func (m *tagModel) prepareForInsert() {
	m.setTimestamps()
	if m.ID.IsZero() {
		m.setID()
	}
}

func (m *tagModel) toDomain() (*entity.Tag, error) {
	id, err := valueobject.NewTagID(m.ID.Hex())
	if err != nil {
		return nil, err
	}
	name, err := valueobject.NewTagName(m.Name)
	if err != nil {
		return nil, err
	}
	slug, err := valueobject.NewTagSlug(m.Slug)
	if err != nil {
		return nil, err
	}

	tag := entity.TagFrom(
		id,
		name,
		slug,
		m.CreatedAt,
		m.UpdatedAt,
	)

	return tag, nil
}

func tagModelFromDomain(tag *entity.Tag) *tagModel {
	var id primitive.ObjectID
	if tag.ID().String() != "" {
		var err error
		id, err = primitive.ObjectIDFromHex(tag.ID().String())
		if err != nil {
			id = primitive.NilObjectID
		}
	}

	return &tagModel{
		ID:        id,
		Name:      tag.Name().String(),
		Slug:      tag.Slug().String(),
		CreatedAt: tag.CreatedAt(),
		UpdatedAt: tag.UpdatedAt(),
	}
}
