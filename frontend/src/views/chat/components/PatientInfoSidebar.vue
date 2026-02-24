<template>
  <div class="patient-info-sidebar" v-if="visible && (hasData || hasProgress)">
    <div class="sidebar-header">
      <div class="sidebar-title">
        <t-icon name="chart-bubble" />
        <span>{{ $t('agent.consultationInfo') }}</span>
      </div>
      <t-button variant="text" size="small" @click="handleClose">
        <t-icon name="close" />
      </t-button>
    </div>
    
    <div class="sidebar-content">
      <!-- 问诊进度 -->
      <div class="progress-section" v-if="hasProgress">
        <div class="section-title-main">
          <t-icon name="chart-line" />
          <span>{{ $t('agent.consultationProgress') }}</span>
        </div>
        <div class="progress-card">
          <div class="progress-header">
            <span class="progress-stage">{{ currentStage }}/{{ totalStages }}</span>
            <span class="progress-percent">{{ progressPercent }}%</span>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
          </div>
          <div class="progress-current" v-if="currentStageName">
            <t-icon name="play-circle" size="14px" />
            <span>{{ currentStageName }}</span>
          </div>
          <!-- 步骤列表 -->
          <div class="steps-list" v-if="planSteps && planSteps.length > 0">
            <div 
              v-for="(step, index) in planSteps" 
              :key="step.id"
              class="step-item"
              :class="{ 
                'step-completed': step.status === 'completed',
                'step-active': step.status === 'in_progress',
                'step-pending': step.status === 'pending'
              }"
            >
              <div class="step-icon">
                <t-icon v-if="step.status === 'completed'" name="check-circle-filled" />
                <t-icon v-else-if="step.status === 'in_progress'" name="loading" class="rotating" />
                <span v-else class="step-number">{{ index + 1 }}</span>
              </div>
              <div class="step-content">
                <div class="step-name">{{ getStepName(step.id) }}</div>
                <div class="step-desc">{{ step.description }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 已收集信息 -->
      <div class="collected-section" v-if="hasData">
        <div class="section-title-main">
          <t-icon name="usergroup" />
          <span>{{ $t('agent.collectedInfo') }}</span>
        </div>
        
        <!-- 基本信息 -->
        <div class="info-section" v-if="basicInfo.length > 0">
          <div class="section-title">{{ $t('agent.basicInfo') }}</div>
          <div class="info-list">
            <div class="info-item" v-for="item in basicInfo" :key="item.key">
              <span class="info-label">{{ item.key }}</span>
              <span class="info-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
        
        <!-- 主诉信息 -->
        <div class="info-section" v-if="chiefComplaint.length > 0">
          <div class="section-title">{{ $t('agent.chiefComplaint') }}</div>
          <div class="info-list">
            <div class="info-item" v-for="item in chiefComplaint" :key="item.key">
              <span class="info-label">{{ item.key }}</span>
              <span class="info-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
        
        <!-- 现病史 -->
        <div class="info-section" v-if="presentIllness.length > 0">
          <div class="section-title">{{ $t('agent.presentIllness') }}</div>
          <div class="info-list">
            <div class="info-item" v-for="item in presentIllness" :key="item.key">
              <span class="info-label">{{ item.key }}</span>
              <span class="info-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
        
        <!-- 既往史 -->
        <div class="info-section" v-if="pastHistory.length > 0">
          <div class="section-title">{{ $t('agent.pastHistory') }}</div>
          <div class="info-list">
            <div class="info-item" v-for="item in pastHistory" :key="item.key">
              <span class="info-label">{{ item.key }}</span>
              <span class="info-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
        
        <!-- 过敏史 -->
        <div class="info-section" v-if="allergyHistory.length > 0">
          <div class="section-title">{{ $t('agent.allergyHistory') }}</div>
          <div class="info-list">
            <div class="info-item" v-for="item in allergyHistory" :key="item.key">
              <span class="info-label">{{ item.key }}</span>
              <span class="info-value">{{ item.value }}</span>
            </div>
          </div>
        </div>

        <!-- 空状态提示 -->
        <div class="empty-state" v-if="!hasData">
          <t-icon name="info-circle" size="32px" />
          <p>{{ $t('agent.noCollectedData') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface InfoItem {
  key: string;
  value: string;
}

interface PlanStep {
  id: string;
  description: string;
  status: string;
}

const props = defineProps<{
  collectedData: Record<string, string>;
  visible: boolean;
  planSteps?: PlanStep[];
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

// Field category mapping
const basicFields = ['姓名', '年龄', '性别', 'name', 'age', 'gender'];
const chiefFields = ['主诉', '持续时间', '严重程度', 'chief_complaint', 'duration', 'severity'];
const presentFields = ['发病时间', '诱因', '症状特点', '伴随症状', '已有治疗', 'onset_time', 'cause', 'characteristics', 'accompanying_symptoms', 'treatment_taken'];
const pastFields = ['既往疾病', '手术史', '住院史', '长期用药', 'diseases', 'surgeries', 'hospitalizations', 'medications'];
const allergyFields = ['药物过敏', '食物过敏', '其他过敏', 'drug_allergy', 'food_allergy', 'other_allergy'];

// Medical consultation stage mapping
const STAGE_NAMES: Record<string, string> = {
  'step1': '问候与身份确认',
  'step2': '主诉采集',
  'step3': '现病史采集',
  'step4': '既往史采集',
  'step5': '过敏史采集',
  'step6': '信息总结',
};

const hasData = computed(() => {
  return props.collectedData && Object.keys(props.collectedData).length > 0;
});

const hasProgress = computed(() => {
  return props.planSteps && props.planSteps.length > 0;
});

const categorizeData = (fields: string[]): InfoItem[] => {
  if (!props.collectedData) return [];
  
  return Object.entries(props.collectedData)
    .filter(([key]) => fields.some(f => key.toLowerCase().includes(f.toLowerCase()) || f.toLowerCase().includes(key.toLowerCase())))
    .map(([key, value]) => ({ key, value }));
};

const basicInfo = computed(() => categorizeData(basicFields));
const chiefComplaint = computed(() => categorizeData(chiefFields));
const presentIllness = computed(() => categorizeData(presentFields));
const pastHistory = computed(() => categorizeData(pastFields));
const allergyHistory = computed(() => categorizeData(allergyFields));

// Progress calculations
const currentStage = computed(() => {
  if (!props.planSteps || props.planSteps.length === 0) return 0;
  
  const inProgressIndex = props.planSteps.findIndex(s => s.status === 'in_progress');
  if (inProgressIndex >= 0) return inProgressIndex + 1;
  
  const completedCount = props.planSteps.filter(s => s.status === 'completed').length;
  if (completedCount === props.planSteps.length) return props.planSteps.length;
  
  return completedCount + 1;
});

const totalStages = computed(() => {
  return props.planSteps?.length || 0;
});

const progressPercent = computed(() => {
  if (totalStages.value === 0) return 0;
  return Math.round((currentStage.value / totalStages.value) * 100);
});

const currentStageName = computed(() => {
  if (!props.planSteps || props.planSteps.length === 0) return '';
  
  const currentStep = props.planSteps.find(s => s.status === 'in_progress');
  if (currentStep) {
    return STAGE_NAMES[currentStep.id] || currentStep.description;
  }
  
  const allCompleted = props.planSteps.every(s => s.status === 'completed');
  if (allCompleted) {
    return t('agent.completed');
  }
  
  const pendingStep = props.planSteps.find(s => s.status === 'pending');
  if (pendingStep) {
    return STAGE_NAMES[pendingStep.id] || pendingStep.description;
  }
  
  return '';
});

const getStepName = (stepId: string): string => {
  return STAGE_NAMES[stepId] || stepId;
};

const handleClose = () => {
  emit('close');
};
</script>

<style lang="less" scoped>
.patient-info-sidebar {
  width: 320px;
  height: 100%;
  background: var(--td-bg-color-container);
  border-left: 1px solid var(--td-border-level-1-color);
  display: flex;
  flex-direction: column;
  
  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--td-border-level-1-color);
    background: linear-gradient(135deg, rgba(0, 82, 217, 0.05) 0%, rgba(0, 82, 217, 0.02) 100%);
    
    .sidebar-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      font-weight: 600;
      color: var(--td-text-color-primary);
      
      :deep(.t-icon) {
        font-size: 18px;
        color: var(--td-brand-color);
      }
    }
  }
  
  .sidebar-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    
    &::-webkit-scrollbar {
      width: 6px;
    }
    
    &::-webkit-scrollbar-thumb {
      background: var(--td-scrollbar-color);
      border-radius: 3px;
    }

    .section-title-main {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      font-weight: 600;
      color: var(--td-text-color-primary);
      margin-bottom: 12px;
      padding-bottom: 8px;
      border-bottom: 2px solid var(--td-brand-color);
      
      :deep(.t-icon) {
        font-size: 16px;
        color: var(--td-brand-color);
      }
    }

    .progress-section {
      margin-bottom: 24px;
      
      .progress-card {
        background: linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%);
        border: 1px solid #b7eb8f;
        border-radius: 8px;
        padding: 12px;
        
        .progress-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;
          
          .progress-stage,
          .progress-percent {
            font-size: 13px;
            font-weight: 600;
            color: #52c41a;
            background: rgba(82, 196, 26, 0.1);
            padding: 2px 8px;
            border-radius: 10px;
          }
        }
        
        .progress-bar {
          height: 8px;
          background: rgba(82, 196, 26, 0.2);
          border-radius: 4px;
          overflow: hidden;
          margin-bottom: 8px;
          
          .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #52c41a 0%, #73d13d 100%);
            border-radius: 4px;
            transition: width 0.3s ease;
          }
        }
        
        .progress-current {
          display: flex;
          align-items: center;
          gap: 4px;
          font-size: 12px;
          color: #389e0d;
          font-weight: 500;
          margin-bottom: 12px;
          
          :deep(.t-icon) {
            color: #52c41a;
          }
        }

        .steps-list {
          .step-item {
            display: flex;
            gap: 10px;
            padding: 8px 0;
            border-bottom: 1px solid rgba(0, 0, 0, 0.06);
            
            &:last-child {
              border-bottom: none;
            }
            
            .step-icon {
              flex-shrink: 0;
              width: 24px;
              height: 24px;
              display: flex;
              align-items: center;
              justify-content: center;
              border-radius: 50%;
              font-size: 14px;
              background: rgba(0, 0, 0, 0.06);
              color: var(--td-text-color-placeholder);
              
              .step-number {
                font-size: 12px;
                font-weight: 600;
              }

              .rotating {
                animation: rotate 1s linear infinite;
              }
            }
            
            .step-content {
              flex: 1;
              min-width: 0;
              
              .step-name {
                font-size: 13px;
                font-weight: 500;
                color: var(--td-text-color-primary);
                margin-bottom: 2px;
              }
              
              .step-desc {
                font-size: 11px;
                color: var(--td-text-color-placeholder);
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
              }
            }
            
            &.step-completed {
              .step-icon {
                background: #52c41a;
                color: white;
              }
              
              .step-content {
                .step-name {
                  color: var(--td-text-color-secondary);
                  text-decoration: line-through;
                }
              }
            }
            
            &.step-active {
              .step-icon {
                background: var(--td-brand-color);
                color: white;
              }
              
              .step-content {
                .step-name {
                  color: var(--td-brand-color);
                  font-weight: 600;
                }
              }
            }
          }
        }
      }
    }

    .collected-section {
      .info-section {
        margin-bottom: 16px;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        .section-title {
          font-size: 12px;
          font-weight: 500;
          color: var(--td-text-color-secondary);
          margin-bottom: 8px;
          padding: 4px 8px;
          background: var(--td-bg-color-component);
          border-radius: 4px;
        }
        
        .info-list {
          .info-item {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            padding: 6px 8px;
            font-size: 12px;
            border-radius: 4px;
            transition: background 0.2s;
            
            &:hover {
              background: var(--td-bg-color-component-hover);
            }
            
            .info-label {
              color: var(--td-text-color-secondary);
              flex-shrink: 0;
              margin-right: 12px;
              font-weight: 500;
              
              &::after {
                content: ':';
                margin-left: 2px;
              }
            }
            
            .info-value {
              color: var(--td-text-color-primary);
              text-align: right;
              word-break: break-word;
              flex: 1;
            }
          }
        }
      }

      .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 32px 16px;
        color: var(--td-text-color-placeholder);
        
        :deep(.t-icon) {
          margin-bottom: 12px;
          opacity: 0.5;
        }
        
        p {
          font-size: 13px;
          margin: 0;
        }
      }
    }
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// Dark mode
::global(.dark) {
  .patient-info-sidebar {
    background: var(--td-bg-color-container);
    
    .sidebar-content {
      .progress-section {
        .progress-card {
          background: linear-gradient(135deg, #162312 0%, #1d3712 100%);
          border-color: #49aa19;
          
          .progress-header {
            .progress-stage,
            .progress-percent {
              color: #95de64;
              background: rgba(73, 170, 25, 0.2);
            }
          }
          
          .progress-bar {
            background: rgba(73, 170, 25, 0.3);
            
            .progress-fill {
              background: linear-gradient(90deg, #49aa19 0%, #6abe39 100%);
            }
          }
          
          .progress-current {
            color: #95de64;
            
            :deep(.t-icon) {
              color: #95de64;
            }
          }

          .steps-list .step-item {
            border-bottom-color: rgba(255, 255, 255, 0.1);
          }
        }
      }
    }
  }
}
</style>
