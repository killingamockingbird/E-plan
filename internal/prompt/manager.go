package prompt

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"E-plan/internal/domain" // 保持你的模块路径
)

// Manager 负责管理和渲染大模型的 Prompt 模板
type Manager struct {
	systemTemplate *template.Template
}

// NewManager 通过文件路径加载并解析模板
func NewManager(templatePath string) (*Manager, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("读取模板文件失败: %w", err)
	}

	tmpl, err := template.New("system").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("解析模板失败: %w", err)
	}

	return &Manager{
		systemTemplate: tmpl,
	}, nil
}

// BuildSystemPrompt 渲染最终的 System Prompt
// 修改点 1：增加了 pastData 和 forecast 参数，对齐 Agent 编排器
func (m *Manager) BuildSystemPrompt(user *domain.User, pastData []*domain.DailySummary, forecast *domain.WeatherCondition) (string, error) {
	// 初始化模板数据容器
	data := map[string]interface{}{
		"HasPastData": len(pastData) > 0,
		"PastData":    pastData,
		// 默认一些布尔开关，方便在模板里用 {{if .HasUser}} 做条件渲染
		"HasUser":    false,
		"HasWeather": false,
	}

	// 修改点 2：安全的提取用户数据（防 nil panic）
	if user != nil {
		data["HasUser"] = true
		data["User"] = user // 直接将整个对象传进去，模板里可以用 {{.User.Level}} 访问

		// 兼容你原来的扁平化写法
		data["ExperienceLevel"] = string(user.Level)
		data["FocusArea"] = string(user.Focus)
		data["Target"] = user.Target

		// 假设 BaseInfo 存在且不是 nil
		// 如果 BaseInfo 是指针类型，这里还需要多做一层 if user.BaseInfo != nil 的判断
		data["Age"] = user.BaseInfo.Age
		data["Height"] = user.BaseInfo.Height
		data["Weight"] = user.BaseInfo.Weight
	}

	// 修改点 3：注入天气预报数据
	if forecast != nil {
		data["HasWeather"] = true
		data["Forecast"] = forecast // 模板里可以用 {{.Forecast.Condition}} 访问
	}

	// 执行模板渲染
	var buf bytes.Buffer
	err := m.systemTemplate.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("模板渲染执行失败: %w", err)
	}

	return buf.String(), nil
}
