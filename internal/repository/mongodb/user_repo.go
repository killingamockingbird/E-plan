package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"E-plan/internal/domain"
	"E-plan/internal/repository"
	"E-plan/pkg/util"
)

type userRepo struct {
	col *mongo.Collection
}

func NewUserRepo(db *mongo.Database) repository.UserRepository {
	return &userRepo{col: db.Collection("users")}
}

func (r *userRepo) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	var doc repository.UserDoc
	err := r.col.FindOne(ctx, bson.M{"_id": userID}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	// 转换为 Domain 模型
	return &domain.User{
		ID:        doc.ID,
		Name:      doc.Name,
		Level:     domain.ExperienceLevel(doc.Level),
		Focus:     domain.PrimaryFocus(doc.Focus),
		Target:    doc.Target,
		BaseInfo:  doc.BaseInfo,
		CreatedAt: doc.CreatedAt,
	}, nil
}

func (r *userRepo) SaveUser(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = util.GenerateMongoID()
	}

	doc := repository.UserDoc{
		ID:        user.ID,
		Name:      user.Name,
		Level:     string(user.Level),
		Focus:     string(user.Focus),
		Target:    user.Target,
		BaseInfo:  user.BaseInfo,
		CreatedAt: user.CreatedAt,
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{"$set": doc}, opts)
	return err
}
