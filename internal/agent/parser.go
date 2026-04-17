package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"E-plan/internal/domain"
)

// ParseLLMOutput extracts and parses a JSON object from the LLM output.
//
// We intentionally avoid time.Time fields in the expected JSON to make the LLM output stable.
// Server code fills date/time/user-owned fields after parsing.
func ParseLLMOutput(rawText string) (*domain.TrainingPlan, *domain.AnalysisReport, error) {
	cleanText := strings.TrimSpace(rawText)
	if strings.HasPrefix(cleanText, "```json") {
		cleanText = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(cleanText, "```json"), "```"))
	} else if strings.HasPrefix(cleanText, "```") {
		cleanText = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(cleanText, "```"), "```"))
	}

	type llmReport struct {
		BodyStatus string   `json:"body_status"`
		Summary    string   `json:"summary"`
		Highlights []string `json:"highlights"`
		Warnings   []string `json:"warnings"`
	}
	type llmPlan struct {
		ActivityType  domain.ActivityType `json:"activity_type"`
		Title         string              `json:"title"`
		Instructions  string              `json:"instructions"`
		TargetMetrics domain.PlanTargets  `json:"target_metrics"`
		Reasoning     string              `json:"reasoning"`
	}
	type llmResponse struct {
		Report llmReport `json:"report"`
		Plan   llmPlan   `json:"plan"`
	}

	var parsed llmResponse
	if err := json.Unmarshal([]byte(cleanText), &parsed); err != nil {
		return nil, nil, errors.New("LLM did not output valid JSON: " + err.Error())
	}

	report := &domain.AnalysisReport{
		BodyStatus: parsed.Report.BodyStatus,
		Summary:    parsed.Report.Summary,
		Highlights: parsed.Report.Highlights,
		Warnings:   parsed.Report.Warnings,
	}
	plan := &domain.TrainingPlan{
		ActivityType:  parsed.Plan.ActivityType,
		Title:         parsed.Plan.Title,
		Instructions:  parsed.Plan.Instructions,
		TargetMetrics: parsed.Plan.TargetMetrics,
		Reasoning:     parsed.Plan.Reasoning,
	}

	return plan, report, nil
}

// ParseAnalysisReportJSON 解析分析师 Agent 输出的 JSON
func ParseAnalysisReportJSON(rawContent string) (*domain.AnalysisReport, error) {
	cleanJSON := cleanLLMJSON(rawContent)

	var report domain.AnalysisReport
	if err := json.Unmarshal([]byte(cleanJSON), &report); err != nil {
		return nil, fmt.Errorf("解析 AnalysisReport JSON 失败: %w, 原始文本: %s", err, rawContent)
	}

	return &report, nil
}

// ParseTrainingPlanJSON 解析教练 Agent 输出的 JSON (顺便帮你把这个也补齐)
func ParseTrainingPlanJSON(rawContent string) (*domain.TrainingPlan, error) {
	cleanJSON := cleanLLMJSON(rawContent)

	var plan domain.TrainingPlan
	if err := json.Unmarshal([]byte(cleanJSON), &plan); err != nil {
		return nil, fmt.Errorf("解析 TrainingPlan JSON 失败: %w, 原始文本: %s", err, rawContent)
	}

	return &plan, nil
}

// cleanLLMJSON 清洗大模型输出的 Markdown 格式符号
// 很多时候 LLM 会输出: ```json\n { ... } \n```，这会导致 Unmarshal 报错
func cleanLLMJSON(raw string) string {
	cleaned := strings.TrimSpace(raw)

	// 去除前缀
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}

	// 去除后缀
	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}

	return strings.TrimSpace(cleaned)
}
