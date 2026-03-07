<template>
  <div class="retrieval-test-settings">
    <div class="section-header">
      <h2>{{ $t('retrievalTest.title') }}</h2>
      <p class="section-description">{{ $t('retrievalTest.description') }}</p>
    </div>

    <div class="settings-group">
      <!-- 知识库选择 -->
      <div class="setting-row">
        <div class="setting-info">
          <label>{{ $t('retrievalTest.selectKnowledgeBase') }}</label>
          <p class="desc">{{ $t('retrievalTest.selectKnowledgeBaseDesc') }}</p>
        </div>
        <div class="setting-control">
          <t-select
            v-model="selectedKbId"
            :placeholder="$t('retrievalTest.selectKnowledgeBasePlaceholder')"
            :loading="loadingKbs"
            style="width: 280px;"
            filterable
          >
            <t-option
              v-for="kb in knowledgeBases"
              :key="kb.id"
              :value="kb.id"
              :label="kb.name"
            >
              {{ kb.name }}
            </t-option>
          </t-select>
        </div>
      </div>

      <!-- 查询文本 -->
      <div class="setting-row query-row">
        <div class="setting-info full-width">
          <label>{{ $t('retrievalTest.queryText') }}</label>
          <t-textarea
            v-model="queryText"
            :placeholder="$t('retrievalTest.queryPlaceholder')"
            :autosize="{ minRows: 2, maxRows: 4 }"
            style="margin-top: 8px;"
          />
        </div>
      </div>

      <!-- Embedding模型选择 -->
      <div class="setting-row">
        <div class="setting-info">
          <label>{{ $t('retrievalTest.embeddingModel') }}</label>
          <p class="desc">{{ $t('retrievalTest.embeddingModelDesc') }}</p>
        </div>
        <div class="setting-control">
          <t-select
            v-model="selectedEmbeddingModelId"
            :placeholder="$t('retrievalTest.useDefaultModel')"
            :loading="loadingModels"
            style="width: 280px;"
            clearable
            filterable
          >
            <t-option
              v-for="model in embeddingModels"
              :key="model.id"
              :value="model.id"
              :label="model.name"
            >
              {{ model.name }}
            </t-option>
          </t-select>
        </div>
      </div>

      <!-- Rerank模型选择 -->
      <div class="setting-row">
        <div class="setting-info">
          <label>{{ $t('retrievalTest.rerankModel') }}</label>
          <p class="desc">{{ $t('retrievalTest.rerankModelDesc') }}</p>
        </div>
        <div class="setting-control">
          <t-select
            v-model="selectedRerankModelId"
            :placeholder="$t('retrievalTest.noRerank')"
            :loading="loadingModels"
            style="width: 280px;"
            clearable
            filterable
          >
            <t-option
              v-for="model in rerankModels"
              :key="model.id"
              :value="model.id"
              :label="model.name"
            >
              {{ model.name }}
            </t-option>
          </t-select>
        </div>
      </div>

      <!-- 参数配置（可折叠） -->
      <div class="setting-row config-row">
        <div class="config-header" @click="showConfig = !showConfig">
          <div class="config-title">
            <t-icon :name="showConfig ? 'chevron-down' : 'chevron-right'" />
            <span>{{ $t('retrievalTest.paramConfig') }}</span>
          </div>
        </div>
        
        <Transition name="config">
          <div v-if="showConfig" class="config-panel">
            <!-- 向量阈值 -->
            <div class="config-item">
              <label>{{ $t('retrievalTest.vectorThreshold') }}</label>
              <div class="slider-row">
                <t-slider
                  v-model="vectorThreshold"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  style="flex: 1;"
                />
                <t-input-number
                  v-model="vectorThreshold"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  :decimal-places="2"
                  style="width: 100px; margin-left: 16px;"
                />
              </div>
            </div>

            <!-- 关键词阈值 -->
            <div class="config-item">
              <label>{{ $t('retrievalTest.keywordThreshold') }}</label>
              <div class="slider-row">
                <t-slider
                  v-model="keywordThreshold"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  style="flex: 1;"
                />
                <t-input-number
                  v-model="keywordThreshold"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  :decimal-places="2"
                  style="width: 100px; margin-left: 16px;"
                />
              </div>
            </div>

            <!-- 返回数量 -->
            <div class="config-item">
              <label>{{ $t('retrievalTest.matchCount') }}</label>
              <t-input-number
                v-model="matchCount"
                :min="1"
                :max="100"
                style="width: 140px;"
              />
            </div>

            <!-- Rerank阈值（仅在选择了Rerank模型时显示） -->
            <template v-if="selectedRerankModelId">
              <div class="config-item">
                <label>{{ $t('retrievalTest.rerankThreshold') }}</label>
                <div class="slider-row">
                  <t-slider
                    v-model="rerankThreshold"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    style="flex: 1;"
                  />
                  <t-input-number
                    v-model="rerankThreshold"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :decimal-places="2"
                    style="width: 100px; margin-left: 16px;"
                  />
                </div>
              </div>

              <!-- Rerank TopK -->
              <div class="config-item">
                <label>{{ $t('retrievalTest.rerankTopK') }}</label>
                <t-input-number
                  v-model="rerankTopK"
                  :min="1"
                  :max="50"
                  style="width: 140px;"
                />
              </div>
            </template>

            <!-- 禁用选项 -->
            <div class="config-item checkbox-group">
              <t-checkbox v-model="disableVectorMatch">
                {{ $t('retrievalTest.disableVectorMatch') }}
              </t-checkbox>
              <t-checkbox v-model="disableKeywordsMatch">
                {{ $t('retrievalTest.disableKeywordsMatch') }}
              </t-checkbox>
            </div>
          </div>
        </Transition>
      </div>

      <!-- 搜索按钮 -->
      <div class="setting-row action-row">
        <t-button
          theme="primary"
          :loading="searching"
          :disabled="!selectedKbId || !queryText.trim()"
          @click="handleSearch"
        >
          <template #icon>
            <t-icon name="search" />
          </template>
          {{ searching ? $t('retrievalTest.searching') : $t('retrievalTest.search') }}
        </t-button>
      </div>
    </div>

    <!-- 搜索结果 -->
    <div v-if="searchResults.length > 0 || hasSearched" class="results-section">
      <div class="results-header">
        <h3>{{ $t('retrievalTest.results') }}</h3>
        <t-tag theme="primary" variant="light" size="small">
          {{ searchResults.length }}
        </t-tag>
      </div>

      <div v-if="searchResults.length === 0" class="no-results">
        <t-icon name="search" size="48px" />
        <p>{{ $t('retrievalTest.noResults') }}</p>
      </div>

      <div v-else class="results-list">
        <div
          v-for="(result, index) in searchResults"
          :key="result.id"
          class="result-card"
        >
          <div class="result-header">
            <span class="result-index">#{{ index + 1 }}</span>
            <div class="result-meta">
              <t-tag theme="success" variant="light" size="small">
                {{ $t('retrievalTest.score') }}: {{ result.score?.toFixed(4) || 'N/A' }}
              </t-tag>
              <t-tag 
                :theme="result.match_type === 'embedding' ? 'primary' : 'warning'" 
                variant="light" 
                size="small"
              >
                {{ $t('retrievalTest.matchType') }}: {{ result.match_type || 'unknown' }}
              </t-tag>
            </div>
          </div>
          <div class="result-content">
            <p>{{ result.content }}</p>
          </div>
          <div class="result-footer">
            <span class="source-info">
              <t-icon name="file" />
              {{ result.knowledge_filename || result.knowledge_title || 'Unknown' }}
            </span>
            <span class="chunk-info">
              {{ $t('retrievalTest.chunkIndex') }}: {{ result.chunk_index ?? result.seq ?? 'N/A' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import { useI18n } from 'vue-i18n'
import { listKnowledgeBases, hybridSearch } from '@/api/knowledge-base'
import { listModels, type ModelConfig } from '@/api/model'

const { t } = useI18n()

// 知识库列表
const knowledgeBases = ref<any[]>([])
const loadingKbs = ref(false)
const selectedKbId = ref('')

// 模型列表
const embeddingModels = ref<ModelConfig[]>([])
const rerankModels = ref<ModelConfig[]>([])
const loadingModels = ref(false)
const selectedEmbeddingModelId = ref('')
const selectedRerankModelId = ref('')

// 查询参数
const queryText = ref('')
const vectorThreshold = ref(0.6)
const keywordThreshold = ref(0.5)
const matchCount = ref(10)
const rerankThreshold = ref(0.3)
const rerankTopK = ref(5)
const disableVectorMatch = ref(false)
const disableKeywordsMatch = ref(false)

// UI状态
const showConfig = ref(true)
const searching = ref(false)
const hasSearched = ref(false)
const searchResults = ref<any[]>([])

// 加载知识库列表
const loadKnowledgeBases = async () => {
  loadingKbs.value = true
  try {
    const response: any = await listKnowledgeBases()
    if (response.success && response.data) {
      knowledgeBases.value = response.data
    }
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('common.error'))
  } finally {
    loadingKbs.value = false
  }
}

// 加载模型列表
const loadModels = async () => {
  loadingModels.value = true
  try {
    const [embeddingList, rerankList] = await Promise.all([
      listModels('Embedding'),
      listModels('Rerank')
    ])
    embeddingModels.value = embeddingList || []
    rerankModels.value = rerankList || []
  } catch (error: any) {
    console.error('Failed to load models:', error)
  } finally {
    loadingModels.value = false
  }
}

// 执行搜索
const handleSearch = async () => {
  if (!selectedKbId.value || !queryText.value.trim()) {
    return
  }

  searching.value = true
  hasSearched.value = true
  searchResults.value = []

  try {
    const response: any = await hybridSearch(selectedKbId.value, {
      query_text: queryText.value.trim(),
      vector_threshold: vectorThreshold.value,
      keyword_threshold: keywordThreshold.value,
      match_count: matchCount.value,
      disable_vector_match: disableVectorMatch.value,
      disable_keywords_match: disableKeywordsMatch.value,
      embedding_model_id: selectedEmbeddingModelId.value || undefined,
      rerank_model_id: selectedRerankModelId.value || undefined,
      rerank_threshold: selectedRerankModelId.value ? rerankThreshold.value : undefined,
      rerank_top_k: selectedRerankModelId.value ? rerankTopK.value : undefined,
    })

    if (response.success && response.data) {
      searchResults.value = response.data
    } else if (response.data) {
      // Some APIs return data directly without success flag
      searchResults.value = Array.isArray(response.data) ? response.data : []
    }
  } catch (error: any) {
    MessagePlugin.error(error?.message || t('retrievalTest.searchError'))
  } finally {
    searching.value = false
  }
}

onMounted(() => {
  loadKnowledgeBases()
  loadModels()
})
</script>

<style lang="less" scoped>
.retrieval-test-settings {
  width: 100%;
}

.section-header {
  margin-bottom: 32px;

  h2 {
    font-size: 20px;
    font-weight: 600;
    color: #333333;
    margin: 0 0 8px 0;
  }

  .section-description {
    font-size: 14px;
    color: #666666;
    margin: 0;
    line-height: 1.5;
  }
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.setting-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 20px 0;
  border-bottom: 1px solid #e5e7eb;

  &:last-child {
    border-bottom: none;
  }
}

.setting-info {
  flex: 1;
  max-width: 65%;
  padding-right: 24px;

  &.full-width {
    max-width: 100%;
    padding-right: 0;
  }

  label {
    font-size: 15px;
    font-weight: 500;
    color: #333333;
    display: block;
    margin-bottom: 4px;
  }

  .desc {
    font-size: 13px;
    color: #666666;
    margin: 0;
    line-height: 1.5;
  }
}

.setting-control {
  flex-shrink: 0;
  min-width: 280px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.query-row {
  flex-direction: column;

  .setting-info {
    width: 100%;
  }
}

.config-row {
  flex-direction: column;
  padding: 16px 0;
}

.config-header {
  cursor: pointer;
  user-select: none;

  .config-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 15px;
    font-weight: 500;
    color: #333333;

    :deep(.t-icon) {
      font-size: 16px;
      color: #666666;
    }
  }
}

.config-panel {
  margin-top: 16px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.config-item {
  margin-bottom: 16px;

  &:last-child {
    margin-bottom: 0;
  }

  label {
    font-size: 14px;
    font-weight: 500;
    color: #333333;
    display: block;
    margin-bottom: 8px;
  }
}

.slider-row {
  display: flex;
  align-items: center;
  width: 100%;

  :deep(.t-slider) {
    flex: 1;
    min-width: 150px;
  }
}

.checkbox-group {
  display: flex;
  gap: 24px;
}

.action-row {
  justify-content: flex-start;
  border-bottom: none;
}

// 参数面板动画
.config-enter-active,
.config-leave-active {
  transition: all 0.2s ease;
}

.config-enter-from,
.config-leave-to {
  opacity: 0;
  max-height: 0;
  margin-top: 0;
}

.config-enter-to,
.config-leave-from {
  opacity: 1;
  max-height: 500px;
}

// 结果区域
.results-section {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
}

.results-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;

  h3 {
    font-size: 16px;
    font-weight: 600;
    color: #333333;
    margin: 0;
  }
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  color: #999999;

  :deep(.t-icon) {
    margin-bottom: 16px;
    opacity: 0.5;
  }

  p {
    margin: 0;
    font-size: 14px;
  }
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.result-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  transition: all 0.2s ease;

  &:hover {
    border-color: #d0d0d0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.result-index {
  font-size: 14px;
  font-weight: 600;
  color: #D97706;
}

.result-meta {
  display: flex;
  gap: 8px;
}

.result-content {
  margin-bottom: 12px;

  p {
    margin: 0;
    font-size: 14px;
    color: #333333;
    line-height: 1.6;
    word-break: break-word;
    white-space: pre-wrap;
  }
}

.result-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  color: #666666;
}

.source-info {
  display: flex;
  align-items: center;
  gap: 4px;

  :deep(.t-icon) {
    font-size: 14px;
  }
}

.chunk-info {
  color: #999999;
}
</style>
