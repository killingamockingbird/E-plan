package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"E-plan/internal/domain"
	"E-plan/internal/repository"
	"E-plan/pkg/util"
)

type planRepo struct {
	reportsCol *mongo.Collection
	plansCol   *mongo.Collection
}

func NewPlanRepo(db *mongo.Database) repository.PlanRepository {
	return &planRepo{
		reportsCol: db.Collection("analysis_reports"),
		plansCol:   db.Collection("training_plans"),
	}
}

func (r *planRepo) SaveReport(ctx context.Context, report *domain.AnalysisReport) error {
	if report.ID == "" {
		report.ID = util.GenerateMongoID()
	}
	doc := repository.AnalysisReportDoc{
		ID:          report.ID,
		UserID:      report.UserID,
		Type:        string(report.Type),
		PeriodStart: report.PeriodStart,
		PeriodEnd:   report.PeriodEnd,
		BodyStatus:  report.BodyStatus,
		Summary:     report.Summary,
		Highlights:  report.Highlights,
		Warnings:    report.Warnings,
		CreatedAt:   time.Now(),
	}
	_, err := r.reportsCol.InsertOne(ctx, doc)
	return err
}

func (r *planRepo) SavePlan(ctx context.Context, plan *domain.TrainingPlan) error {
	if plan.ID == "" {
		plan.ID = util.GenerateMongoID()
	}
	doc := repository.TrainingPlanDoc{
		ID:              plan.ID,
		UserID:          plan.UserID,
		TargetDate:      plan.TargetDate,
		ActivityType:    string(plan.ActivityType),
		Title:           plan.Title,
		Instructions:    plan.Instructions,
		TargetMetrics:   plan.TargetMetrics,
		Reasoning:       plan.Reasoning,
		ForecastWeather: plan.ForecastWeather,
		CreatedAt:       time.Now(),
	}
	_, err := r.plansCol.InsertOne(ctx, doc)
	return err
}

func (r *planRepo) GetPlanByDate(ctx context.Context, userID string, targetDate time.Time) (*domain.TrainingPlan, error) {
	// 查找某天开始到结束的计划 (假设 targetDate 已经 truncate 到当天的 00:00)
	nextDay := targetDate.AddDate(0, 0, 1)
	filter := bson.M{
		"user_id":     userID,
		"target_date": bson.M{"$gte": targetDate, "$lt": nextDay},
	}

	var doc repository.TrainingPlanDoc
	err := r.plansCol.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 当天没有计划，返回 nil 而不是 error
		}
		return nil, err
	}

	return &domain.TrainingPlan{
		ID:              doc.ID,
		UserID:          doc.UserID,
		TargetDate:      doc.TargetDate,
		ActivityType:    domain.ActivityType(doc.ActivityType),
		Title:           doc.Title,
		Instructions:    doc.Instructions,
		TargetMetrics:   doc.TargetMetrics,
		Reasoning:       doc.Reasoning,
		ForecastWeather: doc.ForecastWeather,
		CreatedAt:       doc.CreatedAt,
	}, nil
}
