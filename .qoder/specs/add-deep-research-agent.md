# Deep Research Agent 实现方案

## 概述

为 WiseDx 添加"深度研究员"内置智能体，专注于知识库深度研究（不含联网搜索），并配备定制的研究进度 UI 组件。

## 需修改的文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/types/custom_agent.go` | 修改 | 添加工厂函数 + 注册到 Registry |
| `frontend/src/api/agent/index.ts` | 修改 | 添加 `BUILTIN_DEEP_RESEARCHER_ID` 常量 |
| `frontend/src/views/chat/components/ResearchProgress.vue` | **新建** | 研究进度组件（仿 ConsultationProgress） |
| `frontend/src/views/chat/components/AgentStreamDisplay.vue` | 修改 | 接入 ResearchProgress 组件 |
| `frontend/src/i18n/locales/zh-CN.ts` | 修改 | 添加 `researchProgress` 翻译 |
| `frontend/src/i18n/locales/en-US.ts` | 修改 | 添加 `researchProgress` 翻译 |

---

## 步骤 1: 后端 - 添加工厂函数并注册

**文件**: `internal/types/custom_agent.go`

### 1a. 添加 `GetBuiltinDeepResearcherAgent` 函数

在 `GetBuiltinDataAnalystAgent` 函数之前（约第 318 行），添加新函数。

**Agent 配置:**

```go
func GetBuiltinDeepResearcherAgent(tenantID uint64) *CustomAgent {
    return &CustomAgent{
        ID:          BuiltinDeepResearcherID,
        Name:        "深度研究员",
        Description: "专注于深度研究和综合分析，能够制定研究计划、多维度检索信息、深入思考并给出全面的分析报告",
        Avatar:      "🔬",
        IsBuiltin:   true,
        TenantID:    tenantID,
        Config: CustomAgentConfig{
            AgentMode:    AgentModeSmartReasoning,
            Temperature:  0.3,
            MaxCompletionTokens: 8192,
            MaxIterations: 50,
            KBSelectionMode: "all",
            RetrieveKBOnlyWhenMentioned: false,
            AllowedTools: []string{
                "thinking",
                "todo_write",
                "knowledge_search",
                "grep_chunks",
                "query_knowledge_graph",
                "get_document_info",
                "list_knowledge_chunks",
            },
            WebSearchEnabled:    false,
            WebSearchMaxResults: 0,
            ReflectionEnabled:   true,
            MultiTurnEnabled:    true,
            HistoryTurns:        10,
            FAQPriorityEnabled:       true,
            FAQDirectAnswerThreshold: 0.9,
            FAQScoreBoost:            1.2,
            EmbeddingTopK:    10,
            KeywordThreshold: 0.3,
            VectorThreshold:  0.5,
            RerankTopK:       10,
            RerankThreshold:  0.3,
            SystemPrompt: `...见下方...`,
        },
    }
}
```

**System Prompt 核心内容:**

```
### Role
你是 WiseDx 深度研究员，一个专注于知识库深度研究和综合分析的专业研究智能体。

### Mission
针对用户的研究课题，制定系统性研究计划，通过多维度检索知识库中的信息，进行深入思考分析，最终生成全面、结构化的研究报告。

### Critical Constraints
1. **知识库专注**：所有研究素材必须来自知识库，不使用网络搜索
2. **先规划再执行**：必须先调用 todo_write 建立研究计划
3. **多维检索**：对同一主题使用多种检索方式（语义搜索 knowledge_search + 关键词搜索 grep_chunks + 图谱查询 query_knowledge_graph）
4. **深度综合**：不简单堆砌搜索结果，必须通过 thinking 进行深度分析和综合

### Workflow（5个研究阶段）

#### 阶段1: 研究规划
- 分析用户课题，拆解核心问题
- 确定研究方向和检索关键词
- 调用 todo_write 建立研究计划

#### 阶段2: 资料检索
- 使用 knowledge_search 进行语义检索
- 使用 grep_chunks 进行关键词精确匹配
- 使用 query_knowledge_graph 探索实体关系
- 使用 get_document_info 了解文档元数据
- 使用 list_knowledge_chunks 精读关键文档

#### 阶段3: 深度分析
- 使用 thinking 对检索结果进行深入分析
- 识别关键发现、模式和关联
- 标记信息缺口和矛盾之处

#### 阶段4: 交叉验证
- 对重要发现使用不同检索方式验证
- 补充缺失信息
- 解决矛盾点

#### 阶段5: 报告生成
- 生成结构化 Markdown 研究报告

### 【最重要】首次响应必须调用 todo_write

在你的第一条响应中，必须立即调用 todo_write 创建研究计划：

{
  "task": "深度研究: [用户研究课题]",
  "steps": [
    {"id": "step1", "description": "研究规划", "status": "in_progress"},
    {"id": "step2", "description": "资料检索", "status": "pending"},
    {"id": "step3", "description": "深度分析", "status": "pending"},
    {"id": "step4", "description": "交叉验证", "status": "pending"},
    {"id": "step5", "description": "报告生成", "status": "pending"}
  ]
}

### 阶段完整后才更新 todo_write
只有当当前阶段的任务完成后，才调用 todo_write 更新状态并进入下一阶段。

### 工具使用指南

**每轮回复的调用顺序：**
1. **thinking**（必调，但仅调用一次）：分析当前研究进展，规划下一步
2. **检索工具**（按需）：knowledge_search / grep_chunks / query_knowledge_graph / get_document_info / list_knowledge_chunks
3. **todo_write**（仅阶段完成时）：更新研究进度

⚠️ 防止思考死循环规则（同医疗问诊Agent）：
- 每轮只能调用 thinking 一次
- thinking 后必须立即执行下一步动作

### 研究报告输出格式

使用 Markdown 格式输出，包含：
- **研究摘要**：核心发现概述
- **主要发现**：按主题分点阐述
- **详细分析**：按研究角度展开论述
- **知识图谱关联**：实体关系发现（如适用）
- **研究局限性**：信息缺口和不确定性说明
- **参考来源**：引用的知识库文档和分块ID

Current Time: {{current_time}}
```

### 1b. 注册到 BuiltinAgentRegistry

在 `BuiltinAgentRegistry` map（约第 989 行）中添加:

```go
BuiltinDeepResearcherID: GetBuiltinDeepResearcherAgent,
```

注意: `builtinAgentIDsOrdered` 已包含 `BuiltinDeepResearcherID`（第 1001 行），无需修改。

---

## 步骤 2: 前端 - 添加常量

**文件**: `frontend/src/api/agent/index.ts`

在第 100 行 `BUILTIN_MEDICAL_CONSULTANT_ID` 后面添加:

```typescript
export const BUILTIN_DEEP_RESEARCHER_ID = 'builtin-deep-researcher';
```

---

## 步骤 3: 前端 - 新建 ResearchProgress 组件

**文件**: `frontend/src/views/chat/components/ResearchProgress.vue` (新建)

仿照 `ConsultationProgress.vue` 的完整结构，差异点:

| 对比项 | ConsultationProgress | ResearchProgress |
|--------|---------------------|------------------|
| agentId 检查 | `builtin-medical-consultant` | `builtin-deep-researcher` |
| STAGE_NAMES | 6 阶段 (问候~信息总结) | 5 阶段 (研究规划~报告生成) |
| 图标 emoji | 📊 | 🔬 |
| i18n title key | `agent.consultationProgress` | `agent.researchProgress` |
| 主色调 | 绿色 `#52c41a` / `#389e0d` | 蓝紫色 `#722ed1` / `#531dab` |
| 背景渐变 | `#f6ffed → #d9f7be` | `#f9f0ff → #efdbff` |
| 边框色 | `#b7eb8f` | `#d3adf7` |
| 暗色模式背景 | `#162312 → #1d3712` | `#120338 → #1a0a4e` |
| 暗色边框 | `#49aa19` | `#9254de` |
| 暗色文字 | `#95de64` | `#b37feb` |

**STAGE_NAMES 映射表:**
```typescript
const STAGE_NAMES: Record<string, string> = {
  'step1': '研究规划',
  'step2': '资料检索',
  'step3': '深度分析',
  'step4': '交叉验证',
  'step5': '报告生成',
};
```

其他所有 computed 逻辑（currentStage、totalStages、progressPercent、currentStageName）与 ConsultationProgress 完全一致。

---

## 步骤 4: 前端 - 接入 AgentStreamDisplay

**文件**: `frontend/src/views/chat/components/AgentStreamDisplay.vue`

### 4a. 导入组件

在现有 `import ConsultationProgress` 附近添加:
```typescript
import ResearchProgress from './ResearchProgress.vue';
```

### 4b. 添加 computed 属性

在 `shouldShowConsultationProgress`（第 454 行）之后添加:
```typescript
const shouldShowResearchProgress = computed(() => {
  return props.agentId === 'builtin-deep-researcher' && latestPlanSteps.value.length > 0;
});
```

`latestPlanSteps` 已存在（第 438 行），两个 Agent 共享此逻辑，无需修改。

### 4c. 模板中添加组件

在 `ConsultationProgress`（第 5-10 行）之后添加:
```html
<ResearchProgress
  v-if="shouldShowResearchProgress"
  :steps="latestPlanSteps"
  :visible="true"
  :agent-id="agentId"
/>
```

---

## 步骤 5: 前端 - 添加 i18n 翻译

**文件**: `frontend/src/i18n/locales/zh-CN.ts`

在 `consultationProgress: "问诊进度"` (第 762 行) 附近添加:
```typescript
researchProgress: "研究进度",
```

**文件**: `frontend/src/i18n/locales/en-US.ts`

在 `consultationProgress: 'Consultation Progress'` (第 215 行) 附近添加:
```typescript
researchProgress: 'Research Progress',
```

---

## 验证方案

### 后端验证
1. `go build ./...` 确认编译通过
2. 检查 `/api/v1/agents` 列表接口返回的 agents 中包含 `builtin-deep-researcher`

### 前端验证
1. `cd frontend && npm run build` 确认构建通过
2. UI 验证:
   - Agent 选择器中显示"深度研究员"并带有 🔬 emoji
   - 选择深度研究员后发送研究问题，观察:
     - Agent 首先调用 todo_write 建立 5 阶段研究计划
     - ResearchProgress 进度条正确显示（蓝紫色主题）
     - 各阶段推进时进度条同步更新
     - 最终生成 Markdown 格式研究报告
