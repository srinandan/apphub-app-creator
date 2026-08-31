<script setup>
import { ref } from 'vue'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  serverUrl: {
    type: String,
    default: 'http://localhost:8080'
  }
})

const emit = defineEmits(['close', 'save', 'test-connection'])

const tempUrl = ref(props.serverUrl)
const isTesting = ref(false)
const testResult = ref(null)

function handleTest() {
  isTesting.value = true
  testResult.value = null

  const url = tempUrl.value.replace(/\/+$/, '')
  const start = performance.now()

  fetch(`${url}/`, { method: 'GET' })
    .then(res => {
      const elapsed = Math.round(performance.now() - start)
      if (res.ok) {
        testResult.value = { success: true, message: `Connected successfully (${elapsed}ms)` }
      } else {
        testResult.value = { success: false, message: `Server returned HTTP ${res.status}: ${res.statusText}` }
      }
    })
    .catch(err => {
      testResult.value = { success: false, message: `Connection failed: ${err.message}. Make sure './apphub-app-creator server' is running.` }
    })
    .finally(() => {
      isTesting.value = false
    })
}

function handleSave() {
  const sanitized = tempUrl.value.replace(/\/+$/, '')
  emit('save', sanitized)
  emit('close')
}
</script>

<template>
  <div v-if="isOpen" class="modal-backdrop" @click.self="emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <div class="modal-title-wrap">
          <span class="icon">settings</span>
          <h3>Backend Server Settings</h3>
        </div>
        <button class="btn btn-ghost btn-sm close-modal-btn" @click="emit('close')">×</button>
      </div>

      <div class="modal-body">
        <p class="modal-desc">
          Specify the HTTP URL of the <code>apphub-app-creator server</code> process.
        </p>

        <div class="form-group">
          <label class="form-label" for="input-server-url">Backend Server URL</label>
          <input 
            id="input-server-url"
            class="form-input code-font" 
            v-model="tempUrl" 
            placeholder="http://localhost:8080" 
          />
          <div class="form-hint">
            To start the server locally, run: <code>./apphub-app-creator server --port=8080</code>
          </div>
        </div>

        <!-- Connection Test Result -->
        <div v-if="testResult" class="test-result-box" :class="testResult.success ? 'success' : 'error'">
          <span class="icon">{{ testResult.success ? 'check_circle' : 'error' }}</span>
          <span>{{ testResult.message }}</span>
        </div>
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="isTesting" @click="handleTest">
          <span v-if="isTesting" class="spinner-sm"></span>
          <span v-else class="icon">wifi_tethering</span>
          <span>Test Connection</span>
        </button>
        <button type="button" class="btn btn-secondary btn-sm" @click="emit('close')">
          Cancel
        </button>
        <button type="button" class="btn btn-primary btn-sm" @click="handleSave">
          Save Settings
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}
.modal-title-wrap .icon {
  color: var(--color-brand);
}
.modal-title-wrap h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 16px;
}
.modal-desc code, .form-hint code {
  color: var(--color-brand-text);
  background: var(--bg-surface-elevated);
  padding: 2px 4px;
  border-radius: 4px;
  font-family: var(--font-mono);
  border: 1px solid var(--border-color);
  word-break: break-all;
  overflow-wrap: anywhere;
}

.test-result-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  margin-top: 12px;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.test-result-box.success {
  background-color: var(--color-success-bg);
  border: 1px solid var(--color-success);
  color: var(--color-success);
}

.test-result-box.error {
  background-color: var(--color-error-bg);
  border: 1px solid var(--color-error);
  color: var(--color-error);
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #ffffff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
