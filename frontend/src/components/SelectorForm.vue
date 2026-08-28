<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  isLoading: {
    type: Boolean,
    default: false
  },
  modelValue: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['update:modelValue', 'execute'])

const activeTab = ref('autoDetect')
const showAdvancedAttributes = ref(false)

const tabs = [
  { id: 'autoDetect', label: 'Auto-Detect', icon: 'auto_awesome', desc: 'Discovers apps using standard well-known labels' },
  { id: 'label', label: 'Resource Label', icon: 'label', desc: 'Discovers assets matching a GCP resource label' },
  { id: 'tag', label: 'Resource Tag', icon: 'sell', desc: 'Discovers assets matching a Resource Manager tag' },
  { id: 'logLabel', label: 'Cloud Logging', icon: 'receipt_long', desc: 'Discovers active runtimes from Cloud Logging entries' },
  { id: 'perK8sNamespace', label: 'K8s Namespace', icon: 'lan', desc: 'Generates one App Hub app per Kubernetes namespace' },
  { id: 'perK8sAppLabel', label: 'K8s App Label', icon: 'category', desc: 'Generates one app per app.kubernetes.io/name label' },
  { id: 'projectKeys', label: 'Projects', icon: 'folder_copy', desc: 'Groups assets from multiple projects into a named application' },
  { id: 'contains', label: 'Substring', icon: 'search', desc: 'Filters resources whose resource name contains substring' },
]

const commonRegions = [
  'global', 'us-central1', 'us-east1', 'us-east4', 'us-west1', 'us-west2',
  'europe-west1', 'europe-west3', 'europe-west4',
  'asia-east1', 'asia-northeast1', 'asia-southeast1'
]

const newLocationInput = ref('')
const newProjectInput = ref('')

// Owner inputs
const newDevOwner = ref({ email: '', displayName: '' })
const newOpOwner = ref({ email: '', displayName: '' })
const newBizOwner = ref({ email: '', displayName: '' })

function setTab(tabId) {
  activeTab.value = tabId
  const val = { ...props.modelValue }
  val.selector = {
    autoDetect: false,
    label: null,
    tag: null,
    logLabel: null,
    contains: null,
    projectKeys: null,
    perK8sNamespace: false,
    perK8sAppLabel: false
  }

  if (tabId === 'autoDetect') {
    val.selector.autoDetect = true
  } else if (tabId === 'label') {
    val.selector.label = { key: 'app', value: '' }
  } else if (tabId === 'tag') {
    val.selector.tag = { key: 'environment', value: '' }
  } else if (tabId === 'logLabel') {
    val.selector.logLabel = { key: 'service_name', value: '' }
  } else if (tabId === 'perK8sNamespace') {
    val.selector.perK8sNamespace = true
  } else if (tabId === 'perK8sAppLabel') {
    val.selector.perK8sAppLabel = true
  } else if (tabId === 'projectKeys') {
    val.selector.projectKeys = { appName: 'my-agg-app', projectIds: ['my-project-1', 'my-project-2'] }
  } else if (tabId === 'contains') {
    val.selector.contains = 'frontend'
  }

  emit('update:modelValue', val)
}

function toggleLocation(loc) {
  const val = { ...props.modelValue }
  const current = val.scope?.locations || []
  if (current.includes(loc)) {
    val.scope.locations = current.filter(l => l !== loc)
  } else {
    val.scope.locations = [...current, loc]
  }
  emit('update:modelValue', val)
}

function addCustomLocation() {
  const loc = newLocationInput.value.trim().toLowerCase()
  if (!loc) return
  const val = { ...props.modelValue }
  const current = val.scope?.locations || []
  if (!current.includes(loc)) {
    val.scope.locations = [...current, loc]
    emit('update:modelValue', val)
  }
  newLocationInput.value = ''
}

function removeLocation(loc) {
  const val = { ...props.modelValue }
  val.scope.locations = (val.scope?.locations || []).filter(l => l !== loc)
  emit('update:modelValue', val)
}

function addProjectId() {
  const id = newProjectInput.value.trim()
  if (!id) return
  const val = { ...props.modelValue }
  if (val.selector?.projectKeys) {
    if (!val.selector.projectKeys.projectIds.includes(id)) {
      val.selector.projectKeys.projectIds.push(id)
      emit('update:modelValue', val)
    }
  }
  newProjectInput.value = ''
}

function removeProjectId(id) {
  const val = { ...props.modelValue }
  if (val.selector?.projectKeys) {
    val.selector.projectKeys.projectIds = val.selector.projectKeys.projectIds.filter(p => p !== id)
    emit('update:modelValue', val)
  }
}

function addOwner(type) {
  const val = { ...props.modelValue }
  if (!val.options) val.options = {}
  if (!val.options.attributes) {
    val.options.attributes = {
      criticality: { type: 'MISSION_CRITICAL' },
      environment: { type: 'PRODUCTION' },
      developerOwners: [],
      operatorOwners: [],
      businessOwners: []
    }
  }

  let source = null
  let target = null
  if (type === 'dev') {
    source = newDevOwner.value
    target = val.options.attributes.developerOwners = val.options.attributes.developerOwners || []
  } else if (type === 'op') {
    source = newOpOwner.value
    target = val.options.attributes.operatorOwners = val.options.attributes.operatorOwners || []
  } else if (type === 'biz') {
    source = newBizOwner.value
    target = val.options.attributes.businessOwners = val.options.attributes.businessOwners || []
  }

  if (source && source.email.trim()) {
    target.push({ email: source.email.trim(), displayName: source.displayName.trim() || source.email.trim() })
    source.email = ''
    source.displayName = ''
    emit('update:modelValue', val)
  }
}

function removeOwner(type, index) {
  const val = { ...props.modelValue }
  if (val.options?.attributes) {
    if (type === 'dev' && val.options.attributes.developerOwners) {
      val.options.attributes.developerOwners.splice(index, 1)
    } else if (type === 'op' && val.options.attributes.operatorOwners) {
      val.options.attributes.operatorOwners.splice(index, 1)
    } else if (type === 'biz' && val.options.attributes.businessOwners) {
      val.options.attributes.businessOwners.splice(index, 1)
    }
    emit('update:modelValue', val)
  }
}
</script>

<template>
  <div class="selector-form card">
    <!-- Top Selector Tabs Strip -->
    <div class="section-title">
      <span class="icon">filter_alt</span>
      <span>1. Discovery Strategy & Selectors</span>
    </div>

    <div class="tabs-grid">
      <button 
        v-for="t in tabs" 
        :key="t.id"
        class="tab-btn"
        :class="{ active: activeTab === t.id }"
        @click="setTab(t.id)"
        :id="'tab-' + t.id"
      >
        <span class="icon">{{ t.icon }}</span>
        <span class="tab-label">{{ t.label }}</span>
      </button>
    </div>

    <!-- Active Tab Configuration Pane -->
    <div class="tab-pane">
      <div v-if="activeTab === 'autoDetect'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">auto_awesome</span>
          <div>
            <div class="pane-title">Auto-Detect Discovery</div>
            <div class="pane-desc">Scans assets for well-known application grouping labels: <code>app</code>, <code>application</code>, and <code>app.kubernetes.io/name</code>.</div>
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'label'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">label</span>
          <div>
            <div class="pane-title">GCP Resource Label Filter</div>
            <div class="pane-desc">Discover resources grouped by custom GCP labels (e.g. <code>env=prod</code> or <code>service=payment</code>).</div>
          </div>
        </div>
        <div class="grid-2col" v-if="modelValue.selector?.label">
          <div class="form-group">
            <label class="form-label" for="input-label-key">Label Key *</label>
            <input 
              id="input-label-key"
              class="form-input code-font" 
              v-model="modelValue.selector.label.key" 
              placeholder="e.g. env" 
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="input-label-value">Label Value (Optional)</label>
            <input 
              id="input-label-value"
              class="form-input code-font" 
              v-model="modelValue.selector.label.value" 
              placeholder="e.g. prod (leave empty for any value)" 
            />
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'tag'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">sell</span>
          <div>
            <div class="pane-title">Resource Manager Tag Filter</div>
            <div class="pane-desc">Discover resources bound to GCP Resource Manager Tags (e.g., <code>123456789/cost-center</code>).</div>
          </div>
        </div>
        <div class="grid-2col" v-if="modelValue.selector?.tag">
          <div class="form-group">
            <label class="form-label" for="input-tag-key">Tag Key *</label>
            <input 
              id="input-tag-key"
              class="form-input code-font" 
              v-model="modelValue.selector.tag.key" 
              placeholder="e.g. 1234567890/environment" 
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="input-tag-value">Tag Value (Optional)</label>
            <input 
              id="input-tag-value"
              class="form-input code-font" 
              v-model="modelValue.selector.tag.value" 
              placeholder="e.g. production" 
            />
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'logLabel'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">receipt_long</span>
          <div>
            <div class="pane-title">Cloud Logging Discovery</div>
            <div class="pane-desc">Query Cloud Logging entries to identify active runtime services and deployments.</div>
          </div>
        </div>
        <div class="grid-2col" v-if="modelValue.selector?.logLabel">
          <div class="form-group">
            <label class="form-label" for="input-log-key">Log Label Key *</label>
            <input 
              id="input-log-key"
              class="form-input code-font" 
              v-model="modelValue.selector.logLabel.key" 
              placeholder="e.g. service_name" 
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="input-log-value">Log Label Value (Optional)</label>
            <input 
              id="input-log-value"
              class="form-input code-font" 
              v-model="modelValue.selector.logLabel.value" 
              placeholder="e.g. frontend" 
            />
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'perK8sNamespace'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">lan</span>
          <div>
            <div class="pane-title">Kubernetes Namespace Isolation</div>
            <div class="pane-desc">Automatically discovers all GKE namespaces and creates a dedicated App Hub application for each namespace.</div>
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'perK8sAppLabel'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">category</span>
          <div>
            <div class="pane-title">Kubernetes App Name Label</div>
            <div class="pane-desc">Groups Kubernetes resources by their standard <code>app.kubernetes.io/name</code> metadata label.</div>
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'projectKeys'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">folder_copy</span>
          <div>
            <div class="pane-title">Multi-Project Aggregator</div>
            <div class="pane-desc">Aggregates all workloads and services across multiple GCP projects into a single named App Hub application.</div>
          </div>
        </div>
        <div v-if="modelValue.selector?.projectKeys">
          <div class="form-group">
            <label class="form-label" for="input-proj-app-name">Application Name *</label>
            <input 
              id="input-proj-app-name"
              class="form-input code-font" 
              v-model="modelValue.selector.projectKeys.appName" 
              placeholder="e.g. core-billing-app" 
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label">Project IDs Included</label>
            <div class="chips-container">
              <span 
                v-for="pid in modelValue.selector.projectKeys.projectIds" 
                :key="pid" 
                class="chip"
              >
                <span class="icon" style="font-size: 14px;">terminal</span>
                <code>{{ pid }}</code>
                <button type="button" class="chip-remove" @click="removeProjectId(pid)">×</button>
              </span>
            </div>
            <div class="input-with-btn" style="margin-top: 8px;">
              <input 
                class="form-input code-font" 
                v-model="newProjectInput" 
                placeholder="Add GCP project ID (e.g. payment-prod-123)"
                @keyup.enter="addProjectId"
                id="input-add-project"
              />
              <button type="button" class="btn btn-secondary btn-sm" @click="addProjectId">
                <span class="icon">add</span> Add Project
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="activeTab === 'contains'" class="tab-pane-box">
        <div class="pane-header">
          <span class="icon">search</span>
          <div>
            <div class="pane-title">Resource Name Substring Filter</div>
            <div class="pane-desc">Filters resources whose full GCP resource URI or name contains the given substring.</div>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label" for="input-contains">Substring Pattern *</label>
          <input 
            id="input-contains"
            class="form-input code-font" 
            v-model="modelValue.selector.contains" 
            placeholder="e.g. backend-api" 
            required
          />
        </div>
      </div>
    </div>

    <!-- Section 2: Scope Configuration -->
    <div class="section-title" style="margin-top: 24px;">
      <span class="icon">travel_explore</span>
      <span>2. Scope & Target Locations</span>
    </div>

    <div class="scope-grid">
      <div class="form-group">
        <label class="form-label" for="input-parent-scope">
          <span>Parent Scope *</span>
          <span class="form-hint">projects/{id} or folders/{id}</span>
        </label>
        <input 
          id="input-parent-scope"
          class="form-input code-font" 
          v-model="modelValue.scope.parent" 
          placeholder="projects/my-gcp-project" 
          required
        />
      </div>

      <div class="form-group">
        <label class="form-label" for="input-mgmt-project">
          <span>Management Project (App Hub Host) *</span>
          <span class="form-hint">Host project for App Hub</span>
        </label>
        <input 
          id="input-mgmt-project"
          class="form-input code-font" 
          v-model="modelValue.scope.managementProject" 
          placeholder="my-apphub-host-project" 
          required
        />
      </div>
    </div>

    <!-- Locations Chip Picker -->
    <div class="form-group">
      <label class="form-label">
        <span>GCP Locations / Regions to Scan</span>
        <span class="form-hint">{{ modelValue.scope.locations.length }} selected</span>
      </label>
      <div class="chips-container">
        <button 
          v-for="loc in commonRegions" 
          :key="loc"
          type="button"
          class="location-chip"
          :class="{ selected: modelValue.scope.locations.includes(loc) }"
          @click="toggleLocation(loc)"
        >
          <span class="icon" style="font-size: 14px;">{{ loc === 'global' ? 'public' : 'location_on' }}</span>
          <span>{{ loc }}</span>
        </button>
      </div>

      <!-- Custom Location Input -->
      <div class="input-with-btn" style="margin-top: 10px; max-width: 400px;">
        <input 
          class="form-input code-font form-input-sm" 
          v-model="newLocationInput" 
          placeholder="Add other location (e.g. us-south1)"
          @keyup.enter="addCustomLocation"
          id="input-custom-location"
        />
        <button type="button" class="btn btn-secondary btn-sm" @click="addCustomLocation">
          <span class="icon">add</span> Add
        </button>
      </div>
    </div>

    <!-- Section 3: Action & Attributes -->
    <div class="section-title" style="margin-top: 24px;">
      <span class="icon">tune</span>
      <span>3. Execution Mode & App Hub Attributes</span>
    </div>

    <div class="action-card">
      <div class="action-toggle-row">
        <div class="toggle-info">
          <div class="toggle-title">Report-Only / Dry Run Preview</div>
          <div class="toggle-desc">
            Discovers resources and displays planned App Hub applications without mutating GCP resources.
          </div>
        </div>
        <label class="switch" for="switch-report-only">
          <input 
            type="checkbox" 
            id="switch-report-only"
            v-model="modelValue.action.reportOnly" 
          />
          <span class="slider"></span>
        </label>
      </div>

      <div class="mode-badge-row">
        <span 
          class="badge" 
          :class="modelValue.action.reportOnly ? 'badge-warning' : 'badge-success'"
        >
          <span class="icon" style="font-size: 14px;">{{ modelValue.action.reportOnly ? 'visibility' : 'cloud_upload' }}</span>
          {{ modelValue.action.reportOnly ? 'DRY-RUN PREVIEW (Safe)' : 'LIVE CREATE (Mutates GCP App Hub)' }}
        </span>
      </div>
    </div>

    <!-- Attributes Accordion Toggle -->
    <div class="attributes-toggle-area">
      <button 
        type="button" 
        class="btn btn-ghost btn-sm"
        @click="showAdvancedAttributes = !showAdvancedAttributes"
        id="btn-toggle-attributes"
      >
        <span class="icon">{{ showAdvancedAttributes ? 'expand_less' : 'expand_more' }}</span>
        <span>App Hub Metadata & Ownership Attributes (Criticality, Environment, Contacts)</span>
      </button>
    </div>

    <!-- Advanced Attributes Drawer -->
    <div v-if="showAdvancedAttributes" class="attributes-drawer">
      <div class="grid-2col">
        <div class="form-group">
          <label class="form-label" for="select-criticality">Criticality Level</label>
          <select 
            id="select-criticality"
            class="form-select" 
            v-if="modelValue.options?.attributes?.criticality"
            v-model="modelValue.options.attributes.criticality.type"
          >
            <option value="MISSION_CRITICAL">MISSION_CRITICAL (Highest)</option>
            <option value="HIGH">HIGH</option>
            <option value="MEDIUM">MEDIUM</option>
            <option value="LOW">LOW</option>
            <option value="CRITICALITY_TYPE_UNSPECIFIED">UNSPECIFIED</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label" for="select-environment">Environment / Severity</label>
          <select 
            id="select-environment"
            class="form-select" 
            v-if="modelValue.options?.attributes?.environment"
            v-model="modelValue.options.attributes.environment.type"
          >
            <option value="PRODUCTION">PRODUCTION</option>
            <option value="STAGING">STAGING</option>
            <option value="TEST">TEST</option>
            <option value="DEVELOPMENT">DEVELOPMENT</option>
            <option value="ENVIRONMENT_TYPE_UNSPECIFIED">UNSPECIFIED</option>
          </select>
        </div>
      </div>

      <!-- Developer Owners -->
      <div class="owners-block" v-if="modelValue.options?.attributes">
        <div class="owners-title">Developer Owners</div>
        <div class="chips-container" v-if="modelValue.options.attributes.developerOwners?.length">
          <span 
            v-for="(dev, idx) in modelValue.options.attributes.developerOwners" 
            :key="idx" 
            class="chip"
          >
            <span class="icon" style="font-size: 14px;">code</span>
            <span>{{ dev.displayName }} ({{ dev.email }})</span>
            <button type="button" class="chip-remove" @click="removeOwner('dev', idx)">×</button>
          </span>
        </div>
        <div class="input-with-btn" style="margin-top: 6px;">
          <input 
            class="form-input form-input-sm" 
            v-model="newDevOwner.email" 
            placeholder="Email (e.g. dev-lead@example.com)"
          />
          <input 
            class="form-input form-input-sm" 
            v-model="newDevOwner.displayName" 
            placeholder="Display Name"
          />
          <button type="button" class="btn btn-secondary btn-sm" @click="addOwner('dev')">
            Add
          </button>
        </div>
      </div>
    </div>

    <!-- Submit CTA Button -->
    <div class="form-actions">
      <button 
        type="button" 
        class="btn btn-primary btn-lg submit-btn"
        :disabled="isLoading || !modelValue.scope.parent"
        @click="emit('execute')"
        id="btn-execute-generate"
      >
        <span v-if="isLoading" class="spinner"></span>
        <span v-else class="icon">{{ modelValue.action.reportOnly ? 'preview' : 'rocket_launch' }}</span>
        <span>{{ isLoading ? 'Discovering Resources...' : (modelValue.action.reportOnly ? 'Generate App Hub Report' : 'Create App Hub Applications') }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.selector-form {
  margin-bottom: 24px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: #60a5fa;
  margin-bottom: 14px;
}

.tabs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: 8px;
  margin-bottom: 16px;
}

.tab-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px 8px;
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn:hover {
  background-color: var(--bg-surface-hover);
  color: var(--text-primary);
  border-color: #3b82f6;
}

.tab-btn.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: #3b82f6;
  color: #60a5fa;
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.2);
}

.tab-btn .icon {
  font-size: 22px;
}

.tab-pane-box {
  background-color: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 16px;
}

.pane-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 14px;
}

.pane-header .icon {
  color: #3b82f6;
  font-size: 24px;
}

.pane-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.pane-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.pane-desc code {
  background-color: var(--bg-surface-elevated);
  padding: 2px 5px;
  border-radius: 4px;
  color: #93c5fd;
  font-family: var(--font-mono);
}

.grid-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.scope-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.chips-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 4px;
}

.location-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  color: var(--text-secondary);
  font-size: 12px;
  font-family: var(--font-mono);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.location-chip:hover {
  border-color: #60a5fa;
  color: var(--text-primary);
}

.location-chip.selected {
  background-color: rgba(59, 130, 246, 0.2);
  border-color: #3b82f6;
  color: #93c5fd;
  font-weight: 600;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  font-size: 12px;
  color: var(--text-primary);
}

.chip-remove {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 14px;
  cursor: pointer;
  line-height: 1;
}
.chip-remove:hover {
  color: var(--color-error);
}

.input-with-btn {
  display: flex;
  gap: 8px;
  align-items: center;
}

.form-input-sm {
  padding: 6px 10px;
  font-size: 13px;
}

.action-card {
  background-color: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 16px;
}

.action-toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.toggle-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.toggle-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

/* Switch Toggle */
.switch {
  position: relative;
  display: inline-block;
  width: 46px;
  height: 24px;
  flex-shrink: 0;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background-color: #374151;
  transition: .3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #3b82f6;
}

input:checked + .slider:before {
  transform: translateX(22px);
}

.mode-badge-row {
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px dashed var(--border-subtle);
}

.attributes-toggle-area {
  margin: 12px 0;
}

.attributes-drawer {
  background-color: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 20px;
  animation: fadeIn 0.2s ease-out;
}

.owners-block {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
}

.owners-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.form-actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.submit-btn {
  width: 100%;
  max-width: 320px;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #ffffff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 768px) {
  .grid-2col, .scope-grid {
    grid-template-columns: 1fr;
  }
  .tabs-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .submit-btn {
    max-width: 100%;
  }
}
</style>
