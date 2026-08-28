<script setup>
defineProps({
  serverStatus: {
    type: String,
    default: 'checking' // 'online', 'offline', 'checking'
  },
  serverUrl: {
    type: String,
    default: 'http://localhost:8080'
  },
  isJsonModalOpen: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['open-server-modal', 'open-json-modal', 'load-sample'])

const samples = [
  { id: 'sample1', name: 'Auto-Detect (Global + US)', desc: 'Discovers apps using standard well-known labels' },
  { id: 'sample2', name: 'Resource Label (env=prod)', desc: 'Groups assets by GCP resource label key/value' },
  { id: 'sample3', name: 'Resource Manager Tag', desc: 'Groups assets by GCP organization tag' },
  { id: 'sample4', name: 'Kubernetes Namespaces', desc: 'Generates one app per GKE namespace' },
  { id: 'sample5', name: 'Multi-Project Aggregator', desc: 'Aggregates assets across multiple GCP project IDs' }
]
</script>

<template>
  <header class="header-nav">
    <div class="container header-container">
      <div class="brand-area">
        <div class="logo-icon">
          <span class="icon">hub</span>
        </div>
        <div class="brand-text">
          <div class="brand-title">
            <span>Google Cloud</span> App Hub
            <span class="version-tag">Creator UI</span>
          </div>
          <div class="brand-subtitle">Automated Resource Discovery & App Hub Management</div>
        </div>
      </div>

      <div class="header-actions">
        <!-- Server Status Indicator -->
        <button 
          class="server-status-pill" 
          :class="'status-' + serverStatus"
          @click="emit('open-server-modal')"
          title="Click to configure backend server connection"
          id="btn-server-status"
        >
          <span class="status-dot"></span>
          <span class="status-text">
            {{ serverStatus === 'online' ? 'Backend Online' : (serverStatus === 'offline' ? 'Backend Offline' : 'Connecting...') }}
          </span>
          <span class="server-endpoint">{{ serverUrl }}</span>
        </button>

        <!-- Sample Load Dropdown -->
        <div class="sample-dropdown-wrapper">
          <button class="btn btn-secondary btn-sm" id="btn-samples-menu">
            <span class="icon">library_books</span>
            <span>Samples</span>
            <span class="icon" style="font-size: 16px;">arrow_drop_down</span>
          </button>
          <div class="sample-dropdown-menu">
            <div class="dropdown-header">Load Preset Scenario</div>
            <button 
              v-for="s in samples" 
              :key="s.id"
              class="dropdown-item"
              @click="emit('load-sample', s.id)"
            >
              <div class="item-title">{{ s.name }}</div>
              <div class="item-desc">{{ s.desc }}</div>
            </button>
          </div>
        </div>

        <!-- JSON Editor Toggle -->
        <button 
          class="btn btn-secondary btn-sm" 
          @click="emit('open-json-modal')"
          title="Open raw JSON request editor"
          id="btn-open-json-modal"
        >
          <span class="icon">data_object</span>
          <span>JSON Payload</span>
        </button>

        <!-- Docs Link -->
        <a 
          href="https://cloud.google.com/app-hub/docs/overview" 
          target="_blank" 
          rel="noopener noreferrer"
          class="btn btn-ghost btn-sm"
          title="Google Cloud App Hub Documentation"
        >
          <span class="icon">help_outline</span>
          <span>Docs</span>
        </a>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header-nav {
  background-color: var(--bg-surface);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 100;
  backdrop-filter: blur(8px);
}

.header-container {
  height: 68px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.brand-area {
  display: flex;
  align-items: center;
  gap: 14px;
}

.logo-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #1e40af 0%, #3b82f6 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}
.logo-icon .icon {
  font-size: 24px;
}

.brand-title {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-title span {
  color: #60a5fa;
  font-weight: 500;
}

.version-tag {
  background-color: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.brand-subtitle {
  font-size: 12px;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.server-status-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background-color: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.server-status-pill:hover {
  border-color: var(--border-focus);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  transition: background-color var(--transition-fast);
}

.status-online .status-dot {
  background-color: var(--color-success);
  box-shadow: 0 0 6px var(--color-success);
}
.status-offline .status-dot {
  background-color: var(--color-error);
  box-shadow: 0 0 6px var(--color-error);
}
.status-checking .status-dot {
  background-color: var(--color-warning);
  animation: pulse 1s infinite;
}

.status-text {
  font-weight: 600;
}
.status-online .status-text { color: var(--color-success); }
.status-offline .status-text { color: var(--color-error); }
.status-checking .status-text { color: var(--color-warning); }

.server-endpoint {
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 11px;
  padding-left: 4px;
  border-left: 1px solid var(--border-color);
}

/* Samples Dropdown */
.sample-dropdown-wrapper {
  position: relative;
}

.sample-dropdown-menu {
  display: none;
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 280px;
  background-color: var(--bg-surface-elevated);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  padding: 6px;
  z-index: 200;
}

.sample-dropdown-wrapper:hover .sample-dropdown-menu,
.sample-dropdown-wrapper:focus-within .sample-dropdown-menu {
  display: block;
}

.dropdown-header {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 6px 10px;
}

.dropdown-item {
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  padding: 8px 10px;
  cursor: pointer;
  color: var(--text-primary);
  transition: background-color var(--transition-fast);
}

.dropdown-item:hover {
  background-color: var(--bg-surface-hover);
}

.item-title {
  font-size: 13px;
  font-weight: 500;
  color: #93c5fd;
}

.item-desc {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

@media (max-width: 900px) {
  .brand-subtitle, .server-endpoint {
    display: none;
  }
}
</style>
