package domain

import (
	"time"
)

// ReportType 报告类型
type ReportType string

const (
	ReportDaily  ReportType = "daily"  // 每日复盘
	ReportWeekly ReportType = "weekly" // 周度总结
)

// AnalysisReport AI 生成的运动分析报告（针对已发生的历史数据）
type AnalysisReport struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Type        ReportType `json:"type" db:"type"`
	PeriodStart time.Time  `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time  `json:"period_end" db:"period_end"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`

	// --- AI 核心分析内容 ---

	// 状态评估：例如 "恢复良好", "处于疲劳期", "有氧能力提升"
	BodyStatus string `json:"body_status" db:"body_status"`

	// 综合评价：给用户的自然语言总结（不同经验级别的用户，此处的专业度不同）
	Summary string `json:"summary" db:"summary"`

	// 亮点与不足：结构化的反馈，方便前端 UI 渲染列表
	Highlights []string `json:"highlights" db:"highlights"`
	Warnings   []string `json:"warnings" db:"warnings"` // 如："步频过低可能导致膝盖受力过大"
}

// TrainingPlan AI 生成的未来训练计划
type TrainingPlan struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	TargetDate time.Time `json:"target_date" db:"target_date"` // 计划执行的具体日期
	CreatedAt  time.Time `json:"created_at" db:"created_at"`

	// --- 计划核心内容 ---

	ActivityType ActivityType `json:"activity_type" db:"activity_type"` // 继承 sports_data.go 中的枚举

	// 计划的简短标题，如 "轻松恢复跑 5KM" 或 "上肢力量突破"
	Title string `json:"title" db:"title"`

	// 给用户的具体操作指南
	Instructions string `json:"instructions" db:"instructions"`

	// 结构化的目标参数 (方便与实际执行数据做比对)
	TargetMetrics PlanTargets `json:"target_metrics" db:"target_metrics"`

	// --- 决策依据 (Explainability) ---

	// 为什么这么安排？（大模型输出的推理过程，结合了用户的长期目标）
	Reasoning string `json:"reasoning" db:"reasoning"`

	// 生成此计划时，参考的目标日期的天气情况
	ForecastWeather WeatherCondition `json:"forecast_weather" db:"forecast_weather"`
}

// PlanTargets 计划的目标参数 (组合使用，非必需字段用 omitempty)
type PlanTargets struct {
	DistanceKm    float64 `json:"distance_km,omitempty"`
	DurationSecs  int     `json:"duration_secs,omitempty"`
	TargetHRZone  string  `json:"target_hr_zone,omitempty"`   // 目标心率区间，如 "Zone 2" (适合进阶/专业)
	TargetPace    int     `json:"target_pace_secs,omitempty"` // 目标配速 (适合跑者)
	TotalVolumeKg int     `json:"total_volume_kg,omitempty"`  // 目标总容量 (适合健身者)
}
