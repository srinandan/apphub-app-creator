<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  response: {
    type: Object,
    default: null
  },
  error: {
    type: String,
    default: null
  },
  isLoading: {
    type: Boolean,
    default: false
  },
  reportOnly: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['copy-text', 'notify'])

const searchQuery = ref('')
const viewMode = ref('cards') // 'cards' | 'json'
const expandedApps = ref({})

// Computed statistics
const applicationsList = computed(() => {
  if (!props.response?.applications) return []
  return Object.entries(props.response.applications).map(([key, app]) => ({
    key,
    name: app.name || key,
    workloads: app.workloads || [],
    services: app.services || []
  }))
})

const totalWorkloadsCount = computed(() => {
  return applicationsList.value.reduce((acc, app) => acc + (app.workloads?.length || 0), 0)
})

const totalServicesCount = computed(() => {
  return applicationsList.value.reduce((acc, app) => acc + (app.services?.length || 0), 0)
})

const totalResourcesCount = computed(() => {
  return totalWorkloadsCount.value + totalServicesCount.value
})

const filteredApplications = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return applicationsList.value

  return applicationsList.value.filter(app => {
    if (app.name.toLowerCase().includes(q) || app.key.toLowerCase().includes(q)) return true
    
    // Check in workloads
    const matchedWorkload = app.workloads.some(w => 
      (w.appHubId && w.appHubId.toLowerCase().includes(q)) || 
      (w.uri && w.uri.toLowerCase().includes(q))
    )
    if (matchedWorkload) return true

    // Check in services
    const matchedService = app.services.some(s => 
      (s.appHubId && s.appHubId.toLowerCase().includes(q)) || 
      (s.uri && s.uri.toLowerCase().includes(q))
    )
    return matchedService
  })
})

function toggleAppExpand(key) {
  expandedApps.value[key] = !isExpanded(key)
}

function isExpanded(key) {
  return expandedApps.value[key] !== false // default expanded
}

function expandAll() {
  const map = {}
  applicationsList.value.forEach(app => {
    map[app.key] = true
  })
  expandedApps.value = map
}

function collapseAll() {
  const map = {}
  applicationsList.value.forEach(app => {
    map[app.key] = false
  })
  expandedApps.value = map
}

function copyToClipboard(text, label = 'Content') {
  navigator.clipboard.writeText(text).then(() => {
    emit('notify', { message: `Copied ${label} to clipboard!`, type: 'success' })
  }).catch(() => {
    emit('notify', { message: 'Failed to copy to clipboard', type: 'error' })
  })
}

function downloadJsonReport() {
  if (!props.response) return
  const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(props.response, null, 2))
  const downloadAnchor = document.createElement('a')
  downloadAnchor.setAttribute("href", dataStr)
  downloadAnchor.setAttribute("download", `apphub-discovery-report-${Date.now()}.json`)
  document.body.appendChild(downloadAnchor)
  downloadAnchor.click()
  downloadAnchor.remove()
  emit('notify', { message: 'JSON report downloaded successfully', type: 'success' })
}

function getAssetTypeBadge(uri) {
  if (!uri) return { label: 'Resource', icon: 'layers' }
  if (uri.includes('run.googleapis.com')) return { label: 'Cloud Run', icon: 'directions_run' }
  if (uri.includes('apps/deployments') || uri.includes('apps.k8s.io')) return { label: 'GKE Workload', icon: 'view_in_ar' }
  if (uri.includes('compute.googleapis.com/instances') || uri.includes('InstanceGroup')) return { label: 'Compute Engine', icon: 'dns' }
  if (uri.includes('sqladmin.googleapis.com') || uri.includes('alloydb') || uri.includes('spanner') || uri.includes('redis')) return { label: 'Database / Cache', icon: 'database' }
  if (uri.includes('storage.googleapis.com')) return { label: 'Cloud Storage', icon: 'folder' }
  if (uri.includes('pubsub.googleapis.com')) return { label: 'Pub/Sub', icon: 'mark_chat_read' }
  if (uri.includes('aiplatform') || uri.includes('discoveryengine') || uri.includes('agentregistry')) return { label: 'AI & Agents', icon: 'smart_toy' }
  if (uri.includes('ForwardingRule') || uri.includes('BackendService') || uri.includes('Gateway')) return { label: 'Networking', icon: 'hub' }
  return { label: 'GCP Asset', icon: 'cloud' }
}
</script>

<template>
  <div class="results-viewer card">
    <!-- Results Header -->
    <div class="results-header">
      <div class="results-title-group">
        <div class="title-with-badge">
          <span class="icon">dashboard</span>
          <h2>Discovery & Generation Results</h2>
          <span v-if="response" class="badge" :class="reportOnly ? 'badge-warning' : 'badge-success'">
            {{ reportOnly ? 'Dry Run' : 'Created' }}
          </span>
        </div>
        <p class="results-subtitle">
          Discovered Google Cloud App Hub applications, mapped workloads, and registered services.
        </p>
      </div>

      <!-- Controls: View Toggle & Export -->
      <div class="results-actions" v-if="response && applicationsList.length > 0">
        <div class="view-toggle">
          <button 
            class="toggle-btn" 
            :class="{ active: viewMode === 'cards' }"
            @click="viewMode = 'cards'"
            title="Visual Application Cards"
            id="btn-view-cards"
          >
            <span class="icon" style="font-size: 16px;">view_agenda</span>
            Cards
          </button>
          <button 
            class="toggle-btn" 
            :class="{ active: viewMode === 'json' }"
            @click="viewMode = 'json'"
            title="Raw JSON Output"
            id="btn-view-json"
          >
            <span class="icon" style="font-size: 16px;">code</span>
            JSON
          </button>
        </div>

        <button 
          class="btn btn-secondary btn-sm" 
          @click="downloadJsonReport"
          title="Download full JSON report"
          id="btn-download-json"
        >
          <span class="icon">download</span>
          Export JSON
        </button>
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="error" class="error-banner">
      <div class="error-icon">
        <span class="icon">error</span>
      </div>
      <div class="error-content">
        <div class="error-title">Request Failed</div>
        <div class="error-message">{{ error }}</div>
      </div>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner-large"></div>
      <div class="loading-text">Scanning Google Cloud Asset Inventory & Cloud Logging...</div>
      <div class="loading-subtext">Mapping services and workloads to App Hub applications</div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!response && !error" class="empty-state">
      <div class="empty-icon-box">
        <span class="icon">travel_explore</span>
      </div>
      <h3 class="empty-title">Ready for Discovery</h3>
      <p class="empty-desc">
        Configure your discovery selector and scope on the left, then click <strong>"Generate App Hub Report"</strong> to preview discovered applications and resources.
      </p>
      <div class="quick-tips">
        <div class="tip-card">
          <span class="icon">bolt</span>
          <div class="tip-text"><strong>Tip:</strong> Use <strong>Auto-Detect</strong> to discover standard <code>app</code> and GKE labels.</div>
        </div>
        <div class="tip-card">
          <span class="icon">shield</span>
          <div class="tip-text"><strong>Safe Preview:</strong> Keep <strong>Report-Only</strong> checked to preview without making GCP changes.</div>
        </div>
      </div>
    </div>

    <!-- Results Content -->
    <div v-else-if="response">
      <!-- Metric Summary Cards -->
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-icon metric-apps">
            <span class="icon">hub</span>
          </div>
          <div class="metric-info">
            <div class="metric-value">{{ applicationsList.length }}</div>
            <div class="metric-label">Applications</div>
          </div>
        </div>

        <div class="metric-card">
          <div class="metric-icon metric-workloads">
            <span class="icon">view_in_ar</span>
          </div>
          <div class="metric-info">
            <div class="metric-value">{{ totalWorkloadsCount }}</div>
            <div class="metric-label">Workloads</div>
          </div>
        </div>

        <div class="metric-card">
          <div class="metric-icon metric-services">
            <span class="icon">dns</span>
          </div>
          <div class="metric-info">
            <div class="metric-value">{{ totalServicesCount }}</div>
            <div class="metric-label">Services</div>
          </div>
        </div>

        <div class="metric-card">
          <div class="metric-icon metric-total">
            <span class="icon">inventory_2</span>
          </div>
          <div class="metric-info">
            <div class="metric-value">{{ totalResourcesCount }}</div>
            <div class="metric-label">Total Assets</div>
          </div>
        </div>
      </div>

      <!-- Search & Filter Bar (Cards View) -->
      <div v-if="viewMode === 'cards'" class="search-filter-bar">
        <div class="search-input-wrapper">
          <span class="icon search-icon">search</span>
          <input 
            class="search-input" 
            v-model="searchQuery" 
            placeholder="Search applications, workloads, services, URIs..."
            id="input-search-results"
          />
          <button v-if="searchQuery" class="clear-search-btn" @click="searchQuery = ''">×</button>
        </div>

        <div class="filter-actions">
          <button class="btn btn-ghost btn-sm" @click="expandAll">
            <span class="icon">unfold_more</span> Expand All
          </button>
          <button class="btn btn-ghost btn-sm" @click="collapseAll">
            <span class="icon">unfold_less</span> Collapse All
          </button>
        </div>
      </div>

      <!-- No Match Filter State -->
      <div v-if="viewMode === 'cards' && filteredApplications.length === 0 && applicationsList.length > 0" class="no-matches">
        <span class="icon">search_off</span>
        <div>No applications matching "<strong>{{ searchQuery }}</strong>"</div>
        <button class="btn btn-secondary btn-sm" @click="searchQuery = ''">Clear Search</button>
      </div>

      <!-- Cards View -->
      <div v-if="viewMode === 'cards'" class="apps-list">
        <div 
          v-for="app in filteredApplications" 
          :key="app.key" 
          class="app-card"
        >
          <!-- Card Header -->
          <div class="app-card-header" @click="toggleAppExpand(app.key)">
            <div class="app-header-left">
              <button class="expand-btn">
                <span class="icon">{{ isExpanded(app.key) ? 'expand_more' : 'chevron_right' }}</span>
              </button>
              <div class="app-icon">
                <span class="icon">deployed_code</span>
              </div>
              <div class="app-name-wrap">
                <div class="app-name">{{ app.name }}</div>
                <div class="app-key">ID: <code>{{ app.key }}</code></div>
              </div>
            </div>

            <div class="app-header-right">
              <span class="badge badge-workload" title="Discovered Workloads">
                <span class="icon" style="font-size: 13px;">view_in_ar</span>
                {{ app.workloads.length }} Workload{{ app.workloads.length === 1 ? '' : 's' }}
              </span>
              <span class="badge badge-service" title="Discovered Services">
                <span class="icon" style="font-size: 13px;">dns</span>
                {{ app.services.length }} Service{{ app.services.length === 1 ? '' : 's' }}
              </span>
              <button 
                class="btn btn-ghost btn-sm copy-btn" 
                @click.stop="copyToClipboard(app.name, 'App Name')"
                title="Copy Application Name"
              >
                <span class="icon" style="font-size: 16px;">content_copy</span>
              </button>
            </div>
          </div>

          <!-- Card Body: Workloads & Services -->
          <div v-if="isExpanded(app.key)" class="app-card-body">
            <!-- Workloads Section -->
            <div class="resource-group">
              <div class="group-header">
                <div class="group-title workload-color">
                  <span class="icon">view_in_ar</span>
                  <span>Workloads ({{ app.workloads.length }})</span>
                </div>
              </div>

              <div v-if="app.workloads.length === 0" class="no-resources">
                No workloads discovered for this application.
              </div>

              <div v-else class="resource-items-list">
                <div 
                  v-for="(w, idx) in app.workloads" 
                  :key="idx" 
                  class="resource-item"
                >
                  <div class="resource-item-top">
                    <span class="asset-badge">
                      <span class="icon" style="font-size: 13px;">{{ getAssetTypeBadge(w.uri).icon }}</span>
                      {{ getAssetTypeBadge(w.uri).label }}
                    </span>
                    <span class="apphub-id-text">
                      ID: <code>{{ w.appHubId }}</code>
                    </span>
                    <button 
                      class="copy-icon-btn" 
                      @click="copyToClipboard(w.appHubId, 'Workload ID')"
                      title="Copy Workload ID"
                    >
                      <span class="icon">content_copy</span>
                    </button>
                  </div>
                  <div class="resource-uri-row">
                    <span class="uri-label">URI:</span>
                    <code class="uri-code" :title="w.uri">{{ w.uri }}</code>
                    <button 
                      class="copy-icon-btn" 
                      @click="copyToClipboard(w.uri, 'Workload URI')"
                      title="Copy Resource URI"
                    >
                      <span class="icon">content_copy</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Services Section -->
            <div class="resource-group">
              <div class="group-header">
                <div class="group-title service-color">
                  <span class="icon">dns</span>
                  <span>Services ({{ app.services.length }})</span>
                </div>
              </div>

              <div v-if="app.services.length === 0" class="no-resources">
                No services discovered for this application.
              </div>

              <div v-else class="resource-items-list">
                <div 
                  v-for="(s, idx) in app.services" 
                  :key="idx" 
                  class="resource-item"
                >
                  <div class="resource-item-top">
                    <span class="asset-badge asset-badge-service">
                      <span class="icon" style="font-size: 13px;">{{ getAssetTypeBadge(s.uri).icon }}</span>
                      {{ getAssetTypeBadge(s.uri).label }}
                    </span>
                    <span class="apphub-id-text">
                      ID: <code>{{ s.appHubId }}</code>
                    </span>
                    <button 
                      class="copy-icon-btn" 
                      @click="copyToClipboard(s.appHubId, 'Service ID')"
                      title="Copy Service ID"
                    >
                      <span class="icon">content_copy</span>
                    </button>
                  </div>
                  <div class="resource-uri-row">
                    <span class="uri-label">URI:</span>
                    <code class="uri-code" :title="s.uri">{{ s.uri }}</code>
                    <button 
                      class="copy-icon-btn" 
                      @click="copyToClipboard(s.uri, 'Service URI')"
                      title="Copy Resource URI"
                    >
                      <span class="icon">content_copy</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- JSON Raw Output View -->
      <div v-else-if="viewMode === 'json'" class="json-view-box">
        <div class="json-toolbar">
          <span class="json-title">HTTP Response (JSON)</span>
          <button 
            class="btn btn-secondary btn-sm" 
            @click="copyToClipboard(JSON.stringify(response, null, 2), 'JSON Response')"
            id="btn-copy-raw-json"
          >
            <span class="icon">content_copy</span> Copy Raw JSON
          </button>
        </div>
        <pre class="json-code"><code>{{ JSON.stringify(response, null, 2) }}</code></pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.results-viewer {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.results-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  min-width: 0;
}

.results-title-group {
  min-width: 0;
  flex: 1;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  min-width: 0;
}

.title-with-badge h2 {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  word-break: break-word;
}

.title-with-badge .icon {
  color: var(--color-brand);
  font-size: 24px;
}

.results-subtitle {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.results-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.view-toggle {
  display: flex;
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 2px;
}

.toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.toggle-btn.active {
  background-color: var(--bg-surface);
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
}

/* Metrics Cards */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 14px;
  min-width: 0;
}

.metric-card {
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 14px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.metric-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.metric-apps { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }
.metric-workloads { background: rgba(168, 85, 247, 0.15); color: #a855f7; }
.metric-services { background: rgba(14, 165, 233, 0.15); color: #0284c7; }
.metric-total { background: rgba(16, 185, 129, 0.15); color: #059669; }

.metric-info {
  min-width: 0;
}

.metric-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.1;
  word-break: break-all;
}

.metric-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-top: 2px;
}

/* Search Bar */
.search-filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  min-width: 0;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 240px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  font-size: 18px;
}

.search-input {
  width: 100%;
  padding: 9px 36px 9px 36px;
  background-color: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.search-input:focus {
  border-color: var(--border-focus);
}

.clear-search-btn {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 16px;
  cursor: pointer;
}

.filter-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

/* App Cards */
.apps-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.app-card {
  background-color: var(--bg-surface);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: border-color var(--transition-fast);
  min-width: 0;
}

.app-card:hover {
  border-color: var(--color-brand);
}

.app-card-header {
  padding: 14px 18px;
  background-color: var(--bg-surface-elevated);
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
  gap: 12px;
  flex-wrap: wrap;
  min-width: 0;
}

.app-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex: 1;
}

.expand-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.app-icon {
  width: 32px;
  height: 32px;
  background-color: var(--color-brand-subtle);
  color: var(--color-brand-text);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.app-name-wrap {
  min-width: 0;
  flex: 1;
}

.app-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-word;
  overflow-wrap: anywhere;
}

.app-key {
  font-size: 11px;
  color: var(--text-muted);
  word-break: break-all;
  overflow-wrap: anywhere;
}
.app-key code {
  font-family: var(--font-mono);
  color: var(--color-brand-text);
  word-break: break-all;
  overflow-wrap: anywhere;
}

.app-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
  flex-shrink: 0;
}

.copy-btn {
  padding: 4px;
}

.app-card-body {
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  background-color: var(--bg-surface);
  border-top: 1px solid var(--border-color);
  min-width: 0;
}

.group-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}

.workload-color { color: var(--badge-workload-color); }
.service-color { color: var(--badge-service-color); }

.no-resources {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
  padding: 8px 12px;
  background-color: var(--bg-surface-elevated);
  border-radius: var(--radius-sm);
}

.resource-items-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.resource-item {
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 10px 14px;
  min-width: 0;
}

.resource-item-top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  flex-wrap: wrap;
  min-width: 0;
}

.asset-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  background-color: var(--badge-workload-bg);
  color: var(--badge-workload-color);
  padding: 2px 8px;
  border-radius: 4px;
  flex-shrink: 0;
}

.asset-badge-service {
  background-color: var(--badge-service-bg);
  color: var(--badge-service-color);
}

.apphub-id-text {
  font-size: 12px;
  color: var(--text-secondary);
  word-break: break-all;
  overflow-wrap: anywhere;
  min-width: 0;
  flex: 1;
}
.apphub-id-text code {
  font-family: var(--font-mono);
  color: var(--text-primary);
  background-color: var(--bg-surface);
  border: 1px solid var(--border-color);
  padding: 1px 5px;
  border-radius: 3px;
  word-break: break-all;
  overflow-wrap: anywhere;
}

.resource-uri-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  min-width: 0;
}

.uri-label {
  color: var(--text-muted);
  font-weight: 600;
  flex-shrink: 0;
}

.uri-code {
  font-family: var(--font-mono);
  color: var(--text-secondary);
  word-break: break-all;
  overflow-wrap: anywhere;
  min-width: 0;
  flex: 1;
}

.copy-icon-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}
.copy-icon-btn:hover {
  color: var(--color-brand-text);
}
.copy-icon-btn .icon {
  font-size: 14px;
}

/* JSON View */
.json-view-box {
  background-color: var(--bg-code);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  min-width: 0;
}

.json-toolbar {
  padding: 10px 16px;
  background-color: var(--bg-surface-elevated);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.json-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.json-code {
  padding: 16px;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-code);
  max-height: 500px;
  overflow: auto;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Error Banner */
.error-banner {
  display: flex;
  gap: 14px;
  background-color: var(--color-error-bg);
  border: 1px solid var(--color-error);
  border-radius: var(--radius-md);
  padding: 16px;
  color: var(--color-error);
  min-width: 0;
}

.error-icon .icon {
  font-size: 24px;
  color: var(--color-error);
}

.error-content {
  min-width: 0;
  flex: 1;
}

.error-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-error);
}

.error-message {
  font-size: 13px;
  margin-top: 2px;
  font-family: var(--font-mono);
  word-break: break-word;
  overflow-wrap: anywhere;
}

/* Empty & Loading States */
.empty-state, .loading-state {
  text-align: center;
  padding: 48px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.empty-icon-box {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-lg);
  background-color: var(--color-brand-subtle);
  color: var(--color-brand);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}
.empty-icon-box .icon {
  font-size: 32px;
}

.empty-title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-desc {
  font-size: 13px;
  color: var(--text-muted);
  max-width: 460px;
  margin: 8px 0 24px;
  line-height: 1.6;
  word-break: break-word;
}

.quick-tips {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  justify-content: center;
}

.tip-card {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 8px 14px;
  font-size: 12px;
  color: var(--text-secondary);
}
.tip-card .icon {
  color: var(--color-warning);
  font-size: 16px;
}

.loading-spinner-large {
  width: 40px;
  height: 40px;
  border: 3px solid var(--color-brand-subtle);
  border-top-color: var(--color-brand);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 16px;
}

.loading-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.loading-subtext {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.no-matches {
  text-align: center;
  padding: 32px;
  background-color: var(--bg-surface-elevated);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}
.no-matches .icon {
  font-size: 32px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
