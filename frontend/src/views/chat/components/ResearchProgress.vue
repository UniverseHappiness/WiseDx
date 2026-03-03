<template>
  <div class="research-progress" v-if="visible && totalStages > 0">
    <div class="progress-header">
      <span class="progress-icon">🔬</span>
      <span class="progress-title">{{ $t('agent.researchProgress') }}</span>
      <span class="progress-stage">{{ currentStage }}/{{ totalStages }}</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
    </div>
    <div class="progress-current">
      <t-icon name="location" size="14px" />
      <span>{{ $t('agent.currentStage') }}: {{ currentStageName }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// Deep research stage mapping
const STAGE_NAMES: Record<string, string> = {
  'step1': '研究规划',
  'step2': '资料检索',
  'step3': '深度分析',
  'step4': '交叉验证',
  'step5': '报告生成',
};

interface PlanStep {
  id: string;
  description: string;
  status: string;
}

const props = defineProps<{
  steps: PlanStep[];
  visible?: boolean;
  agentId?: string;
}>();

// Calculate current stage based on step status
const currentStage = computed(() => {
  if (!props.steps || props.steps.length === 0) return 0;
  
  // Find first in_progress or first pending step
  const inProgressIndex = props.steps.findIndex(s => s.status === 'in_progress');
  if (inProgressIndex >= 0) return inProgressIndex + 1;
  
  // If no in_progress, check for completed steps
  const completedCount = props.steps.filter(s => s.status === 'completed').length;
  if (completedCount === props.steps.length) return props.steps.length;
  
  return completedCount + 1;
});

const totalStages = computed(() => {
  return props.steps?.length || 0;
});

const progressPercent = computed(() => {
  if (totalStages.value === 0) return 0;
  return Math.round((currentStage.value / totalStages.value) * 100);
});

const currentStageName = computed(() => {
  if (!props.steps || props.steps.length === 0) return '';
  
  // Find current step
  const currentStep = props.steps.find(s => s.status === 'in_progress');
  if (currentStep) {
    return STAGE_NAMES[currentStep.id] || currentStep.description;
  }
  
  // If all completed
  const allCompleted = props.steps.every(s => s.status === 'completed');
  if (allCompleted) {
    return '已完成';
  }
  
  // Find first pending
  const pendingStep = props.steps.find(s => s.status === 'pending');
  if (pendingStep) {
    return STAGE_NAMES[pendingStep.id] || pendingStep.description;
  }
  
  return '';
});
</script>

<style lang="less" scoped>
.research-progress {
  margin: 12px 0;
  padding: 12px 16px;
  background: linear-gradient(135deg, #f9f0ff 0%, #efdbff 100%);
  border: 1px solid #d3adf7;
  border-radius: 8px;
  
  .progress-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    
    .progress-icon {
      font-size: 16px;
    }
    
    .progress-title {
      font-size: 14px;
      font-weight: 500;
      color: #531dab;
      flex: 1;
    }
    
    .progress-stage {
      font-size: 13px;
      color: #722ed1;
      font-weight: 600;
      background: rgba(114, 46, 209, 0.1);
      padding: 2px 8px;
      border-radius: 10px;
    }
  }
  
  .progress-bar {
    height: 8px;
    background: rgba(114, 46, 209, 0.2);
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 8px;
    
    .progress-fill {
      height: 100%;
      background: linear-gradient(90deg, #722ed1 0%, #9254de 100%);
      border-radius: 4px;
      transition: width 0.3s ease;
    }
  }
  
  .progress-current {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: #722ed1;
    
    :deep(.t-icon) {
      color: #722ed1;
    }
  }
}

// Dark mode support
:global(.dark) {
  .research-progress {
    background: linear-gradient(135deg, #120338 0%, #1a0a4e 100%);
    border-color: #9254de;
    
    .progress-header {
      .progress-title {
        color: #b37feb;
      }
      
      .progress-stage {
        color: #b37feb;
        background: rgba(146, 84, 222, 0.2);
      }
    }
    
    .progress-bar {
      background: rgba(146, 84, 222, 0.3);
      
      .progress-fill {
        background: linear-gradient(90deg, #9254de 0%, #b37feb 100%);
      }
    }
    
    .progress-current {
      color: #b37feb;
      
      :deep(.t-icon) {
        color: #b37feb;
      }
    }
  }
}
</style>
