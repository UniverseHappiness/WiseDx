<template>
  <div class="consultation-progress" v-if="visible && totalStages > 0">
    <div class="progress-header">
      <span class="progress-icon">📊</span>
      <span class="progress-title">{{ $t('agent.consultationProgress') }}</span>
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

// Medical consultation stage mapping
const STAGE_NAMES: Record<string, string> = {
  'step1': '问候与身份确认',
  'step2': '主诉采集',
  'step3': '现病史采集',
  'step4': '既往史采集',
  'step5': '过敏史采集',
  'step6': '信息总结',
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

// Only show for medical consultant agent
const isMedicalAgent = computed(() => {
  return props.agentId === 'builtin-medical-consultant';
});

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
.consultation-progress {
  margin: 12px 0;
  padding: 12px 16px;
  background: linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%);
  border: 1px solid #b7eb8f;
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
      color: #389e0d;
      flex: 1;
    }
    
    .progress-stage {
      font-size: 13px;
      color: #52c41a;
      font-weight: 600;
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
    font-size: 13px;
    color: #52c41a;
    
    :deep(.t-icon) {
      color: #52c41a;
    }
  }
}

// Dark mode support
:global(.dark) {
  .consultation-progress {
    background: linear-gradient(135deg, #162312 0%, #1d3712 100%);
    border-color: #49aa19;
    
    .progress-header {
      .progress-title {
        color: #95de64;
      }
      
      .progress-stage {
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
  }
}
</style>
