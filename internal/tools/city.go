package tools

import (
	"E-plan/internal/tools/common"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// CityResp 城市查询接口响应体
type CityResp struct {
	Code     string     `json:"code"`
	Location []Location `json:"location"`
}

// Location 具体的城市信息实体
type Location struct {
	ID      string `json:"id"`      // 城市ID
	Name    string `json:"name"`    // 城市名称
	Adm1    string `json:"adm1"`    // 上级行政区
	Adm2    string `json:"adm2"`    // 下级行政区
	Country string `json:"country"` // 国家
	Lat     string `json:"lat"`     // 纬度
	Lon     string `json:"lon"`     // 经度
	Rank    string `json:"rank"`    // 排序权重
}

// GetCityInformation 获取城市信息 (已改造为支持 Context 和错误返回)
// 建议将其挂载到之前定义的 WeatherClient 上，复用 privateKeyPath
func (c *WeatherClient) GetCityInformation(ctx context.Context, location string) ([]Location, error) {
	loc := url.QueryEscape(location)
	GeoUrl := fmt.Sprintf("https://jk436h87cj.re.qweatherapi.com/geo/v2/city/lookup?location=%s", loc)

	// 1. 从 Client 配置中读取私钥路径，告别硬编码
	privateKey, err := common.LoadEd25519PrivateKey(c.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载私钥失败: %w", err)
	}

	signedToken := common.GetToken(privateKey)

	// 2. 携带 Context 发起 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", GeoUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("构建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+signedToken)
	req.Header.Set("Content-Type", "application/json")

	// 3. 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求城市 API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("城市 API 返回非 200 状态码: %d", resp.StatusCode)
	}

	// 4. 解析响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var cityResp CityResp
	if err := json.Unmarshal(body, &cityResp); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w, 原始数据: %s", err, string(body))
	}

	if cityResp.Code != "200" {
		return nil, fmt.Errorf("获取城市信息业务报错, 错误码: %s", cityResp.Code)
	}

	// 5. 判空处理
	if len(cityResp.Location) == 0 {
		return nil, fmt.Errorf("未查询到该城市的信息: %s", location)
	}

	return cityResp.Location, nil
}
