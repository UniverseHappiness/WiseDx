# 代理定制API

<cite>
**本文档引用的文件**
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go)
- [internal/router/router.go](file://internal/router/router.go)
- [internal/application/service/custom_agent.go](file://internal/application/service/custom_agent.go)
- [internal/types/custom_agent.go](file://internal/types/custom_agent.go)
- [internal/agent/engine.go](file://internal/agent/engine.go)
- [internal/agent/tools/registry.go](file://internal/agent/tools/registry.go)
- [internal/handler/mcp_service.go](file://internal/handler/mcp_service.go)
- [internal/application/service/mcp_service.go](file://internal/application/service/mcp_service.go)
- [internal/types/mcp.go](file://internal/types/mcp.go)
- [internal/middleware/auth.go](file://internal/middleware/auth.go)
- [frontend/src/api/agent/index.ts](file://frontend/src/api/agent/index.ts)
- [frontend/src/views/agent/AgentEditorModal.vue](file://frontend/src/views/agent/AgentEditorModal.vue)
- [frontend/src/views/settings/AgentSettings.vue](file://frontend/src/views/settings/AgentSettings.vue)
</cite>

## 目录
1. [简介](#简介)
2. [项目结构](#项目结构)
3. [核心组件](#核心组件)
4. [架构概览](#架构概览)
5. [详细组件分析](#详细组件分析)
6. [依赖关系分析](#依赖关系分析)
7. [性能考虑](#性能考虑)
8. [故障排除指南](#故障排除指南)
9. [结论](#结论)

## 简介

WiseDx的代理定制API为智能代理提供了完整的生命周期管理能力，包括创建、配置、部署、监控等全栈功能。该系统支持两种代理模式：快速问答模式（RAG）和智能推理模式（ReAct），并提供了丰富的工具系统和MCP（Model Context Protocol）集成能力。

系统采用分层架构设计，通过HTTP API提供RESTful接口，支持JWT和API Key双重认证机制。代理配置采用JSON Schema定义，支持参数化设置和行为控制。系统内置了多种预置代理，包括医疗问诊助手、诊断报告生成器等专业应用场景。

## 项目结构

WiseDx代理定制API采用典型的三层架构模式：

```mermaid
graph TB
subgraph "前端层"
FE[Vue.js 前端应用]
API[Agent API]
end
subgraph "接口层"
Router[路由路由器]
Handler[处理器层]
Auth[认证中间件]
end
subgraph "业务逻辑层"
Service[服务层]
Engine[代理引擎]
Tools[工具系统]
end
subgraph "数据访问层"
Repo[仓库层]
DB[(数据库)]
end
FE --> API
API --> Router
Router --> Handler
Handler --> Auth
Auth --> Service
Service --> Engine
Engine --> Tools
Service --> Repo
Repo --> DB
```

**图表来源**
- [internal/router/router.go](file://internal/router/router.go#L422-L441)
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go#L1-L372)

**章节来源**
- [internal/router/router.go](file://internal/router/router.go#L422-L441)
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go#L1-L372)

## 核心组件

### 代理生命周期管理

系统提供完整的代理生命周期管理接口：

| 操作 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 创建代理 | POST | `/api/v1/agents` | 创建新的自定义代理 |
| 获取代理列表 | GET | `/api/v1/agents` | 获取当前租户的所有代理 |
| 获取代理详情 | GET | `/api/v1/agents/:id` | 根据ID获取代理详情 |
| 更新代理 | PUT | `/api/v1/agents/:id` | 更新代理配置 |
| 删除代理 | DELETE | `/api/v1/agents/:id` | 删除指定代理 |
| 复制代理 | POST | `/api/v1/agents/:id/copy` | 复制指定代理 |

### 代理配置系统

代理配置采用嵌套结构，支持多种配置维度：

```mermaid
classDiagram
class CustomAgent {
+string ID
+string Name
+string Description
+string Avatar
+bool IsBuiltin
+uint64 TenantID
+CustomAgentConfig Config
+EnsureDefaults()
+IsAgentMode() bool
}
class CustomAgentConfig {
+string AgentMode
+string SystemPrompt
+string ContextTemplate
+string ModelID
+float64 Temperature
+int MaxIterations
+[]string AllowedTools
+string KBSelectionMode
+[]string KnowledgeBases
+bool WebSearchEnabled
+int HistoryTurns
+int EmbeddingTopK
+float64 KeywordThreshold
+bool FAQPriorityEnabled
+string MCPSelectionMode
+[]string MCPServices
}
class MCPService {
+string ID
+string Name
+bool Enabled
+MCPTransportType TransportType
+string URL
+MCPHeaders Headers
+MCPAuthConfig AuthConfig
+MCPAdvancedConfig AdvancedConfig
}
CustomAgent --> CustomAgentConfig
CustomAgentConfig --> MCPService
```

**图表来源**
- [internal/types/custom_agent.go](file://internal/types/custom_agent.go#L40-L164)
- [internal/types/mcp.go](file://internal/types/mcp.go#L22-L38)

**章节来源**
- [internal/types/custom_agent.go](file://internal/types/custom_agent.go#L40-L164)
- [internal/types/mcp.go](file://internal/types/mcp.go#L22-L38)

## 架构概览

代理系统采用事件驱动架构，通过事件总线实现模块解耦：

```mermaid
sequenceDiagram
participant Client as 客户端
participant API as API网关
participant Handler as 处理器
participant Service as 服务层
participant Engine as 代理引擎
participant Tools as 工具系统
participant EventBus as 事件总线
Client->>API : 创建代理请求
API->>Handler : 路由到代理处理器
Handler->>Service : 调用代理服务
Service->>Service : 验证和持久化
Service-->>Handler : 返回代理对象
Handler-->>Client : 返回创建结果
Client->>API : 启动代理执行
API->>Handler : 路由到代理处理器
Handler->>Service : 获取代理配置
Service->>Engine : 创建代理引擎
Engine->>Engine : 初始化工具注册表
Engine->>Tools : 加载可用工具
Engine->>EventBus : 发送执行事件
Tools-->>Engine : 执行工具结果
Engine->>EventBus : 发送工具结果事件
Engine-->>Handler : 返回执行状态
Handler-->>Client : 返回执行结果
```

**图表来源**
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go#L55-L97)
- [internal/agent/engine.go](file://internal/agent/engine.go#L77-L155)

**章节来源**
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go#L55-L97)
- [internal/agent/engine.go](file://internal/agent/engine.go#L77-L155)

## 详细组件分析

### 代理处理器层

代理处理器负责HTTP请求的接收和响应处理：

```mermaid
flowchart TD
Start([请求到达]) --> Validate[参数验证]
Validate --> Parse[解析请求体]
Parse --> Sanitize[清理输入数据]
Sanitize --> CallService[调用服务层]
CallService --> HandleError{处理错误}
HandleError --> |错误| ReturnError[返回错误响应]
HandleError --> |成功| BuildResponse[构建响应]
BuildResponse --> End([返回成功响应])
ReturnError --> End
```

**图表来源**
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go#L60-L66)

**章节来源**
- [internal/handler/custom_agent.go](file://internal/handler/custom_agent.go#L55-L97)

### 代理服务层

服务层实现业务逻辑和数据验证：

| 方法 | 功能 | 错误处理 |
|------|------|----------|
| CreateAgent | 创建新代理 | 名称必填、租户ID验证 |
| GetAgentByID | 获取代理详情 | 代理不存在、内置代理合并 |
| ListAgents | 获取代理列表 | 租户隔离、内置代理合并 |
| UpdateAgent | 更新代理 | 权限检查、内置代理保护 |
| DeleteAgent | 删除代理 | 内置代理保护、租户验证 |
| CopyAgent | 复制代理 | 数据完整性保证 |

**章节来源**
- [internal/application/service/custom_agent.go](file://internal/application/service/custom_agent.go#L36-L83)

### 代理引擎

代理引擎是ReAct推理的核心执行单元：

```mermaid
stateDiagram-v2
[*] --> 初始化
初始化 --> 思考阶段
思考阶段 --> 工具调用
工具调用 --> 观察阶段
观察阶段 --> 思考阶段
思考阶段 --> 完成条件
完成条件 --> 最终回答
最终回答 --> [*]
思考阶段 --> 终止 : 达到最大迭代次数
工具调用 --> 终止 : 工具执行失败
```

**图表来源**
- [internal/agent/engine.go](file://internal/agent/engine.go#L159-L521)

**章节来源**
- [internal/agent/engine.go](file://internal/agent/engine.go#L159-L521)

### 工具注册系统

工具系统提供灵活的扩展机制：

```mermaid
classDiagram
class ToolRegistry {
+map[string]Tool tools
+RegisterTool(tool)
+GetTool(name) Tool
+ListTools() []string
+ExecuteTool(ctx, name, args) ToolResult
+Cleanup(ctx)
}
class Tool {
<<interface>>
+Name() string
+Description() string
+Parameters() json.RawMessage
+Execute(ctx, args) ToolResult
}
class BaseTool {
+string name
+string description
+json schema
+Name() string
+Description() string
+Parameters() json.RawMessage
}
ToolRegistry --> Tool
BaseTool ..|> Tool
```

**图表来源**
- [internal/agent/tools/registry.go](file://internal/agent/tools/registry.go#L12-L58)

**章节来源**
- [internal/agent/tools/registry.go](file://internal/agent/tools/registry.go#L12-L58)

### MCP服务集成

MCP（Model Context Protocol）提供外部工具和服务集成：

```mermaid
sequenceDiagram
participant Agent as 代理
participant MCP as MCP服务
participant Manager as MCP管理器
participant Client as MCP客户端
Agent->>Manager : 请求MCP工具
Manager->>Client : 获取或创建客户端
Client->>MCP : 连接服务
MCP-->>Client : 返回工具列表
Client-->>Manager : 返回可用工具
Manager-->>Agent : 返回工具定义
Agent->>Client : 执行工具调用
Client->>MCP : 调用远程工具
MCP-->>Client : 返回执行结果
Client-->>Agent : 返回工具结果
```

**图表来源**
- [internal/application/service/mcp_service.go](file://internal/application/service/mcp_service.go#L323-L350)

**章节来源**
- [internal/application/service/mcp_service.go](file://internal/application/service/mcp_service.go#L323-L350)

## 依赖关系分析

系统采用清晰的依赖层次结构：

```mermaid
graph TB
subgraph "外部依赖"
Gin[Gin Web框架]
JWT[JWT认证]
GORM[GORM ORM]
end
subgraph "内部模块"
Handler[处理器层]
Service[服务层]
Engine[代理引擎]
Tools[工具系统]
MCP[MCP系统]
Types[数据类型]
Middleware[中间件]
end
subgraph "基础设施"
DB[(数据库)]
Redis[(Redis缓存)]
Event[(事件总线)]
end
Gin --> Handler
JWT --> Middleware
GORM --> Service
Handler --> Service
Service --> Engine
Engine --> Tools
Engine --> MCP
Service --> DB
Service --> Event
Tools --> Event
MCP --> Event
Types --> Handler
Types --> Service
Middleware --> Handler
```

**图表来源**
- [internal/router/router.go](file://internal/router/router.go#L422-L441)
- [internal/middleware/auth.go](file://internal/middleware/auth.go#L59-L196)

**章节来源**
- [internal/router/router.go](file://internal/router/router.go#L422-L441)
- [internal/middleware/auth.go](file://internal/middleware/auth.go#L59-L196)

## 性能考虑

### 缓存策略

系统实现了多层次缓存机制：

- **工具缓存**: 工具注册表缓存，避免重复初始化
- **代理配置缓存**: 内置代理配置缓存，减少数据库查询
- **MCP连接池**: 复用MCP客户端连接，降低连接开销

### 并发控制

- **请求限流**: 基于令牌桶算法的请求频率控制
- **并发安全**: 工具执行的互斥锁保护
- **资源池**: 代理引擎的资源池管理

### 监控指标

系统提供完整的性能监控：

- **执行时间**: 代理执行各阶段耗时统计
- **工具调用**: 工具执行成功率和耗时
- **内存使用**: 工具和代理的内存占用监控
- **错误率**: 各组件的错误统计和告警

## 故障排除指南

### 常见错误类型

| 错误类型 | HTTP状态码 | 描述 | 解决方案 |
|----------|------------|------|----------|
| 参数错误 | 400 | 请求参数无效 | 检查请求格式和必填字段 |
| 未授权 | 401 | 认证失败 | 验证JWT或API Key |
| 禁止访问 | 403 | 权限不足 | 检查用户权限和租户访问 |
| 资源不存在 | 404 | 代理或工具不存在 | 确认ID有效性 |
| 服务器错误 | 500 | 系统内部错误 | 查看日志和监控指标 |

### 调试方法

1. **启用调试日志**: 设置日志级别为DEBUG
2. **检查事件流**: 监听事件总线查看执行状态
3. **验证配置**: 确认代理配置的完整性和有效性
4. **测试工具**: 单独测试工具的可用性和参数

### 性能优化建议

- **合理配置**: 根据使用场景调整代理参数
- **资源限制**: 设置合理的超时和重试机制
- **监控告警**: 建立完善的监控和告警体系
- **定期维护**: 清理无用的代理和工具配置

**章节来源**
- [internal/middleware/auth.go](file://internal/middleware/auth.go#L19-L39)

## 结论

WiseDx的代理定制API提供了一个功能完整、架构清晰的智能代理管理平台。系统支持多种代理模式和工具集成，具有良好的扩展性和可维护性。

通过分层架构设计和事件驱动机制，系统实现了模块间的松耦合，便于功能扩展和性能优化。内置的监控和日志系统为运维提供了有力支撑。

未来可以在以下方面进一步完善：
- 增加更多预置代理模板
- 优化工具加载和执行性能
- 扩展MCP协议支持
- 增强安全沙箱机制
- 完善自动化部署和扩缩容能力