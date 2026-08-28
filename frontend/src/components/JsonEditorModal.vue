<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  payload: {
    type: Object,
    required: true
  },
  isLoading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'apply', 'execute'])

const jsonText = ref('')
const parseError = ref('')

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    jsonText.value = JSON.stringify(props.payload, null, 2)
    parseError.value = ''
  }
}, { immediate: true })

function formatJson() {
  try {
    const parsed = JSON.parse(jsonText.value)
    jsonText.value = JSON.stringify(parsed, null, 2)
    parseError.value = ''
  } catch (err) {
    parseError.value = 'Invalid JSON: ' + err.message
  }
}

function handleApply() {
  try {
    const parsed = JSON.parse(jsonText.value)
    emit('apply', parsed)
    emit('close')
  } catch (err) {
    parseError.value = 'Invalid JSON: ' + err.message
  }
}

function handleExecute() {
  try {
    const parsed = JSON.parse(jsonText.value)
    emit('apply', parsed)
    emit('execute')
    emit('close')
  } catch (err) {
    parseError.value = 'Invalid JSON: ' + err.message
  }
}
</script>

<template>
  <div v-if="isOpen" class="modal-backdrop" @click.self="emit('close')">
    <div class="modal-content modal-large">
      <div class="modal-header">
        <div class="modal-title-wrap">
          <span class="icon">data_object</span>
          <h3>Raw JSON Request Payload</h3>
        </div>
        <button class="btn btn-ghost btn-sm close-modal-btn" @click="emit('close')">×</button>
      </div>

      <div class="modal-body">
        <p class="modal-desc">
          Edit the exact JSON payload sent to the backend <code>POST /generate</code> endpoint.
        </p>

        <div v-if="parseError" class="json-error-alert">
          <span class="icon">error</span>
          <span>{{ parseError }}</span>
        </div>

        <textarea 
          class="form-textarea code-font json-textarea" 
          v-model="jsonText"
          rows="18"
          spellcheck="false"
          id="textarea-json-payload"
        ></textarea>
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary btn-sm" @click="formatJson">
          <span class="icon">format_align_left</span> Prettify JSON
        </button>
        <button type="button" class="btn btn-secondary btn-sm" @click="emit('close')">
          Cancel
        </button>
        <button type="button" class="btn btn-secondary btn-sm" @click="handleApply">
          Sync to Form
        </button>
        <button 
          type="button" 
          class="btn btn-primary btn-sm" 
          :disabled="isLoading" 
          @click="handleExecute"
          id="btn-modal-execute"
        >
          <span class="icon">send</span> Send Request
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-large {
  max-width: 800px;
}

.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.modal-title-wrap .icon {
  color: #3b82f6;
  font-size: 22px;
}
.modal-title-wrap h3 {
  font-size: 16px;
  font-weight: 600;
}

.close-modal-btn {
  font-size: 20px;
  line-height: 1;
  padding: 4px 8px;
}

.modal-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 12px;
}
.modal-desc code {
  font-family: var(--font-mono);
  color: #93c5fd;
  background-color: var(--bg-app);
  padding: 2px 4px;
  border-radius: 4px;
}

.json-error-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: rgba(239, 68, 68, 0.15);
  border: 1px solid var(--color-error);
  color: #fca5a5;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  margin-bottom: 12px;
}

.json-textarea {
  width: 100%;
  font-size: 13px;
  line-height: 1.5;
  background-color: #0b0f19;
  border-color: var(--border-color);
  color: #38bdf8;
  resize: vertical;
}
</style>
