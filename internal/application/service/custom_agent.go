package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
)

// Custom agent related errors
var (
	ErrAgentNotFound       = errors.New("agent not found")
	ErrCannotModifyBuiltin = errors.New("cannot modify built-in agent basic info")
	ErrCannotDeleteBuiltin = errors.New("cannot delete built-in agent")
	ErrAgentNameRequired   = errors.New("agent name is required")
)

// customAgentService implements the CustomAgentService interface
type customAgentService struct {
	repo interfaces.CustomAgentRepository
}

// NewCustomAgentService creates a new custom agent service
func NewCustomAgentService(repo interfaces.CustomAgentRepository) interfaces.CustomAgentService {
	return &customAgentService{
		repo: repo,
	}
}

// CreateAgent creates a new custom agent
func (s *customAgentService) CreateAgent(ctx context.Context, agent *types.CustomAgent) (*types.CustomAgent, error) {
	// Validate required fields
	if strings.TrimSpace(agent.Name) == "" {
		return nil, ErrAgentNameRequired
	}

	// Generate UUID and set creation timestamps
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	// Get tenant ID from context
	tenantID, ok := ctx.Value(types.TenantIDContextKey).(uint64)
	if !ok {
		return nil, ErrInvalidTenantID
	}
	agent.TenantID = tenantID

	// Set timestamps
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()

	// Ensure agent mode is set for user-created agents
	if agent.Config.AgentMode == "" {
		agent.Config.AgentMode = types.AgentModeQuickAnswer
	}

	// Cannot create built-in agents
	agent.IsBuiltin = false

	// Set defaults
	agent.EnsureDefaults()

	logger.Infof(ctx, "Creating custom agent, ID: %s, tenant ID: %d, name: %s, agent_mode: %s",
		agent.ID, agent.TenantID, agent.Name, agent.Config.AgentMode)

	if err := s.repo.CreateAgent(ctx, agent); err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"agent_id":  agent.ID,
			"tenant_id": agent.TenantID,
		})
		return nil, err
	}

	logger.Infof(ctx, "Custom agent created successfully, ID: %s, name: %s", agent.ID, agent.Name)
	return agent, nil
}

// GetAgentByID retrieves an agent by its ID (including built-in agents)
func (s *customAgentService) GetAgentByID(ctx context.Context, id string) (*types.CustomAgent, error) {
	if id == "" {
		logger.Error(ctx, "Agent ID is empty")
		return nil, errors.New("agent ID cannot be empty")
	}

	// Get tenant ID from context
	tenantID, ok := ctx.Value(types.TenantIDContextKey).(uint64)
	if !ok {
		return nil, ErrInvalidTenantID
	}

	// Check if it's a built-in agent using the registry
	if types.IsBuiltinAgentID(id) {
		// Get the default built-in agent from registry (always has latest SystemPrompt)
		defaultAgent := types.GetBuiltinAgent(id, tenantID)
		if defaultAgent == nil {
			return nil, ErrAgentNotFound
		}

		// Try to get customized config from database
		dbAgent, err := s.repo.GetAgentByID(ctx, id, tenantID)
		if err == nil {
			// Merge: use code's SystemPrompt + user's customized config
			mergedAgent := s.mergeBuiltinAgentConfig(defaultAgent, dbAgent)
			return mergedAgent, nil
		}

		// Not in database, return default built-in agent
		return defaultAgent, nil
	}

	// Query from database
	agent, err := s.repo.GetAgentByID(ctx, id, tenantID)
	if err != nil {
		if errors.Is(err, repository.ErrCustomAgentNotFound) {
			return nil, ErrAgentNotFound
		}
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"agent_id": id,
		})
		return nil, err
	}

	return agent, nil
}

// mergeBuiltinAgentConfig merges the default built-in agent with user's customized config.
// SystemPrompt always uses the code version (latest), other configs use user's customized values.
func (s *customAgentService) mergeBuiltinAgentConfig(defaultAgent, dbAgent *types.CustomAgent) *types.CustomAgent {
	// Start with default agent as base
	merged := &types.CustomAgent{
		ID:          defaultAgent.ID,
		Name:        defaultAgent.Name,
		Description: defaultAgent.Description,
		Avatar:      defaultAgent.Avatar,
		IsBuiltin:   true,
		TenantID:    defaultAgent.TenantID,
		CreatedAt:   dbAgent.CreatedAt,
		UpdatedAt:   dbAgent.UpdatedAt,
	}

	// Use code's SystemPrompt (always latest)
	merged.Config.SystemPrompt = defaultAgent.Config.SystemPrompt
	merged.Config.AgentMode = defaultAgent.Config.AgentMode
	merged.Config.AllowedTools = defaultAgent.Config.AllowedTools
	merged.Config.SupportedFileTypes = defaultAgent.Config.SupportedFileTypes

	// Use user's customized config for other fields (if set)
	if dbAgent.Config.ModelID != "" {
		merged.Config.ModelID = dbAgent.Config.ModelID
	} else {
		merged.Config.ModelID = defaultAgent.Config.ModelID
	}

	// ReRank model: use user's value if set
	if dbAgent.Config.RerankModelID != "" {
		merged.Config.RerankModelID = dbAgent.Config.RerankModelID
	} else {
		merged.Config.RerankModelID = defaultAgent.Config.RerankModelID
	}

	// Context template: use user's value if set (for normal mode)
	if dbAgent.Config.ContextTemplate != "" {
		merged.Config.ContextTemplate = dbAgent.Config.ContextTemplate
	} else {
		merged.Config.ContextTemplate = defaultAgent.Config.ContextTemplate
	}

	// Temperature: use user's value if explicitly set (check if different from 0)
	if dbAgent.Config.Temperature != 0 {
		merged.Config.Temperature = dbAgent.Config.Temperature
	} else {
		merged.Config.Temperature = defaultAgent.Config.Temperature
	}

	// Thinking mode: use user's value
	if dbAgent.Config.Thinking != nil {
		merged.Config.Thinking = dbAgent.Config.Thinking
	} else {
		merged.Config.Thinking = defaultAgent.Config.Thinking
	}

	// Other configurable fields from user
	if dbAgent.Config.MaxCompletionTokens > 0 {
		merged.Config.MaxCompletionTokens = dbAgent.Config.MaxCompletionTokens
	} else {
		merged.Config.MaxCompletionTokens = defaultAgent.Config.MaxCompletionTokens
	}

	if dbAgent.Config.MaxIterations > 0 {
		merged.Config.MaxIterations = dbAgent.Config.MaxIterations
	} else {
		merged.Config.MaxIterations = defaultAgent.Config.MaxIterations
	}

	// Use user's retrieval config or defaults
	if dbAgent.Config.EmbeddingTopK > 0 {
		merged.Config.EmbeddingTopK = dbAgent.Config.EmbeddingTopK
	} else {
		merged.Config.EmbeddingTopK = defaultAgent.Config.EmbeddingTopK
	}

	if dbAgent.Config.KeywordThreshold > 0 {
		merged.Config.KeywordThreshold = dbAgent.Config.KeywordThreshold
	} else {
		merged.Config.KeywordThreshold = defaultAgent.Config.KeywordThreshold
	}

	if dbAgent.Config.VectorThreshold > 0 {
		merged.Config.VectorThreshold = dbAgent.Config.VectorThreshold
	} else {
		merged.Config.VectorThreshold = defaultAgent.Config.VectorThreshold
	}

	if dbAgent.Config.RerankTopK > 0 {
		merged.Config.RerankTopK = dbAgent.Config.RerankTopK
	} else {
		merged.Config.RerankTopK = defaultAgent.Config.RerankTopK
	}

	if dbAgent.Config.RerankThreshold > 0 {
		merged.Config.RerankThreshold = dbAgent.Config.RerankThreshold
	} else {
		merged.Config.RerankThreshold = defaultAgent.Config.RerankThreshold
	}

	// Boolean flags - use user's value if they have a database record
	merged.Config.WebSearchEnabled = dbAgent.Config.WebSearchEnabled
	merged.Config.WebSearchMaxResults = dbAgent.Config.WebSearchMaxResults
	merged.Config.ReflectionEnabled = dbAgent.Config.ReflectionEnabled
	merged.Config.MultiTurnEnabled = dbAgent.Config.MultiTurnEnabled
	merged.Config.HistoryTurns = dbAgent.Config.HistoryTurns
	merged.Config.KBSelectionMode = dbAgent.Config.KBSelectionMode
	merged.Config.KnowledgeBases = dbAgent.Config.KnowledgeBases
	merged.Config.RetrieveKBOnlyWhenMentioned = dbAgent.Config.RetrieveKBOnlyWhenMentioned

	// MCP configuration
	if dbAgent.Config.MCPSelectionMode != "" {
		merged.Config.MCPSelectionMode = dbAgent.Config.MCPSelectionMode
	} else {
		merged.Config.MCPSelectionMode = defaultAgent.Config.MCPSelectionMode
	}
	merged.Config.MCPServices = dbAgent.Config.MCPServices

	// FAQ strategy settings
	merged.Config.FAQPriorityEnabled = dbAgent.Config.FAQPriorityEnabled
	if dbAgent.Config.FAQDirectAnswerThreshold > 0 {
		merged.Config.FAQDirectAnswerThreshold = dbAgent.Config.FAQDirectAnswerThreshold
	} else {
		merged.Config.FAQDirectAnswerThreshold = defaultAgent.Config.FAQDirectAnswerThreshold
	}
	if dbAgent.Config.FAQScoreBoost > 0 {
		merged.Config.FAQScoreBoost = dbAgent.Config.FAQScoreBoost
	} else {
		merged.Config.FAQScoreBoost = defaultAgent.Config.FAQScoreBoost
	}

	// Advanced settings
	merged.Config.EnableQueryExpansion = dbAgent.Config.EnableQueryExpansion
	merged.Config.EnableRewrite = dbAgent.Config.EnableRewrite
	if dbAgent.Config.RewritePromptSystem != "" {
		merged.Config.RewritePromptSystem = dbAgent.Config.RewritePromptSystem
	} else {
		merged.Config.RewritePromptSystem = defaultAgent.Config.RewritePromptSystem
	}
	if dbAgent.Config.RewritePromptUser != "" {
		merged.Config.RewritePromptUser = dbAgent.Config.RewritePromptUser
	} else {
		merged.Config.RewritePromptUser = defaultAgent.Config.RewritePromptUser
	}
	if dbAgent.Config.FallbackStrategy != "" {
		merged.Config.FallbackStrategy = dbAgent.Config.FallbackStrategy
	} else {
		merged.Config.FallbackStrategy = defaultAgent.Config.FallbackStrategy
	}
	if dbAgent.Config.FallbackResponse != "" {
		merged.Config.FallbackResponse = dbAgent.Config.FallbackResponse
	} else {
		merged.Config.FallbackResponse = defaultAgent.Config.FallbackResponse
	}
	if dbAgent.Config.FallbackPrompt != "" {
		merged.Config.FallbackPrompt = dbAgent.Config.FallbackPrompt
	} else {
		merged.Config.FallbackPrompt = defaultAgent.Config.FallbackPrompt
	}

	// Set defaults if needed
	if merged.Config.KBSelectionMode == "" {
		merged.Config.KBSelectionMode = defaultAgent.Config.KBSelectionMode
	}
	if merged.Config.HistoryTurns == 0 {
		merged.Config.HistoryTurns = defaultAgent.Config.HistoryTurns
	}

	return merged
}

// ListAgents lists all agents for the current tenant (including built-in agents)
func (s *customAgentService) ListAgents(ctx context.Context) ([]*types.CustomAgent, error) {
	tenantID, ok := ctx.Value(types.TenantIDContextKey).(uint64)
	if !ok {
		return nil, ErrInvalidTenantID
	}

	// Get all agents from database (including built-in agents with customized config)
	allAgents, err := s.repo.ListAgentsByTenantID(ctx, tenantID)
	if err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"tenant_id": tenantID,
		})
		return nil, err
	}

	// Track which built-in agents exist in database
	builtinInDB := make(map[string]bool)
	for _, agent := range allAgents {
		if types.IsBuiltinAgentID(agent.ID) {
			builtinInDB[agent.ID] = true
		}
	}

	// Build result: built-in agents first, then custom agents
	builtinIDs := types.GetBuiltinAgentIDs()
	result := make([]*types.CustomAgent, 0, len(allAgents)+len(builtinIDs))

	// Add built-in agents in order
	for _, builtinID := range builtinIDs {
		defaultAgent := types.GetBuiltinAgent(builtinID, tenantID)
		if defaultAgent == nil {
			continue
		}

		if builtinInDB[builtinID] {
			// Merge code's SystemPrompt with user's customized config
			for _, dbAgent := range allAgents {
				if dbAgent.ID == builtinID {
					merged := s.mergeBuiltinAgentConfig(defaultAgent, dbAgent)
					result = append(result, merged)
					break
				}
			}
		} else {
			// Use default built-in agent
			result = append(result, defaultAgent)
		}
	}

	// Add custom agents
	for _, agent := range allAgents {
		if !types.IsBuiltinAgentID(agent.ID) {
			result = append(result, agent)
		}
	}

	return result, nil
}

// UpdateAgent updates an agent's information
func (s *customAgentService) UpdateAgent(ctx context.Context, agent *types.CustomAgent) (*types.CustomAgent, error) {
	if agent.ID == "" {
		logger.Error(ctx, "Agent ID is empty")
		return nil, errors.New("agent ID cannot be empty")
	}

	// Get tenant ID from context
	tenantID, ok := ctx.Value(types.TenantIDContextKey).(uint64)
	if !ok {
		return nil, ErrInvalidTenantID
	}

	// Handle built-in agents specially using registry
	if types.IsBuiltinAgentID(agent.ID) {
		return s.updateBuiltinAgent(ctx, agent, tenantID)
	}

	// Get existing agent
	existingAgent, err := s.repo.GetAgentByID(ctx, agent.ID, tenantID)
	if err != nil {
		if errors.Is(err, repository.ErrCustomAgentNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}

	// Cannot modify built-in status
	if existingAgent.IsBuiltin {
		return nil, ErrCannotModifyBuiltin
	}

	// Validate name
	if strings.TrimSpace(agent.Name) == "" {
		return nil, ErrAgentNameRequired
	}

	// Update fields
	existingAgent.Name = agent.Name
	existingAgent.Description = agent.Description
	existingAgent.Avatar = agent.Avatar
	existingAgent.Config = agent.Config
	existingAgent.UpdatedAt = time.Now()

	// Ensure defaults
	existingAgent.EnsureDefaults()

	logger.Infof(ctx, "Updating custom agent, ID: %s, name: %s", agent.ID, agent.Name)

	if err := s.repo.UpdateAgent(ctx, existingAgent); err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"agent_id": agent.ID,
		})
		return nil, err
	}

	logger.Infof(ctx, "Custom agent updated successfully, ID: %s", agent.ID)
	return existingAgent, nil
}

// updateBuiltinAgent updates a built-in agent's configuration (but not basic info)
func (s *customAgentService) updateBuiltinAgent(ctx context.Context, agent *types.CustomAgent, tenantID uint64) (*types.CustomAgent, error) {
	// Get the default built-in agent from registry
	defaultAgent := types.GetBuiltinAgent(agent.ID, tenantID)
	if defaultAgent == nil {
		return nil, ErrAgentNotFound
	}

	// Try to get existing customized config from database
	existingAgent, err := s.repo.GetAgentByID(ctx, agent.ID, tenantID)
	if err != nil && !errors.Is(err, repository.ErrCustomAgentNotFound) {
		return nil, err
	}

	if existingAgent != nil {
		// Update existing record - only update config, keep basic info unchanged
		existingAgent.Config = agent.Config
		existingAgent.UpdatedAt = time.Now()
		existingAgent.EnsureDefaults()

		logger.Infof(ctx, "Updating built-in agent config, ID: %s", agent.ID)

		if err := s.repo.UpdateAgent(ctx, existingAgent); err != nil {
			logger.ErrorWithFields(ctx, err, map[string]interface{}{
				"agent_id": agent.ID,
			})
			return nil, err
		}

		logger.Infof(ctx, "Built-in agent config updated successfully, ID: %s", agent.ID)
		return existingAgent, nil
	}

	// Create new record for built-in agent with customized config
	newAgent := &types.CustomAgent{
		ID:          defaultAgent.ID,
		Name:        defaultAgent.Name,
		Description: defaultAgent.Description,
		Avatar:      defaultAgent.Avatar,
		IsBuiltin:   true,
		TenantID:    tenantID,
		Config:      agent.Config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	newAgent.EnsureDefaults()

	logger.Infof(ctx, "Creating built-in agent config record, ID: %s, tenant ID: %d", agent.ID, tenantID)

	if err := s.repo.CreateAgent(ctx, newAgent); err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"agent_id":  agent.ID,
			"tenant_id": tenantID,
		})
		return nil, err
	}

	logger.Infof(ctx, "Built-in agent config record created successfully, ID: %s", agent.ID)
	return newAgent, nil
}

// DeleteAgent deletes an agent
func (s *customAgentService) DeleteAgent(ctx context.Context, id string) error {
	if id == "" {
		logger.Error(ctx, "Agent ID is empty")
		return errors.New("agent ID cannot be empty")
	}

	// Cannot delete built-in agents using registry check
	if types.IsBuiltinAgentID(id) {
		return ErrCannotDeleteBuiltin
	}

	// Get tenant ID from context
	tenantID, ok := ctx.Value(types.TenantIDContextKey).(uint64)
	if !ok {
		return ErrInvalidTenantID
	}

	// Get existing agent to verify ownership
	existingAgent, err := s.repo.GetAgentByID(ctx, id, tenantID)
	if err != nil {
		if errors.Is(err, repository.ErrCustomAgentNotFound) {
			return ErrAgentNotFound
		}
		return err
	}

	// Cannot delete built-in agents
	if existingAgent.IsBuiltin {
		return ErrCannotDeleteBuiltin
	}

	logger.Infof(ctx, "Deleting custom agent, ID: %s", id)

	if err := s.repo.DeleteAgent(ctx, id, tenantID); err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"agent_id": id,
		})
		return err
	}

	logger.Infof(ctx, "Custom agent deleted successfully, ID: %s", id)
	return nil
}

// CopyAgent creates a copy of an existing agent
func (s *customAgentService) CopyAgent(ctx context.Context, id string) (*types.CustomAgent, error) {
	if id == "" {
		logger.Error(ctx, "Agent ID is empty")
		return nil, errors.New("agent ID cannot be empty")
	}

	// Get tenant ID from context
	tenantID, ok := ctx.Value(types.TenantIDContextKey).(uint64)
	if !ok {
		return nil, ErrInvalidTenantID
	}

	// Get the source agent
	sourceAgent, err := s.GetAgentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a new agent with copied data
	newAgent := &types.CustomAgent{
		ID:          uuid.New().String(),
		Name:        sourceAgent.Name + " (副本)",
		Description: sourceAgent.Description,
		Avatar:      sourceAgent.Avatar,
		IsBuiltin:   false, // Copied agents are never built-in
		TenantID:    tenantID,
		Config:      sourceAgent.Config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Ensure defaults
	newAgent.EnsureDefaults()

	logger.Infof(ctx, "Copying agent, source ID: %s, new ID: %s", id, newAgent.ID)

	if err := s.repo.CreateAgent(ctx, newAgent); err != nil {
		logger.ErrorWithFields(ctx, err, map[string]interface{}{
			"source_agent_id": id,
			"new_agent_id":    newAgent.ID,
		})
		return nil, err
	}

	logger.Infof(ctx, "Agent copied successfully, source ID: %s, new ID: %s", id, newAgent.ID)
	return newAgent, nil
}
