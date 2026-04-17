package domain

import "time"

// ExperienceLevel 用户的运动经验阶段
type ExperienceLevel string

const (
	LevelBeginner   ExperienceLevel = "beginner"     // 运动小白 (需要多鼓励，防受伤，科普常识)
	LevelEnthusiast ExperienceLevel = "enthusiast"   // 爱好者 (有一定基础，追求规律性或小突破)
	LevelAdvanced   ExperienceLevel = "advanced"     // 进阶/精英 (追求成绩，关注心率区间、Lactate Threshold)
	LevelPro        ExperienceLevel = "professional" // 专业运动员 (极度关注数据细节、恢复、周期化训练)
)

// PrimaryFocus 用户的主要运动偏好/类型
type PrimaryFocus string

const (
	FocusRunning PrimaryFocus = "running" // 跑步偏好
	FocusFitness PrimaryFocus = "fitness" // 健身/力量训练偏好
	FocusCycling PrimaryFocus = "cycling" // 骑行偏好
	FocusGeneral PrimaryFocus = "general" // 综合健康/减脂偏好 (不限定特定运动)
)

// User 代表系统中的一个用户
type User struct {
	ID        string          `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	City      string          `json:"city" db:"city"`
	Level     ExperienceLevel `json:"level" db:"level"`         // 经验阶段
	Focus     PrimaryFocus    `json:"focus" db:"focus"`         // 运动偏好
	Target    string          `json:"target" db:"target"`       // 具体近期目标，如："无伤跑完5公里" 或 "卧推达到100kg"
	BaseInfo  BasicInfo       `json:"base_info" db:"base_info"` // 身体基础数据
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// BasicInfo 用户的基本生理信息
type BasicInfo struct {
	Age    int     `json:"age"`
	Height float64 `json:"height"` // cm
	Weight float64 `json:"weight"` // kg
	Gender string  `json:"gender"` // 性别 (在基础代谢计算中通常需要)
}

// ToLLMPersona 动态生成给 LLM 的 System Prompt 人设注入
func (u *User) ToLLMPersona() string {
	levelDesc := ""
	switch u.Level {
	case LevelBeginner:
		levelDesc = "一名刚开始接触运动的『运动小白』。你需要多使用鼓励的语气，用最通俗易懂的语言解释运动常识，计划要极度循序渐进，首要原则是【绝对避免受伤】和【培养运动习惯】。"
	case LevelEnthusiast:
		levelDesc = "一名有一定基础的『运动爱好者』。你需要提供有建设性的进阶建议，可以适当使用基础的运动术语，帮助其突破瓶颈。"
	case LevelAdvanced, LevelPro:
		levelDesc = "一名『硬核/专业运动者』。请务必使用严谨的运动科学术语（如VO2Max、TSS、步幅、触地时间等），计划需包含热身、主课表、冷身，并严格考量疲劳度与超量恢复机制。"
	}

	focusDesc := ""
	switch u.Focus {
	case FocusRunning:
		focusDesc = "他/她目前特别关注【跑步】领域。"
	case FocusFitness:
		focusDesc = "他/她目前特别关注【力量与健身房训练】领域。"
	case FocusCycling:
		focusDesc = "他/她目前特别关注【骑行】领域。"
	case FocusGeneral:
		focusDesc = "他/她目前特别关注【综合健康】领域。"

	}

	return "你的服务对象是：" + levelDesc + " " + focusDesc + " 当前的短期核心目标是：[" + u.Target + "]。"
}
