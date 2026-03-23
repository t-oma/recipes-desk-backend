package mongorepo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"recipes-desk/internal/modules/tags/domain"
	"recipes-desk/internal/modules/tags/domain/entity"
	"recipes-desk/internal/modules/tags/domain/ports"
)

const collectionName = "tags"

// Verify TagRepository implements ports.TagRepository.
var _ ports.TagRepository = (*TagRepository)(nil)

type TagRepository struct {
	collection *mongo.Collection
}

func NewTags(db *mongo.Database) *TagRepository {
	return &TagRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *TagRepository) InitIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return r.wrapError(err, "create tag indexes")
	}
	return nil
}

func (r *TagRepository) Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	tagModel := tagModelFromDomain(tag)
	tagModel.prepareForInsert()

	_, err := r.collection.InsertOne(ctx, tagModel)
	if err != nil {
		return nil, r.wrapError(err, "create tag")
	}

	return tagModel.toDomain()
}

func (r *TagRepository) FindByID(ctx context.Context, id string) (*entity.Tag, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	var model tagModel
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		return nil, r.wrapError(err, "find tag by id")
	}

	return model.toDomain()
}

func (r *TagRepository) Search(
	ctx context.Context,
	query string,
	skip, limit int64,
) ([]entity.Tag, int64, error) {
	filter := bson.M{
		"name": bson.M{
			"$regex":   query,
			"$options": "i", // case-insensitive
		},
	}

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "slug", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, r.wrapError(err, "find tags by name")
	}
	defer cursor.Close(ctx)

	var tagModels []tagModel
	if err = cursor.All(ctx, &tagModels); err != nil {
		return nil, 0, r.wrapError(err, "decode tags")
	}

	tags := make([]entity.Tag, len(tagModels))
	for i, model := range tagModels {
		tag, err := model.toDomain()
		if err != nil {
			return nil, 0, err
		}
		tags[i] = *tag
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, r.wrapError(err, "count search results")
	}

	return tags, total, nil
}

func (r *TagRepository) Exists(ctx context.Context, slug string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"slug": slug})
	if err != nil {
		return false, r.wrapError(err, "check tag exists")
	}
	return count > 0, nil
}

// wrapError converts MongoDB errors to domain errors.
func (r *TagRepository) wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.ErrNotFound
	}
	if mongo.IsTimeout(err) {
		return fmt.Errorf("%w: %s: %w", domain.ErrTimeout, operation, err)
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %s: %w", domain.ErrConflict, operation, err)
	}

	return fmt.Errorf("%w: %s: %w", domain.ErrDatabase, operation, err)
}
