package mongodb

import (
	"E-plan/pkg/util"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"E-plan/internal/domain"
	"E-plan/internal/repository"
)

type mongoSportsDataRepo struct {
	collection *mongo.Collection
}

// NewSportsDataRepo 构造函数
func NewSportsDataRepo(db *mongo.Database) repository.SportsDataRepository {
	return &mongoSportsDataRepo{
		// 所有的每日数据都存放在这个集合里
		collection: db.Collection("daily_summaries"),
	}
}

// SaveDailySummary 保存每日汇总 (Upsert 逻辑)
func (r *mongoSportsDataRepo) SaveDailySummary(ctx context.Context, summary *domain.DailySummary) error {
	// 1. Domain 模型转换为 Mongo Doc
	doc := repository.DailySummaryDoc{
		ID:         util.GenerateMongoID(),
		UserID:     summary.UserID,
		Date:       summary.Date,
		TotalCal:   summary.TotalCal,
		RestingHR:  summary.RestingHR,
		SleepHours: summary.SleepHours,
		Sessions:   make([]repository.WorkoutSessionDoc, 0, len(summary.Sessions)),
	}

	for _, s := range summary.Sessions {
		doc.Sessions = append(doc.Sessions, repository.WorkoutSessionDoc{
			SessionID:    s.SessionID,
			Type:         string(s.Type),
			StartTime:    s.StartTime,
			EndTime:      s.EndTime,
			Environment:  s.Environment,
			Base:         s.Base,
			Advanced:     s.Advanced,
			RunningData:  s.RunningData,
			StrengthData: s.StrengthData,
		})
	}

	// 2. 构造查询条件 (相同用户、同一天的数据进行覆盖/更新)
	filter := bson.M{
		"user_id": summary.UserID,
		"date":    summary.Date,
	}

	// 3. 设置 Update 操作，开启 Upsert (不存在则插入，存在则全量更新)
	update := bson.M{"$set": doc}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetRecentSummaries 获取近期数据，供给 Agent 作为长期记忆
func (r *mongoSportsDataRepo) GetRecentSummaries(ctx context.Context, userID string, since time.Time) ([]*domain.DailySummary, error) {
	// 构造查询：user_id 匹配，且 date >= since
	filter := bson.M{
		"user_id": userID,
		"date":    bson.M{"$gte": since},
	}

	// 按日期升序排列，这样交给 LLM 时，时间线是顺的
	opts := options.Find().SetSort(bson.M{"date": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []repository.DailySummaryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	// 将查出来的 Mongo Doc 转换回 Domain 模型给 Agent 使用
	var summaries []*domain.DailySummary
	for _, doc := range docs {
		summary := &domain.DailySummary{
			UserID:     doc.UserID,
			Date:       doc.Date,
			TotalCal:   doc.TotalCal,
			RestingHR:  doc.RestingHR,
			SleepHours: doc.SleepHours,
		}

		for _, s := range doc.Sessions {
			summary.Sessions = append(summary.Sessions, domain.WorkoutSession{
				SessionID:    s.SessionID,
				Type:         domain.ActivityType(s.Type),
				StartTime:    s.StartTime,
				EndTime:      s.EndTime,
				Environment:  s.Environment,
				Base:         s.Base,
				Advanced:     s.Advanced,
				RunningData:  s.RunningData,
				StrengthData: s.StrengthData,
			})
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
