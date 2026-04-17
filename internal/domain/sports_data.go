package domain

import (
	"encoding/json"
	"time"
)

// ActivityType 运动类型枚举
type ActivityType string

const (
	ActivityRunning  ActivityType = "running"
	ActivityCycling  ActivityType = "cycling"
	ActivityStrength ActivityType = "strength" // 力量训练/健身
	ActivitySwimming ActivityType = "swimming"
	ActivityGeneral  ActivityType = "general" // 日常活动/无法分类的运动
)

// DailySummary 用户一天的综合运动汇总
// 建议作为 Agent "每日复盘" 的直接输入
type DailySummary struct {
	UserID     string           `json:"user_id" db:"user_id"`
	Date       time.Time        `json:"date" db:"date"`
	TotalCal   int              `json:"total_calories" db:"total_calories"`                   // 当日总消耗
	RestingHR  int              `json:"resting_heart_rate,omitempty" db:"resting_heart_rate"` // 静息心率（评估疲劳的重要指标）
	SleepHours float64          `json:"sleep_hours,omitempty" db:"sleep_hours"`               // 睡眠时长
	Sessions   []WorkoutSession `json:"sessions" db:"-"`                                      // 当天的具体运动记录列表
}

// WorkoutSession 单次运动会话记录
type WorkoutSession struct {
	SessionID   string           `json:"session_id" db:"session_id"`
	StartTime   time.Time        `json:"start_time" db:"start_time"`
	EndTime     time.Time        `json:"end_time" db:"end_time"`
	Type        ActivityType     `json:"activity_type" db:"activity_type"`
	Environment WeatherCondition `json:"environment,omitempty" db:"environment"` // 运动时的天气上下文

	// 1. 基础运动指标 (所有运动通用)
	Base BaseMetrics `json:"base_metrics" db:"base_metrics"`

	// 2. 高阶生理指标 (通常由专业设备提供，业余小白可能为空)
	Advanced *AdvancedMetrics `json:"advanced_metrics,omitempty" db:"advanced_metrics"`

	// 3. 运动专项指标 (根据 ActivityType 只有其中一项有值)
	RunningData  *RunningMetrics  `json:"running_data,omitempty" db:"running_data"`
	StrengthData *StrengthMetrics `json:"strength_data,omitempty" db:"strength_data"`
}

// BaseMetrics 基础通用指标
type BaseMetrics struct {
	DurationSecs int `json:"duration_secs"`
	Calories     int `json:"calories"`
	AvgHR        int `json:"avg_hr,omitempty"`        // 平均心率
	MaxHR        int `json:"max_hr,omitempty"`        // 最大心率
	FeelingScore int `json:"feeling_score,omitempty"` // 主观疲劳感受 (1-10分)，非常有助于 Agent 调整计划
}

// AdvancedMetrics 高阶/专业生理指标
type AdvancedMetrics struct {
	VO2Max          float64 `json:"vo2_max,omitempty"`                    // 最大摄氧量
	TrainingLoad    int     `json:"training_load,omitempty"`              // 训练负荷 (如 TSS)
	AerobicEffect   float64 `json:"aerobic_effect,omitempty"`             // 有氧训练效果 (0-5.0)
	AnaerobicEffect float64 `json:"anaerobic_effect,omitempty"`           // 无氧训练效果 (0-5.0)
	RecoveryHours   int     `json:"recovery_hours_recommended,omitempty"` // 建议恢复时间
}

// RunningMetrics 跑步专项数据
type RunningMetrics struct {
	DistanceKm     float64 `json:"distance_km"`
	AvgPaceSecs    int     `json:"avg_pace_secs"`              // 平均配速(秒/公里)
	AvgCadence     int     `json:"avg_cadence,omitempty"`      // 平均步频(步/分)
	ElevationGainM int     `json:"elevation_gain_m,omitempty"` // 爬升高度(米)
}

// StrengthMetrics 力量训练专项数据
type StrengthMetrics struct {
	TotalVolumeKg int `json:"total_volume_kg,omitempty"` // 训练总容量(重量x次数x组数)
	SetsCompleted int `json:"sets_completed,omitempty"`  // 完成总组数
}

// ToAgentContext 提取关键信息给 LLM，过滤掉对推理无用的系统级字段
func (ds *DailySummary) ToAgentContext() string {
	// 我们可以构造一个专门用于 Prompt 的精简结构
	type promptData struct {
		Date      string `json:"date"`
		TotalCal  int    `json:"total_calories"`
		RestingHR int    `json:"resting_heart_rate,omitempty"`
		Workouts  []any  `json:"workouts"`
	}

	pd := promptData{
		Date:      ds.Date.Format("2006-01-02"),
		TotalCal:  ds.TotalCal,
		RestingHR: ds.RestingHR,
	}

	for _, session := range ds.Sessions {
		// 精简每次运动的摘要
		workout := map[string]any{
			"type":     session.Type,
			"duration": session.Base.DurationSecs / 60, // 转为分钟更直观
			"avg_hr":   session.Base.AvgHR,
			"feeling":  session.Base.FeelingScore,
		}

		if session.RunningData != nil {
			workout["distance_km"] = session.RunningData.DistanceKm
			workout["pace_secs"] = session.RunningData.AvgPaceSecs
		}
		if session.Advanced != nil {
			workout["training_load"] = session.Advanced.TrainingLoad
		}
		pd.Workouts = append(pd.Workouts, workout)
	}

	// 序列化为 JSON 字符串，这种格式大模型理解起来极其精准
	bytes, _ := json.Marshal(pd)
	return string(bytes)
}
