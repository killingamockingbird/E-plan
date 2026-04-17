# E-plan: AI-Powered Fitness & Health Orchestrator

E-plan is a sophisticated Multi-Agent system built on the **CloudWeGo Eino** framework. It orchestrates specialized agents (Analyst & Coach) to generate personalized fitness plans by analyzing real-time weather data, historical health metrics, and user profiles.



---

## 🌟 Core Features

- **Multi-Agent Orchestration**: Seamless coordination between an **Analyst Agent** (ReAct-based) and a **Coach Agent**.
- **ReAct Reasoning Loop**: The Analyst autonomously decides when to call external tools (e.g., QWeather API) to gather necessary context.
- **Context-Aware Memory**: Persistent session management using **MongoDB** with a daily-rolling session strategy.
- **Robust Tool Integration**: Secure integration with external services using Ed25519 signature verification.
- **Production-Ready Config**: Flexible configuration management via Viper (supporting YAML and Env Vars).

---

## 🏗️ System Architecture

The project follows a Graph-based orchestration pattern:
1. **Load Context**: Fetches user profile and historical health data.
2. **Analyst Agent (Sub-Graph)**: Uses a Reasoning-Action loop to assess the environment (Weather, Health state).
3. **Coach Agent**: Synthesizes the analysis into a structured, actionable workout plan.
4. **Persistence**: Saves the reasoning chain and conversation history back to the database.



---

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- MongoDB 6.0+
- OpenAI or compatible LLM API Key (e.g., GPT-4o, DeepSeek)

### Installation
1. Clone the repository:
   ```bash
   git clone [https://github.com/your-username/E-plan.git](https://github.com/your-username/E-plan.git)
   cd E-plan
