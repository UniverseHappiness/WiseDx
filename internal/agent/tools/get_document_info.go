package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/UniverseHappiness/WiseDx/internal/types"
	"github.com/UniverseHappiness/WiseDx/internal/types/interfaces"
	"github.com/UniverseHappiness/WiseDx/internal/utils"
)

var getDocumentInfoTool = BaseTool{
	name: ToolGetDocumentInfo,
	description: `Retrieve detailed metadata information about documents.

## When to Use

Use this tool when:
- Need to understand document basic information (title, type, size, etc.)
- Check if document exists and is available
- Batch query metadata for multiple documents
- Understand document processing status

Do not use when:
- Need document content (use knowledge_search)
- Need specific text chunks (search results already contain full content)


## Returned Information

- Basic info: title, description, source type
- File info: filename, type, size
- Processing status: whether processed, chunk count
- Metadata: custom tags and properties


## Notes

- Concurrent query for multiple documents provides better performance
- Returns complete document metadata, not just title
- Can check document processing status (parse_status)`,
	schema: utils.GenerateSchema[GetDocumentInfoInput](),
}

// GetDocumentInfoInput defines the input parameters for get document info tool
type GetDocumentInfoInput struct {
	KnowledgeIDs []string `json:"knowledge_ids" jsonschema:"Array of document/knowledge IDs, obtained from knowledge_id field in search results, supports concurrent batch queries"`
}

// GetDocumentInfoTool retrieves detailed information about a document/knowledge
type GetDocumentInfoTool struct {
	BaseTool
	knowledgeService interfaces.KnowledgeService
	chunkService     interfaces.ChunkService
}

// NewGetDocumentInfoTool creates a new get document info tool
func NewGetDocumentInfoTool(
	knowledgeService interfaces.KnowledgeService,
	chunkService interfaces.ChunkService,
) *GetDocumentInfoTool {
	return &GetDocumentInfoTool{
		BaseTool:         getDocumentInfoTool,
		knowledgeService: knowledgeService,
		chunkService:     chunkService,
	}
}

// Execute retrieves document information with concurrent processing
func (t *GetDocumentInfoTool) Execute(ctx context.Context, args json.RawMessage) (*types.ToolResult, error) {
	tenantID := uint64(0)
	if tid, ok := ctx.Value(types.TenantIDContextKey).(uint64); ok {
		tenantID = tid
	}

	// Parse args from json.RawMessage
	var input GetDocumentInfoInput
	if err := json.Unmarshal(args, &input); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse args: %v", err),
		}, err
	}

	// Extract knowledge_ids array
	knowledgeIDs := input.KnowledgeIDs
	if len(knowledgeIDs) == 0 {
		return &types.ToolResult{
			Success: false,
			Error:   "knowledge_ids is required and must be a non-empty array",
		}, fmt.Errorf("knowledge_ids is required")
	}

	// Validate max 10 documents
	if len(knowledgeIDs) > 10 {
		return &types.ToolResult{
			Success: false,
			Error:   "knowledge_ids must contain at least one valid knowledge ID",
		}, fmt.Errorf("no valid knowledge IDs provided")
	}

	// Concurrently get info for each knowledge ID
	type docInfo struct {
		knowledge  *types.Knowledge
		chunkCount int
		err        error
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]*docInfo)

	// Concurrently get info for each knowledge ID
	for _, knowledgeID := range knowledgeIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// Get knowledge metadata
			knowledge, err := t.knowledgeService.GetRepository().GetKnowledgeByID(ctx, tenantID, id)
			if err != nil {
				mu.Lock()
				results[id] = &docInfo{
					err: fmt.Errorf("无法获取文档信息: %v", err),
				}
				mu.Unlock()
				return
			}

			// Get chunk count
			_, total, err := t.chunkService.GetRepository().
				ListPagedChunksByKnowledgeID(ctx, tenantID, id, &types.Pagination{
					Page:     1,
					PageSize: 1000,
				}, []types.ChunkType{"text"}, "", "", "", "", "")
			if err != nil {
				mu.Lock()
				results[id] = &docInfo{
					err: fmt.Errorf("无法获取文档信息: %v", err),
				}
				mu.Unlock()
				return
			}
			chunkCount := int(total)

			mu.Lock()
			results[id] = &docInfo{
				knowledge:  knowledge,
				chunkCount: chunkCount,
			}
			mu.Unlock()
		}(knowledgeID)
	}

	wg.Wait()

	// Collect successful results and errors
	successDocs := make([]*docInfo, 0)
	var errors []string

	for _, knowledgeID := range knowledgeIDs {
		result := results[knowledgeID]
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", knowledgeID, result.err))
		} else if result.knowledge != nil {
			successDocs = append(successDocs, result)
		}
	}

	if len(successDocs) == 0 {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("无法获取任何文档信息。错误: %v", errors),
		}, fmt.Errorf("all document retrievals failed")
	}

	// Format output
	output := "=== 文档信息 ===\n\n"
	output += fmt.Sprintf("成功获取 %d / %d 个文档信息\n\n", len(successDocs), len(knowledgeIDs))

	if len(errors) > 0 {
		output += "=== 部分失败 ===\n"
		for _, errMsg := range errors {
			output += fmt.Sprintf("  - %s\n", errMsg)
		}
		output += "\n"
	}

	formattedDocs := make([]map[string]interface{}, 0, len(successDocs))
	for i, doc := range successDocs {
		k := doc.knowledge

		output += fmt.Sprintf("【文档 #%d】\n", i+1)
		output += fmt.Sprintf("  ID:       %s\n", k.ID)
		output += fmt.Sprintf("  标题:     %s\n", k.Title)

		if k.Description != "" {
			output += fmt.Sprintf("  描述:     %s\n", k.Description)
		}

		output += fmt.Sprintf("  来源:     %s\n", formatSource(k.Type, k.Source))

		if k.FileName != "" {
			output += fmt.Sprintf("  文件名:   %s\n", k.FileName)
			output += fmt.Sprintf("  文件类型: %s\n", k.FileType)
			output += fmt.Sprintf("  文件大小: %s\n", formatFileSize(k.FileSize))
		}

		output += fmt.Sprintf("  处理状态: %s\n", formatParseStatus(k.ParseStatus))
		output += fmt.Sprintf("  分块数量: %d 个\n", doc.chunkCount)

		if k.Metadata != nil {
			if metadata, err := k.Metadata.Map(); err == nil && len(metadata) > 0 {
				output += "  元数据:\n"
				for key, value := range metadata {
					output += fmt.Sprintf("    - %s: %v\n", key, value)
				}
			}
		}

		output += "\n"

		formattedDocs = append(formattedDocs, map[string]interface{}{
			"knowledge_id": k.ID,
			"title":        k.Title,
			"description":  k.Description,
			"type":         k.Type,
			"source":       k.Source,
			"file_name":    k.FileName,
			"file_type":    k.FileType,
			"file_size":    k.FileSize,
			"parse_status": k.ParseStatus,
			"chunk_count":  doc.chunkCount,
			"metadata":     k.GetMetadata(),
		})
	}

	// Extract first document title for summary
	var firstTitle string
	if len(successDocs) > 0 && successDocs[0].knowledge != nil {
		firstTitle = successDocs[0].knowledge.Title
	}

	return &types.ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]interface{}{
			"documents":    formattedDocs,
			"total_docs":   len(successDocs),
			"requested":    len(knowledgeIDs),
			"errors":       errors,
			"display_type": "document_info",
			"title":        firstTitle, // For frontend summary display
		},
	}, nil
}

func formatSource(knowledgeType, source string) string {
	switch knowledgeType {
	case "file":
		return "文件上传"
	case "url":
		return fmt.Sprintf("URL: %s", source)
	case "passage":
		return "文本输入"
	default:
		return knowledgeType
	}
}

func formatFileSize(size int64) string {
	if size == 0 {
		return "未知"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatParseStatus(status string) string {
	switch status {
	case "pending":
		return "⏳ 待处理"
	case "processing":
		return "🔄 处理中"
	case "completed", "success":
		return "✅ 已完成"
	case "failed":
		return "❌ 失败"
	default:
		return status
	}
}
