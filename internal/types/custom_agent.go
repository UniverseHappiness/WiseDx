package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// BuiltinAgentID constants for built-in agents
const (
	// BuiltinQuickAnswerID is the ID for the built-in quick answer (RAG) agent
	BuiltinQuickAnswerID = "builtin-quick-answer"
	// BuiltinSmartReasoningID is the ID for the built-in smart reasoning (ReAct) agent
	BuiltinSmartReasoningID = "builtin-smart-reasoning"
	// BuiltinDeepResearcherID is the ID for the built-in deep researcher agent
	BuiltinDeepResearcherID = "builtin-deep-researcher"
	// BuiltinDataAnalystID is the ID for the built-in data analyst agent
	BuiltinDataAnalystID = "builtin-data-analyst"
	// BuiltinKnowledgeGraphExpertID is the ID for the built-in knowledge graph expert agent
	BuiltinKnowledgeGraphExpertID = "builtin-knowledge-graph-expert"
	// BuiltinDocumentAssistantID is the ID for the built-in document assistant agent
	BuiltinDocumentAssistantID = "builtin-document-assistant"
	// BuiltinMedicalConsultantID is the ID for the built-in medical consultation agent
	BuiltinMedicalConsultantID = "builtin-medical-consultant"
)

// AgentMode constants for agent running mode
const (
	// AgentModeQuickAnswer is the RAG mode for quick Q&A
	AgentModeQuickAnswer = "quick-answer"
	// AgentModeSmartReasoning is the ReAct mode for multi-step reasoning
	AgentModeSmartReasoning = "smart-reasoning"
)

// CustomAgent represents a configurable AI agent (similar to GPTs)
type CustomAgent struct {
	// Unique identifier of the agent (composite primary key with TenantID)
	// For built-in agents, this is 'builtin-quick-answer' or 'builtin-smart-reasoning'
	// For custom agents, this is a UUID
	ID string `yaml:"id" json:"id" gorm:"type:varchar(36);primaryKey"`
	// Name of the agent
	Name string `yaml:"name" json:"name" gorm:"type:varchar(255);not null"`
	// Description of the agent
	Description string `yaml:"description" json:"description" gorm:"type:text"`
	// Avatar/Icon of the agent (emoji or icon name)
	Avatar string `yaml:"avatar" json:"avatar" gorm:"type:varchar(64)"`
	// Whether this is a built-in agent (normal mode / agent mode)
	IsBuiltin bool `yaml:"is_builtin" json:"is_builtin" gorm:"default:false"`
	// Tenant ID (composite primary key with ID)
	TenantID uint64 `yaml:"tenant_id" json:"tenant_id" gorm:"primaryKey"`
	// Created by user ID
	CreatedBy string `yaml:"created_by" json:"created_by" gorm:"type:varchar(36)"`

	// Agent configuration
	Config CustomAgentConfig `yaml:"config" json:"config" gorm:"type:json"`

	// Timestamps
	CreatedAt time.Time      `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time      `yaml:"updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `yaml:"deleted_at" json:"deleted_at" gorm:"index"`
}

// CustomAgentConfig represents the configuration of a custom agent
type CustomAgentConfig struct {
	// ===== Basic Settings =====
	// Agent mode: "quick-answer" for RAG mode, "smart-reasoning" for ReAct agent mode
	AgentMode string `yaml:"agent_mode" json:"agent_mode"`
	// System prompt for the agent (unified prompt, uses {{web_search_status}} placeholder for dynamic behavior)
	SystemPrompt string `yaml:"system_prompt" json:"system_prompt"`
	// Context template for normal mode (how to format retrieved chunks)
	ContextTemplate string `yaml:"context_template" json:"context_template"`

	// ===== Model Settings =====
	// Model ID to use for conversations
	ModelID string `yaml:"model_id" json:"model_id"`
	// ReRank model ID for retrieval
	RerankModelID string `yaml:"rerank_model_id" json:"rerank_model_id"`
	// Temperature for LLM (0-1)
	Temperature float64 `yaml:"temperature" json:"temperature"`
	// Maximum completion tokens (only for normal mode)
	MaxCompletionTokens int `yaml:"max_completion_tokens" json:"max_completion_tokens"`
	// Whether to enable thinking mode (for models that support extended thinking)
	Thinking *bool `yaml:"thinking" json:"thinking"`

	// ===== Agent Mode Settings =====
	// Maximum iterations for ReAct loop (only for agent type)
	MaxIterations int `yaml:"max_iterations" json:"max_iterations"`
	// Allowed tools (only for agent type)
	AllowedTools []string `yaml:"allowed_tools" json:"allowed_tools"`
	// Whether reflection is enabled (only for agent type)
	ReflectionEnabled bool `yaml:"reflection_enabled" json:"reflection_enabled"`
	// MCP service selection mode: "all" = all enabled MCP services, "selected" = specific services, "none" = no MCP
	MCPSelectionMode string `yaml:"mcp_selection_mode" json:"mcp_selection_mode"`
	// Selected MCP service IDs (only used when MCPSelectionMode is "selected")
	MCPServices []string `yaml:"mcp_services" json:"mcp_services"`

	// ===== Knowledge Base Settings =====
	// Knowledge base selection mode: "all" = all KBs, "selected" = specific KBs, "none" = no KB
	KBSelectionMode string `yaml:"kb_selection_mode" json:"kb_selection_mode"`
	// Associated knowledge base IDs (only used when KBSelectionMode is "selected")
	KnowledgeBases []string `yaml:"knowledge_bases" json:"knowledge_bases"`
	// Whether to retrieve knowledge base only when explicitly mentioned with @ (default: false)
	// When true, knowledge base retrieval only happens if user explicitly mentions KB/files with @
	// When false, knowledge base retrieval happens according to KBSelectionMode
	RetrieveKBOnlyWhenMentioned bool `yaml:"retrieve_kb_only_when_mentioned" json:"retrieve_kb_only_when_mentioned"`

	// ===== File Type Restriction Settings =====
	// Supported file types for this agent (e.g., ["csv", "xlsx", "xls"])
	// Empty means all file types are supported
	// When set, only files with matching extensions can be used with this agent
	SupportedFileTypes []string `yaml:"supported_file_types" json:"supported_file_types"`

	// ===== FAQ Strategy Settings =====
	// Whether FAQ priority strategy is enabled (FAQ answers prioritized over document chunks)
	FAQPriorityEnabled bool `yaml:"faq_priority_enabled" json:"faq_priority_enabled"`
	// FAQ direct answer threshold - if similarity > this value, use FAQ answer directly
	FAQDirectAnswerThreshold float64 `yaml:"faq_direct_answer_threshold" json:"faq_direct_answer_threshold"`
	// FAQ score boost multiplier - FAQ results score multiplied by this factor
	FAQScoreBoost float64 `yaml:"faq_score_boost" json:"faq_score_boost"`

	// ===== Web Search Settings =====
	// Whether web search is enabled
	WebSearchEnabled bool `yaml:"web_search_enabled" json:"web_search_enabled"`
	// Maximum web search results
	WebSearchMaxResults int `yaml:"web_search_max_results" json:"web_search_max_results"`

	// ===== Multi-turn Conversation Settings =====
	// Whether multi-turn conversation is enabled
	MultiTurnEnabled bool `yaml:"multi_turn_enabled" json:"multi_turn_enabled"`
	// Number of history turns to keep in context
	HistoryTurns int `yaml:"history_turns" json:"history_turns"`

	// ===== Retrieval Strategy Settings (for both modes) =====
	// Embedding/Vector retrieval top K
	EmbeddingTopK int `yaml:"embedding_top_k" json:"embedding_top_k"`
	// Keyword retrieval threshold
	KeywordThreshold float64 `yaml:"keyword_threshold" json:"keyword_threshold"`
	// Vector retrieval threshold
	VectorThreshold float64 `yaml:"vector_threshold" json:"vector_threshold"`
	// Rerank top K
	RerankTopK int `yaml:"rerank_top_k" json:"rerank_top_k"`
	// Rerank threshold
	RerankThreshold float64 `yaml:"rerank_threshold" json:"rerank_threshold"`

	// ===== Advanced Settings (mainly for normal mode) =====
	// Whether to enable query expansion
	EnableQueryExpansion bool `yaml:"enable_query_expansion" json:"enable_query_expansion"`
	// Whether to enable query rewrite for multi-turn conversations
	EnableRewrite bool `yaml:"enable_rewrite" json:"enable_rewrite"`
	// Rewrite prompt system message
	RewritePromptSystem string `yaml:"rewrite_prompt_system" json:"rewrite_prompt_system"`
	// Rewrite prompt user message template
	RewritePromptUser string `yaml:"rewrite_prompt_user" json:"rewrite_prompt_user"`
	// Fallback strategy: "fixed" for fixed response, "model" for model generation
	FallbackStrategy string `yaml:"fallback_strategy" json:"fallback_strategy"`
	// Fixed fallback response (when FallbackStrategy is "fixed")
	FallbackResponse string `yaml:"fallback_response" json:"fallback_response"`
	// Fallback prompt (when FallbackStrategy is "model")
	FallbackPrompt string `yaml:"fallback_prompt" json:"fallback_prompt"`
}

// Value implements driver.Valuer interface for CustomAgentConfig
func (c CustomAgentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner interface for CustomAgentConfig
func (c *CustomAgentConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

// TableName returns the table name for CustomAgent
func (CustomAgent) TableName() string {
	return "custom_agents"
}

// EnsureDefaults sets default values for the agent
func (a *CustomAgent) EnsureDefaults() {
	if a == nil {
		return
	}
	if a.Config.Temperature == 0 {
		a.Config.Temperature = 0.7
	}
	if a.Config.MaxIterations == 0 {
		a.Config.MaxIterations = 10
	}
	if a.Config.WebSearchMaxResults == 0 {
		a.Config.WebSearchMaxResults = 5
	}
	if a.Config.HistoryTurns == 0 {
		a.Config.HistoryTurns = 5
	}
	// Retrieval strategy defaults
	if a.Config.EmbeddingTopK == 0 {
		a.Config.EmbeddingTopK = 10
	}
	if a.Config.KeywordThreshold == 0 {
		a.Config.KeywordThreshold = 0.3
	}
	if a.Config.VectorThreshold == 0 {
		a.Config.VectorThreshold = 0.5
	}
	if a.Config.RerankTopK == 0 {
		a.Config.RerankTopK = 5
	}
	if a.Config.RerankThreshold == 0 {
		a.Config.RerankThreshold = 0.5
	}
	// Advanced settings defaults
	if a.Config.FallbackStrategy == "" {
		a.Config.FallbackStrategy = "model"
	}
	if a.Config.MaxCompletionTokens == 0 {
		a.Config.MaxCompletionTokens = 2048
	}
	// Agent mode should always enable multi-turn conversation
	if a.Config.AgentMode == AgentModeSmartReasoning {
		a.Config.MultiTurnEnabled = true
	}
}

// IsAgentMode returns true if this agent uses ReAct agent mode
func (a *CustomAgent) IsAgentMode() bool {
	return a.Config.AgentMode == AgentModeSmartReasoning
}

// GetBuiltinQuickAnswerAgent returns the built-in quick answer (RAG) mode agent
func GetBuiltinQuickAnswerAgent(tenantID uint64) *CustomAgent {
	return &CustomAgent{
		ID:          BuiltinQuickAnswerID,
		Name:        "快速问答",
		Description: "基于知识库的 RAG 问答，快速准确地回答问题",
		IsBuiltin:   true,
		TenantID:    tenantID,
		Config: CustomAgentConfig{
			AgentMode:    AgentModeQuickAnswer,
			SystemPrompt: "",
			ContextTemplate: `请根据以下参考资料回答用户问题。

参考资料：
{{contexts}}

用户问题：{{query}}`,
			Temperature:                 0.7,
			MaxCompletionTokens:         2048,
			WebSearchEnabled:            true,
			WebSearchMaxResults:         5,
			MultiTurnEnabled:            true,
			HistoryTurns:                5,
			KBSelectionMode:             "all",
			RetrieveKBOnlyWhenMentioned: false, // Default: retrieve KB based on KBSelectionMode
			// FAQ strategy
			FAQPriorityEnabled:       true,
			FAQDirectAnswerThreshold: 0.9,
			FAQScoreBoost:            1.2,
			// Retrieval strategy
			EmbeddingTopK:    10,
			KeywordThreshold: 0.3,
			VectorThreshold:  0.5,
			RerankTopK:       10,
			RerankThreshold:  0.3,
			// Advanced settings
			EnableQueryExpansion: true,
			EnableRewrite:        true,
			FallbackStrategy:     "model",
		},
	}
}

// GetBuiltinSmartReasoningAgent returns the built-in smart reasoning (ReAct) mode agent
func GetBuiltinSmartReasoningAgent(tenantID uint64) *CustomAgent {
	return &CustomAgent{
		ID:          BuiltinSmartReasoningID,
		Name:        "智能推理",
		Description: "ReAct 推理框架，支持多步思考和工具调用",
		IsBuiltin:   true,
		TenantID:    tenantID,
		Config: CustomAgentConfig{
			AgentMode:                   AgentModeSmartReasoning,
			SystemPrompt:                "",
			Temperature:                 0.7,
			MaxCompletionTokens:         2048,
			MaxIterations:               50,
			KBSelectionMode:             "all",
			RetrieveKBOnlyWhenMentioned: false, // Default: retrieve KB based on KBSelectionMode
			AllowedTools:                []string{"thinking", "todo_write", "knowledge_search", "grep_chunks", "list_knowledge_chunks", "query_knowledge_graph", "get_document_info"},
			WebSearchEnabled:            true,
			WebSearchMaxResults:         5,
			ReflectionEnabled:           false,
			MultiTurnEnabled:            true,
			HistoryTurns:                5,
			// FAQ strategy
			FAQPriorityEnabled:       true,
			FAQDirectAnswerThreshold: 0.9,
			FAQScoreBoost:            1.2,
			// Retrieval strategy
			EmbeddingTopK:    10,
			KeywordThreshold: 0.3,
			VectorThreshold:  0.5,
			RerankTopK:       10,
			RerankThreshold:  0.3,
		},
	}
}

// GetBuiltinDataAnalystAgent returns the built-in data analyst agent
// This agent specializes in analyzing CSV/Excel data using SQL queries via DuckDB
func GetBuiltinDataAnalystAgent(tenantID uint64) *CustomAgent {
	return &CustomAgent{
		ID:          BuiltinDataAnalystID,
		Name:        "数据分析师",
		Description: "专业数据分析智能体，支持 CSV/Excel 文件的 SQL 查询与统计分析",
		Avatar:      "📊",
		IsBuiltin:   true,
		TenantID:    tenantID,
		Config: CustomAgentConfig{
			AgentMode: AgentModeSmartReasoning,
			SystemPrompt: `### Role
You are WeKnora Data Analyst, an intelligent data analysis assistant powered by DuckDB. You specialize in analyzing structured data from CSV and Excel files using SQL queries.

### Mission
Help users explore, analyze, and derive insights from their tabular data through intelligent SQL query generation and execution.

### Critical Constraints
1. **Schema First:** ALWAYS call data_schema before writing any SQL query to understand the table structure.
2. **Read-Only:** Only SELECT queries allowed. INSERT, UPDATE, DELETE, CREATE, DROP are forbidden.
3. **Iterative Refinement:** If a query fails, analyze the error and refine your approach.

### Workflow
1. **Understand:** Call data_schema to get table name, columns, types, and row count.
2. **Plan:** For complex questions, use todo_write to break into sub-queries.
3. **Query:** Call data_analysis with the knowledge_id and SQL query.
4. **Analyze:** Interpret results and provide insights.

### SQL Best Practices for DuckDB
- Use double quotes for identifiers: SELECT "Column Name" FROM "table_name"
- Aggregate functions: COUNT(*), SUM(), AVG(), MIN(), MAX(), MEDIAN(), STDDEV()
- String matching: LIKE, ILIKE (case-insensitive), REGEXP
- Use LIMIT to prevent overwhelming output (default to 100 rows max)

### Tool Guidelines
- **data_schema:** ALWAYS use first. Required before any query.
- **data_analysis:** Execute SQL queries. Only SELECT queries allowed.
- **thinking:** Plan complex analyses, debug query issues.
- **todo_write:** Track multi-step analysis tasks.

### Output Standards
- Present results in well-formatted tables or summaries
- Provide actionable insights, not just raw numbers
- Relate findings back to the user's original question

Current Time: {{current_time}}
`,
			Temperature:                 0.3, // Lower temperature for precise SQL generation
			MaxCompletionTokens:         4096,
			MaxIterations:               30,
			KBSelectionMode:             "all",
			RetrieveKBOnlyWhenMentioned: false, // Default: retrieve KB based on KBSelectionMode
			// Only support CSV and Excel files for data analysis
			// Use standard values (xlsx), backend will auto-include xls via alias
			SupportedFileTypes: []string{"csv", "xlsx"},
			// Core tools for data analysis
			AllowedTools: []string{
				"thinking",
				"todo_write",
				"data_schema",   // Get table schema information
				"data_analysis", // Execute SQL queries on data
			},
			WebSearchEnabled:    false, // Data analysis doesn't need web search
			WebSearchMaxResults: 0,
			ReflectionEnabled:   true, // Enable reflection for query optimization
			MultiTurnEnabled:    true,
			HistoryTurns:        10, // More history for iterative analysis
			// Retrieval strategy (minimal, as we focus on data tools)
			EmbeddingTopK:    5,
			KeywordThreshold: 0.3,
			VectorThreshold:  0.5,
			RerankTopK:       5,
			RerankThreshold:  0.3,
		},
	}
}

// Deprecated: Use GetBuiltinQuickAnswerAgent instead
func GetBuiltinNormalAgent(tenantID uint64) *CustomAgent {
	return GetBuiltinQuickAnswerAgent(tenantID)
}

// Deprecated: Use GetBuiltinSmartReasoningAgent instead
func GetBuiltinAgentAgent(tenantID uint64) *CustomAgent {
	return GetBuiltinSmartReasoningAgent(tenantID)
}

// GetBuiltinMedicalConsultantAgent returns the built-in medical consultation agent
// This agent specializes in structured medical history taking following clinical workflow
func GetBuiltinMedicalConsultantAgent(tenantID uint64) *CustomAgent {
	return &CustomAgent{
		ID:          BuiltinMedicalConsultantID,
		Name:        "智能问诊助手",
		Description: "专业医疗问诊智能体，支持从问候到总结的全流程结构化病史采集",
		Avatar:      "🏥",
		IsBuiltin:   true,
		TenantID:    tenantID,
		Config: CustomAgentConfig{
			AgentMode: AgentModeSmartReasoning,
			SystemPrompt: `### 角色
你是“慧诊”智能问诊助手，一个专业的医疗问诊AI，专注于通过结构化对话收集患者病史信息。

### 使命
通过友好、专业的多轮对话，系统性地采集患者完整病史，为医生诊断提供全面、准确的信息支持。


## 选择式问答规范

对于特定问题，**必须使用 show_options 工具**展示选项按钮，方便患者快速点击选择：

### 使用 show_options 工具的场景：

**性别选择：**
` + "`" + `json
{"question": "请选择您的性别", "options": [{"label": "A. 男", "value": "男"}, {"label": "B. 女", "value": "女"}]}
` + "`" + `

**症状严重程度（1-10分）：**
` + "`" + `json
{"question": "请评估症状严重程度", "options": [
  {"label": "轻微 (1-3分)", "value": "轻微"},
  {"label": "中等 (4-6分)", "value": "中等"},
  {"label": "严重 (7-9分)", "value": "严重"},
  {"label": "极重 (10分)", "value": "极重"}
]}
` + "`" + `

**症状持续时间：**
` + "`" + `json
{"question": "症状持续多长时间？", "options": [
  {"label": "今天刚开始", "value": "今天刚开始"},
  {"label": "1-3天", "value": "1-3天"},
  {"label": "4-7天", "value": "4-7天"},
  {"label": "1-2周", "value": "1-2周"},
  {"label": "2周以上", "value": "2周以上"}
]}
` + "`" + `

**既往疾病（多选）：**
` + "`" + `json
{"question": "您是否有以下疾病？(可多选)", "options": [
  {"label": "高血压", "value": "高血压"},
  {"label": "糖尿病", "value": "糖尿病"},
  {"label": "心脏病", "value": "心脏病"},
  {"label": "无上述疾病", "value": "无"}
], "multi_select": true}
` + "`" + `

**过敏情况：**
` + "`" + `json
{"question": "您是否有以下过敏？(可多选)", "options": [
  {"label": "无已知过敏", "value": "无"},
  {"label": "青霉素类", "value": "青霉素类"},
  {"label": "头孢类", "value": "头孢类"},
  {"label": "海鲜", "value": "海鲜"},
  {"label": "其他", "value": "其他"}
], "multi_select": true}
` + "`" + `

---
## 📋 问诊流程（6个阶段）

### 阶段 1: 问候与身份确认
- 友好问候患者，显示进度条
- 确认患者基本信息（姓名、年龄、性别-提供选项）
- 说明问诊流程和预计时间
**提取信息**: 姓名、年龄、性别

### 阶段 2: 主诉采集
- 显示进度条
- 询问主要不适症状
- 提供**持续时间选项**
- 提供**严重程度选项**
**提取信息**: 主要症状、持续时间、严重程度

### 阶段 3: 现病史采集
- 显示进度条
- 详细询问症状发生发展过程
- 了解症状的部位、性质、程度、诱因、缓解因素
- 询问伴随症状
- 了解已采取的治疗措施及效果
**提取信息**: 发病时间、诱因、症状特点、伴随症状、已有治疗

### 阶段 4: 既往史采集
- 显示进度条
- 提供**既往疾病选项**（可多选）
- 了解手术史、住院史
- 了解长期用药情况
**提取信息**: 既往疾病、手术史、住院史、长期用药

### 阶段 5: 过敏史采集
- 显示进度条
- 提供**过敏情况选项**（可多选）
- 详细了解过敏反应
**提取信息**: 药物过敏、食物过敏、其他过敏

### 阶段 6: 信息总结与报告导出
- 显示完成进度条
- 确认所有收集的信息
- 生成可导出的病史摘要报告

---
## 工作流程

### 【最重要】首次响应必须调用 todo_write

**在你的第一条响应中，必须立即调用 todo_write 创建问诊任务计划：**

` + "`" + `json
{
  "task": "医疗问诊病史采集",
  "steps": [
    {"id": "step1", "description": "问候与身份确认", "status": "in_progress"},
    {"id": "step2", "description": "主诉采集", "status": "pending"},
    {"id": "step3", "description": "现病史采集", "status": "pending"},
    {"id": "step4", "description": "既往史采集", "status": "pending"},
    {"id": "step5", "description": "过敏史采集", "status": "pending"},
    {"id": "step6", "description": "信息总结与报告导出", "status": "pending"}
  ],
  "collected_data": {}
}
` + "`" + `

### 【最重要】阶段信息完整后才更新 todo_write

**不要每次收到用户回复就立即调用 todo_write！**

正确流程：
1. 收到用户回复后，先分析当前阶段的信息是否已完整
2. 如果当前阶段还缺少信息，继续提问收集
3. 只有当一个阶段的所有必要信息都已收集完毕时，才调用 todo_write 更新状态并进入下一阶段

**阶段完整性判断标准：**

| 阶段 | 必要信息 | 完整标准 |
|------|---------|----------|
| 阶段 1 | 姓名、年龄、性别 | 3项全部收集完成 |
| 阶段 2 | 主要症状、持续时间、严重程度 | 3项全部收集完成 |
| 阶段 3 | 发病时间、诱因、症状特点 | 至少收集到关键信息 |
| 阶段 4 | 既往疾病 | 收集到选择结果 |
| 阶段 5 | 过敏情况 | 收集到选择结果 |

**示例：阶段1信息收集过程**

用户回复"张三"（只有姓名）→ 不调用todo_write，继续询问年龄
用户回复"35岁"（姓名+年龄）→ 不调用todo_write，继续询问性别
用户选择"男"（姓名+年龄+性别）→ 阶段1完整！现在调用todo_write更新

` + "`" + `json
{
  "task": "医疗问诊病史采集",
  "steps": [
    {"id": "step1", "description": "问候与身份确认", "status": "completed"},
    {"id": "step2", "description": "主诉采集", "status": "in_progress"},
    {"id": "step3", "description": "现病史采集", "status": "pending"},
    {"id": "step4", "description": "既往史采集", "status": "pending"},
    {"id": "step5", "description": "过敏史采集", "status": "pending"},
    {"id": "step6", "description": "信息总结与报告导出", "status": "pending"}
  ],
  "collected_data": {
    "姓名": "张三",
    "年龄": "35岁",
    "性别": "男"
  }
}
` + "`" + `

### collected_data 字段说明

这个字段用于存储所有收集到的患者信息，便于在侧栏展示。常用字段：
- 姓名、年龄、性别
- 主诉、持续时间、严重程度
- 发病时间、诱因、症状特点、伴随症状、已有治疗
- 既往疾病、手术史、住院史、长期用药
- 药物过敏、食物过敏、其他过敏

---
## ✅ 关键规则

1. **【最重要】首次必调 todo_write**: 第一条响应必须调用 todo_write 创建任务
2. **【最重要】阶段完整才更新**: 只有当当前阶段的所有必要信息都已收集完毕时，才调用 todo_write
3. **【最重要】每次必调 thinking**: 每次回复**必须先调用 thinking 工具**回顾已收集信息，防止遗忘。
4. **禁止即时更新**: 不要每次收到用户回复就立即调用 todo_write，先判断信息是否完整
5. **选项优先**: 适用的问题优先提供选项
6. **循序渐进**: 严格按照阶段顺序进行
7. **信息确认**: 每阶段结束前确认信息准确性

---
## 报告导出格式（最终总结）

完成所有阶段后，生成Markdown格式的报告。

---
## 工具使用指南

**每次回复的调用顺序：**
1. **thinking**（必调，但仅调用一次）：先回顾已收集信息，分析当前状态
2. **show_options**（可选）：如需展示选项
3. **todo_write**（仅阶段完成时）：更新进度

⚠️ **重要规则**：
- 每轮对话只能调用 thinking 工具**一次**
- thinking 输出后必须立即决定下一步动作（show_options/todo_write/直接回答）
- 禁止在一轮中多次调用 thinking 形成循环

---
## ⚠️ 注意事项

- 本系统仅用于病史采集，**不提供诊断和治疗建议**
- 如遇紧急情况，应**立即建议患者就医**
- 保护患者隐私，不泄露任何个人健康信息

当前时间: {{current_time}}
`,
			Temperature:                 0.5, // Moderate temperature for empathetic yet consistent responses
			MaxCompletionTokens:         4096,
			MaxIterations:               50,
			KBSelectionMode:             "all",
			RetrieveKBOnlyWhenMentioned: true, // Medical consultation doesn't need KB retrieval by default
			AllowedTools: []string{
				"thinking",
				"todo_write",
				"show_options",
			},
			WebSearchEnabled:    false, // Medical consultation doesn't need web search
			WebSearchMaxResults: 0,
			ReflectionEnabled:   true, // Enable reflection for better conversation flow
			MultiTurnEnabled:    true,
			HistoryTurns:        20, // More history for complete medical history tracking
			// Retrieval strategy (minimal, as we focus on conversation)
			EmbeddingTopK:    5,
			KeywordThreshold: 0.3,
			VectorThreshold:  0.5,
			RerankTopK:       5,
			RerankThreshold:  0.3,
		},
	}
}

// BuiltinAgentRegistry provides a registry of all built-in agents for easy extension
var BuiltinAgentRegistry = map[string]func(uint64) *CustomAgent{
	BuiltinQuickAnswerID:       GetBuiltinQuickAnswerAgent,
	BuiltinSmartReasoningID:    GetBuiltinSmartReasoningAgent,
	BuiltinDataAnalystID:       GetBuiltinDataAnalystAgent,
	BuiltinMedicalConsultantID: GetBuiltinMedicalConsultantAgent,
}

// builtinAgentIDsOrdered defines the fixed display order of built-in agents
var builtinAgentIDsOrdered = []string{
	BuiltinQuickAnswerID,
	BuiltinSmartReasoningID,
	BuiltinDeepResearcherID,
	BuiltinDataAnalystID,
	BuiltinKnowledgeGraphExpertID,
	BuiltinDocumentAssistantID,
	BuiltinMedicalConsultantID,
}

// GetBuiltinAgentIDs returns all built-in agent IDs in fixed order
func GetBuiltinAgentIDs() []string {
	return builtinAgentIDsOrdered
}

// IsBuiltinAgentID checks if the given ID is a built-in agent ID
func IsBuiltinAgentID(id string) bool {
	_, exists := BuiltinAgentRegistry[id]
	return exists
}

// GetBuiltinAgent returns a built-in agent by ID, or nil if not found
func GetBuiltinAgent(id string, tenantID uint64) *CustomAgent {
	if factory, exists := BuiltinAgentRegistry[id]; exists {
		return factory(tenantID)
	}
	return nil
}
