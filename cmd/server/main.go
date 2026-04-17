package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"E-plan/configs" // 引入你的配置模块
	"E-plan/internal/agent"
	"E-plan/internal/domain"
	"E-plan/internal/prompt"
	"E-plan/internal/repository/mongodb"
	"E-plan/internal/tools"
)

func main() {
	log.Println("🚀 [E-plan] Initializing system components...")

	// ---------------------------------------------------------
	// PHASE 0: Configuration Loading
	// Load application settings from YAML or Environment Variables.
	// ---------------------------------------------------------
	cfg, err := configs.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("CRITICAL: Failed to load configuration: %v", err)
	}
	log.Printf("✅ Configuration: Loaded successfully (Mode: %s)", cfg.Server.Mode)

	// ---------------------------------------------------------
	// PHASE 1: Persistent Storage Initialization
	// Establishing a high-availability connection to MongoDB.
	// ---------------------------------------------------------
	dbTimeout := time.Duration(cfg.MongoDB.TimeoutSeconds) * time.Second
	dbCtx, dbCancel := context.WithTimeout(context.Background(), dbTimeout)
	defer dbCancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoDB.URI)
	mongoClient, err := mongo.Connect(dbCtx, clientOptions)
	if err != nil {
		log.Fatalf("CRITICAL: MongoDB connection failed: %v", err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("ERROR: Graceful MongoDB disconnection failed: %v", err)
		}
	}()

	if err := mongoClient.Ping(dbCtx, nil); err != nil {
		log.Fatalf("CRITICAL: MongoDB liveness probe failed: %v", err)
	}
	log.Println("✅ Data layer: MongoDB connection established")
	db := mongoClient.Database(cfg.MongoDB.Database)

	// ---------------------------------------------------------
	// PHASE 2: Infrastructure & Dependency Injection
	// ---------------------------------------------------------
	userRepo := mongodb.NewUserRepo(db)
	sportsDataRepo := mongodb.NewSportsDataRepo(db)
	planRepo := mongodb.NewPlanRepo(db)
	historyRepo := mongodb.NewHistoryRepo(db)

	// Use configured API Key for weather services
	weatherClient := tools.NewWeatherClient(cfg.WeatherAPI.APIKey)

	promptMgr, err := prompt.NewManager("configs/system_prompt.txt")
	if err != nil {
		log.Fatalf("CRITICAL: Prompt template engine failed to initialize: %v", err)
	}

	// ---------------------------------------------------------
	// PHASE 3: Large Language Model (LLM) Configuration
	// ---------------------------------------------------------
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL:     cfg.LLM.BaseURL,
		APIKey:      cfg.LLM.APIKey,
		Model:       cfg.LLM.Model,
		Temperature: &cfg.LLM.Temperature,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		log.Printf("WARNING: LLM initialization failed: %v\n", err)
		chatModel = nil
	}

	coachClient := agent.NewLLMClient(chatModel)
	analystAgent, err := agent.BuildAnalystAgent(context.Background(), chatModel, weatherClient)
	if err != nil {
		log.Fatalf("CRITICAL: Analyst Agent failed to build: %v", err)
	}

	// ---------------------------------------------------------
	// PHASE 4: Agent Orchestration (DAG Construction)
	// ---------------------------------------------------------
	agentNodes := &agent.AgentNodes{
		UserRepo:     userRepo,
		DataRepo:     sportsDataRepo,
		PromptMgr:    promptMgr,
		PlanRepo:     planRepo,
		HistoryRepo:  historyRepo,
		WeatherAPI:   weatherClient,
		AnalystAgent: analystAgent,
		CoachClient:  coachClient,
	}

	orchestrator := agent.NewOrchestrator(agentNodes)
	runnableGraph, err := orchestrator.Build(context.Background())
	if err != nil {
		log.Fatalf("CRITICAL: Workflow graph compilation error: %v", err)
	}
	log.Println("✅ Orchestration layer: Eino graph ready")

	// ---------------------------------------------------------
	// PHASE 5: Presentation Layer (HTTP/REST API)
	// ---------------------------------------------------------
	http.HandleFunc("/api/v1/agent/daily", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, `{"error": "user_id is required"}`, http.StatusBadRequest)
			return
		}

		sessionID := fmt.Sprintf("%s_%s", userID, time.Now().Format("20060102"))

		// In a real scenario, this data would come from the request body
		mockTodayData := &domain.DailySummary{
			UserID:   userID,
			Date:     time.Now(),
			TotalCal: 500,
		}

		initialState := &agent.AgentState{
			UserID:    userID,
			SessionID: sessionID,
			TodayData: mockTodayData,
		}

		finalState, err := runnableGraph.Invoke(r.Context(), initialState)
		if err != nil {
			log.Printf("ERROR: Graph execution error: %v", err)
			http.Error(w, `{"error": "Agent reasoning failed"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"plan":   finalState.FinalPlan,
				"report": finalState.FinalReport,
			},
		}

		json.NewEncoder(w).Encode(response)
	})

	serverAddr := ":" + cfg.Server.Port
	log.Printf("🌐 Gateway: Listening on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("FATAL: Server crash: %v", err)
	}
}
