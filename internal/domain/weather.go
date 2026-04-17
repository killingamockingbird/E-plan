package domain

// WeatherCondition 天气状况实体
type WeatherCondition struct {
	City        string `json:"city"`
	Condition   string `json:"condition"`   // 例如："晴朗", "小雨", "多云"
	Temperature string `json:"temperature"` // 例如："15°C - 25°C"
}
