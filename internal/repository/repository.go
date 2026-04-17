package repository

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"time"

	"E-plan/internal/domain"
)

// UserRepository 用户信息仓储
type UserRepository interface {
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	SaveUser(ctx context.Context, user *domain.User) error
}

// SportsDataRepository 运动数据仓储 (Agent 的长期客观记忆)
type SportsDataRepository interface {
	SaveDailySummary(ctx context.Context, summary *domain.DailySummary) error
	GetRecentSummaries(ctx context.Context, userID string, since time.Time) ([]*domain.DailySummary, error)
}

// PlanRepository 分析报告与计划仓储 (Agent 的输出成果)
type PlanRepository interface {
	SaveReport(ctx context.Context, report *domain.AnalysisReport) error
	SavePlan(ctx context.Context, plan *domain.TrainingPlan) error
	GetPlanByDate(ctx context.Context, userID string, targetDate time.Time) (*domain.TrainingPlan, error)
}
type HistoryRepository interface {
	LoadHistory(ctx context.Context, sessionID string) ([]*schema.Message, error)
	SaveHistory(ctx context.Context, sessionID string, userID string, messages []*schema.Message) error
}
