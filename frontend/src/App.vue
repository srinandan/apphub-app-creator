<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import HeaderNav from './components/HeaderNav.vue'
import SelectorForm from './components/SelectorForm.vue'
import ResultsViewer from './components/ResultsViewer.vue'
import JsonEditorModal from './components/JsonEditorModal.vue'
import ServerSettingsModal from './components/ServerSettingsModal.vue'

// Server Connection State
const serverUrl = ref(localStorage.getItem('apphub_server_url') || 'http://localhost:8080')
const serverStatus = ref('checking') // 'online' | 'offline' | 'checking'
let healthTimer = null

// Modal States
const isJsonModalOpen = ref(false)
const isServerModalOpen = ref(false)

// Execution States
const isLoading = ref(false)
const executionError = ref(null)
const responseData = ref(null)

// Toast Notifications
const toasts = ref([])

function showToast(message, type = 'info') {
  const id = Date.now() + Math.random()
  toasts.value.push({ id, message, type })
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, 4000)
}

// Request Payload Model
const payload = ref({
  selector: {
    autoDetect: true,
    label: null,
    tag: null,
    logLabel: null,
    contains: null,
    projectKeys: null,
    perK8sNamespace: false,
    perK8sAppLabel: false
  },
  scope: {
    parent: 'projects/my-gcp-project',
    locations: ['us-central1', 'global'],
    managementProject: 'my-apphub-host-project'
  },
  action: {
    reportOnly: true
  },
  options: {
    attributes: {
      criticality: { type: 'MISSION_CRITICAL' },
      environment: { type: 'PRODUCTION' },
      developerOwners: [
        { email: 'dev-lead@example.com', displayName: 'Lead Engineer' }
      ],
      operatorOwners: [
        { email: 'sre-team@example.com', displayName: 'SRE Team' }
      ],
      businessOwners: [
        { email: 'pm@example.com', displayName: 'Product Manager' }
      ]
    },
    assetTypes: []
  }
})

// Check Server Health
async function checkServerHealth() {
  const base = serverUrl.value.replace(/\/+$/, '')
  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 3000)
    const res = await fetch(`${base}/`, { signal: controller.signal })
    clearTimeout(timeoutId)
    if (res.ok) {
      serverStatus.value = 'online'
    } else {
      serverStatus.value = 'offline'
    }
  } catch (err) {
    serverStatus.value = 'offline'
  }
}

// Execute Discovery Request
async function executeDiscovery() {
  if (!payload.value.scope.parent) {
    showToast('Parent scope is required (e.g. projects/my-project)', 'error')
    return
  }

  isLoading.value = true
  executionError.value = null

  const base = serverUrl.value.replace(/\/+$/, '')

  try {
    const res = await fetch(`${base}/generate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload.value)
    })

    const data = await res.json()

    if (!res.ok) {
      throw new Error(data.error || `HTTP error ${res.status}: ${res.statusText}`)
    }

    responseData.value = data
    serverStatus.value = 'online'
    showToast(
      payload.value.action.reportOnly 
        ? 'Discovery report generated successfully!' 
        : 'App Hub applications created successfully!', 
      'success'
    )
  } catch (err) {
    executionError.value = err.message
    showToast(err.message, 'error')
  } finally {
    isLoading.value = false
  }
}

// Load Sample Presets
function loadSamplePreset(sampleId) {
  if (sampleId === 'sample1') {
    // Auto-detect
    payload.value = {
      selector: { autoDetect: true },
      scope: {
        parent: 'projects/demo-ecommerce-prod',
        locations: ['us-central1', 'global'],
        managementProject: 'demo-apphub-host'
      },
      action: { reportOnly: true },
      options: {
        attributes: {
          criticality: { type: 'MISSION_CRITICAL' },
          environment: { type: 'PRODUCTION' },
          developerOwners: [{ email: 'dev@example.com', displayName: 'Dev Team' }]
        }
      }
    }
    showToast('Loaded Auto-Detect sample preset', 'info')
  } else if (sampleId === 'sample2') {
    // Resource Label
    payload.value = {
      selector: { label: { key: 'env', value: 'prod' } },
      scope: {
        parent: 'projects/demo-ecommerce-prod',
        locations: ['us-central1', 'us-east1', 'global'],
        managementProject: 'demo-apphub-host'
      },
      action: { reportOnly: true },
      options: {
        attributes: {
          criticality: { type: 'HIGH' },
          environment: { type: 'PRODUCTION' },
          developerOwners: [{ email: 'payments-dev@example.com', displayName: 'Payments Team' }]
        }
      }
    }
    showToast('Loaded Label (env=prod) sample preset', 'info')
  } else if (sampleId === 'sample3') {
    // Resource Tag
    payload.value = {
      selector: { tag: { key: '1234567890/cost-center', value: 'billing' } },
      scope: {
        parent: 'folders/1234567890',
        locations: ['us-central1', 'global'],
        managementProject: 'demo-apphub-host'
      },
      action: { reportOnly: true },
      options: {
        attributes: {
          criticality: { type: 'MEDIUM' },
          environment: { type: 'DEVELOPMENT' }
        }
      }
    }
    showToast('Loaded Resource Manager Tag sample preset', 'info')
  } else if (sampleId === 'sample4') {
    // Per K8s Namespace
    payload.value = {
      selector: { perK8sNamespace: true },
      scope: {
        parent: 'projects/gke-cluster-prod',
        locations: ['us-central1', 'europe-west1'],
        managementProject: 'demo-apphub-host'
      },
      action: { reportOnly: true },
      options: {}
    }
    showToast('Loaded K8s Namespace isolation preset', 'info')
  } else if (sampleId === 'sample5') {
    // Multi-Project
    payload.value = {
      selector: {
        projectKeys: {
          appName: 'core-banking-suite',
          projectIds: ['bank-frontend-prod', 'bank-ledger-prod', 'bank-auth-prod']
        }
      },
      scope: {
        parent: 'folders/987654321',
        locations: ['us-central1', 'us-east4', 'global'],
        managementProject: 'demo-apphub-host'
      },
      action: { reportOnly: true },
      options: {
        attributes: {
          criticality: { type: 'MISSION_CRITICAL' },
          environment: { type: 'PRODUCTION' }
        }
      }
    }
    showToast('Loaded Multi-Project Aggregator preset', 'info')
  }
}

function handleSaveServerUrl(newUrl) {
  serverUrl.value = newUrl
  localStorage.setItem('apphub_server_url', newUrl)
  checkServerHealth()
  showToast(`Updated backend URL to ${newUrl}`, 'info')
}

onMounted(() => {
  checkServerHealth()
  healthTimer = setInterval(checkServerHealth, 15000)
})

onUnmounted(() => {
  if (healthTimer) clearInterval(healthTimer)
})
</script>

<template>
  <div class="app-layout">
    <!-- Header Navigation -->
    <HeaderNav 
      :server-status="serverStatus"
      :server-url="serverUrl"
      @open-server-modal="isServerModalOpen = true"
      @open-json-modal="isJsonModalOpen = true"
      @load-sample="loadSamplePreset"
    />

    <!-- Main Workspace -->
    <main class="main-content container">
      <!-- Two Column Layout: Left Form, Right Results -->
      <div class="workspace-grid">
        <div class="workspace-column left-pane">
          <SelectorForm 
            v-model="payload"
            :is-loading="isLoading"
            @execute="executeDiscovery"
          />
        </div>

        <div class="workspace-column right-pane">
          <ResultsViewer 
            :response="responseData"
            :error="executionError"
            :is-loading="isLoading"
            :report-only="payload.action.reportOnly"
            @notify="showToast($event.message, $event.type)"
          />
        </div>
      </div>
    </main>

    <!-- JSON Editor Modal -->
    <JsonEditorModal 
      :is-open="isJsonModalOpen"
      :payload="payload"
      :is-loading="isLoading"
      @close="isJsonModalOpen = false"
      @apply="payload = $event; showToast('JSON synchronized to form', 'info')"
      @execute="executeDiscovery"
    />

    <!-- Server Settings Modal -->
    <ServerSettingsModal 
      :is-open="isServerModalOpen"
      :server-url="serverUrl"
      @close="isServerModalOpen = false"
      @save="handleSaveServerUrl"
    />

    <!-- Global Toast Container -->
    <div class="toast-container">
      <div 
        v-for="toast in toasts" 
        :key="toast.id"
        class="toast"
        :class="'toast-' + toast.type"
      >
        <span class="icon" style="font-size: 18px;">
          {{ toast.type === 'success' ? 'check_circle' : (toast.type === 'error' ? 'error' : 'info') }}
        </span>
        <span>{{ toast.message }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-app);
}

.main-content {
  flex: 1;
  padding-top: 24px;
  padding-bottom: 48px;
}

.workspace-grid {
  display: grid;
  grid-template-columns: 460px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}

@media (max-width: 1100px) {
  .workspace-grid {
    grid-template-columns: 1fr;
  }
}
</style>
