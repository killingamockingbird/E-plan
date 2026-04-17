package mongodb

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/cloudwego/eino/schema"
)

// ChatHistoryDoc 聊天记录文档
type ChatHistoryDoc struct {
	SessionID string    `bson:"_id"` // 使用 userID_date 作为主键
	UserID    string    `bson:"user_id"`
	Date      string    `bson:"date"`
	Messages  string    `bson:"messages"` // 直接存 JSON 序列化后的 schema.Message 数组
	UpdatedAt time.Time `bson:"updated_at"`
}

type historyRepo struct {
	col *mongo.Collection
}

func NewHistoryRepo(db *mongo.Database) *historyRepo {
	return &historyRepo{col: db.Collection("chat_histories")}
}

// LoadHistory 加载历史对话
func (r *historyRepo) LoadHistory(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	var doc ChatHistoryDoc
	err := r.col.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return make([]*schema.Message, 0), nil // 没查到说明是新会话，返回空切片
		}
		return nil, err
	}

	var messages []*schema.Message
	if err := json.Unmarshal([]byte(doc.Messages), &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// SaveHistory 保存/更新对话
func (r *historyRepo) SaveHistory(ctx context.Context, sessionID string, userID string, messages []*schema.Message) error {
	msgBytes, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	doc := ChatHistoryDoc{
		SessionID: sessionID,
		UserID:    userID,
		Date:      time.Now().Format("2006-01-02"),
		Messages:  string(msgBytes),
		UpdatedAt: time.Now(),
	}

	opts := options.Update().SetUpsert(true)
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": sessionID}, bson.M{"$set": doc}, opts)
	return err
}
