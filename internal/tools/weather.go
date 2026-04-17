package tools

import (
	"E-plan/internal/tools/common"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"E-plan/internal/domain"
)

// WeatherResp 和风天气 API 响应结构
type WeatherResp struct {
	Code       string     `json:"code"`
	UpdateTime string     `json:"updateTime"`
	FxLink     string     `json:"fxLink"`
	Now        NowWeather `json:"now"`
}

type NowWeather struct {
	Temp      string `json:"temp"`      //温度
	FeelsLike string `json:"feelsLike"` //体感温度
	Icon      string `json:"icon"`      //天气图像
	Text      string `json:"text"`      //文字描述 如 晴
	Wind      string `json:"wind360"`   //风向360
	WindDir   string `json:"windDir"`   //风向
	WindSpeed string `json:"windSpeed"` //风速
	WindScale string `json:"windScale"` //风力等级
	Humidity  string `json:"humidity"`  //湿度
	Pressure  string `json:"pressure"`  //大气压（百帕）
	Vis       string `json:"vis"`       //可见度
	Cloud     string `json:"cloud"`     //云量
}

// WeatherClient 天气服务客户端
type WeatherClient struct {
	privateKeyPath string
}

// NewWeatherClient 初始化客户端，传入私钥路径
func NewWeatherClient(pkPath string) *WeatherClient {
	return &WeatherClient{
		privateKeyPath: pkPath,
	}
}

// GetTomorrowWeather 适配 Agent 编排器需要的方法签名
// 注意：为了贴合你原本的代码，这里底层依然请求的是 now API。
// 如果你想获取真实的明天天气，只需将下面的 GeoUrl 改为 https://api.qweather.com/v7/weather/3d?location=%s 并解析 daily 数组的第二天即可。
func (c *WeatherClient) GetTomorrowWeather(ctx context.Context, location string) (*domain.WeatherCondition, error) {

	// 1. 获取城市信息 (复用你原有的方法)
	locations, err := c.GetCityInformation(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("查询城市基础信息失败: %w", err)
	}

	// 默认取搜索出来的第一个城市匹配项
	targetCity := locations[0]
	GeoUrl := fmt.Sprintf("https://jk436h87cj.re.qweatherapi.com/v7/weather/now?location=%s", targetCity.ID)

	// 2. 加载私钥并生成 Token
	privateKey, err := common.LoadEd25519PrivateKey(c.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载私钥失败: %w", err)
	}
	signedToken := common.GetToken(privateKey)

	// 3. 构建携带 Context 的 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", GeoUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("构建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+signedToken)
	req.Header.Set("Content-Type", "application/json")

	// 4. 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求天气 API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("天气 API 返回非 200 状态码: %d", resp.StatusCode)
	}

	// 5. 解析响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var weatherResp WeatherResp
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w, 原始数据: %s", err, string(body))
	}

	if weatherResp.Code != "200" {
		return nil, fmt.Errorf("天气业务报错, 错误码: %s", weatherResp.Code)
	}

	// 6. 将 API 的结果转换为 Agent 总线 (domain.WeatherCondition) 能理解的统一结构
	w := weatherResp.Now
	result := &domain.WeatherCondition{
		City:      targetCity.Name,
		Condition: w.Text, // e.g., "晴"
		// 结合实际温度和体感温度给到大模型，大模型更容易判断运动强度
		Temperature: fmt.Sprintf("%s°C (体感 %s°C)", w.Temp, w.FeelsLike),
	}

	return result, nil
}
