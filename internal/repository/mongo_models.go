package repository

import (
	"E-plan/internal/domain"
	"time"
)

// UserDoc 用户文档
type UserDoc struct {
	ID        string           `bson:"_id"` // 直接存储 Hex 字符串作为主键
	Name      string           `bson:"name"`
	Level     string           `bson:"level"`
	Focus     string           `bson:"focus"`
	Target    string           `bson:"target"`
	BaseInfo  domain.BasicInfo `bson:"base_info"`
	CreatedAt time.Time        `bson:"created_at"`
}

// DailySummaryDoc 每日运动汇总文档
type DailySummaryDoc struct {
	ID         string              `bson:"_id"`
	UserID     string              `bson:"user_id"`
	Date       time.Time           `bson:"date"`
	TotalCal   int                 `bson:"total_cal"`
	RestingHR  int                 `bson:"resting_hr,omitempty"`
	SleepHours float64             `bson:"sleep_hours,omitempty"`
	Sessions   []WorkoutSessionDoc `bson:"sessions"`
}

// WorkoutSessionDoc 内嵌单次运动记录
type WorkoutSessionDoc struct {
	SessionID    string                  `bson:"session_id"`
	Type         string                  `bson:"type"`
	StartTime    time.Time               `bson:"start_time"`
	EndTime      time.Time               `bson:"end_time"`
	Environment  domain.WeatherCondition `bson:"environment,omitempty"`
	Base         domain.BaseMetrics      `bson:"base_metrics"`
	Advanced     *domain.AdvancedMetrics `bson:"advanced_metrics,omitempty"`
	RunningData  *domain.RunningMetrics  `bson:"running_data,omitempty"`
	StrengthData *domain.StrengthMetrics `bson:"strength_data,omitempty"`
}

// AnalysisReportDoc 分析报告文档
type AnalysisReportDoc struct {
	ID          string    `bson:"_id"`
	UserID      string    `bson:"user_id"`
	Type        string    `bson:"type"`
	PeriodStart time.Time `bson:"period_start"`
	PeriodEnd   time.Time `bson:"period_end"`
	BodyStatus  string    `bson:"body_status"`
	Summary     string    `bson:"summary"`
	Highlights  []string  `bson:"highlights"`
	Warnings    []string  `bson:"warnings"`
	CreatedAt   time.Time `bson:"created_at"`
}

// TrainingPlanDoc 训练计划文档
type TrainingPlanDoc struct {
	ID              string                  `bson:"_id"`
	UserID          string                  `bson:"user_id"`
	TargetDate      time.Time               `bson:"target_date"`
	ActivityType    string                  `bson:"activity_type"`
	Title           string                  `bson:"title"`
	Instructions    string                  `bson:"instructions"`
	TargetMetrics   domain.PlanTargets      `bson:"target_metrics"`
	Reasoning       string                  `bson:"reasoning"`
	ForecastWeather domain.WeatherCondition `bson:"forecast_weather"`
	CreatedAt       time.Time               `bson:"created_at"`
}
