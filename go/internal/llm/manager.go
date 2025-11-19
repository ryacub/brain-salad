package llm

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Manager handles multiple LLM providers with fallback, health checks, and statistics
type Manager struct {
	providers       []Provider
	primary         Provider
	fallbackEnabled bool
	mu              sync.RWMutex
	healthCache     map[string]healthStatus
	stats           map[string]*providerStats
	config          *ManagerConfig
}

// healthStatus tracks provider health information
type healthStatus struct {
	available bool
	lastCheck time.Time
	lastError error
}

// providerStats tracks statistics for a provider
type providerStats struct {
	totalRequests  int64
	successCount   int64
	failureCount   int64
	totalLatency   int64 // in nanoseconds
	lastUsed       time.Time
	mu             sync.RWMutex
}

// ManagerConfig contains configuration for the provider manager
type ManagerConfig struct {
	DefaultProvider     string
	FallbackEnabled     bool
	HealthCheckInterval time.Duration
	Priority            []string
	ProviderConfig      ProviderConfig
}

// DefaultManagerConfig returns the default manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		DefaultProvider:     "",
		FallbackEnabled:     true,
		HealthCheckInterval: 30 * time.Second,
		Priority:            []string{"ollama", "rule_based"},
		ProviderConfig:      DefaultProviderConfig(),
	}
}

// NewManager creates a provider manager with available providers
func NewManager(config *ManagerConfig) *Manager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	manager := &Manager{
		providers:       make([]Provider, 0),
		fallbackEnabled: config.FallbackEnabled,
		healthCache:     make(map[string]healthStatus),
		stats:           make(map[string]*providerStats),
		config:          config,
	}

	// Register available providers based on configuration
	manager.registerAvailableProviders()

	// Set primary provider based on configuration or availability
	if config.DefaultProvider != "" {
		_ = manager.SetPrimaryProvider(config.DefaultProvider)
	} else {
		manager.selectPrimaryProvider()
	}

	return manager
}

// registerAvailableProviders registers all configured providers
func (m *Manager) registerAvailableProviders() {
	// Register Ollama if configured
	if m.config.ProviderConfig.OllamaBaseURL != "" {
		ollama := NewOllamaProvider(
			m.config.ProviderConfig.OllamaBaseURL,
			m.config.ProviderConfig.OllamaModel,
		)
		m.RegisterProvider(ollama)
	}

	// Always register rule-based as fallback
	m.RegisterProvider(NewRuleBasedProvider())

	// Apply priority order if specified
	if len(m.config.Priority) > 0 {
		m.applyPriorityOrder(m.config.Priority)
	}
}

// RegisterProvider registers a new provider
func (m *Manager) RegisterProvider(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers = append(m.providers, p)
	m.stats[p.Name()] = &providerStats{}
	m.healthCache[p.Name()] = healthStatus{
		available: p.IsAvailable(),
		lastCheck: time.Now(),
	}
}

// selectPrimaryProvider selects the first available provider based on priority
func (m *Manager) selectPrimaryProvider() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.providers {
		if p.IsAvailable() {
			m.primary = p
			return
		}
	}
}

// Analyze performs analysis using the primary provider with fallback support
func (m *Manager) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	m.mu.RLock()
	primary := m.primary
	fallbackEnabled := m.fallbackEnabled
	m.mu.RUnlock()

	// Try primary provider
	if primary != nil {
		result, err := m.analyzeWithProvider(primary, req)
		if err == nil {
			return result, nil
		}
		// Log primary failure but continue to fallback
		fmt.Printf("[Manager] Primary provider %s failed: %v\n", primary.Name(), err)
	}

	// If fallback disabled, return error
	if !fallbackEnabled {
		return nil, fmt.Errorf("primary provider failed and fallback disabled")
	}

	// Fallback chain
	m.mu.RLock()
	providers := m.providers
	m.mu.RUnlock()

	var lastErr error
	for _, provider := range providers {
		// Skip primary (already tried)
		if primary != nil && provider.Name() == primary.Name() {
			continue
		}

		if !provider.IsAvailable() {
			continue
		}

		result, err := m.analyzeWithProvider(provider, req)
		if err == nil {
			fmt.Printf("[Manager] Fallback succeeded with provider: %s\n", provider.Name())
			return result, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// analyzeWithProvider performs analysis with a specific provider and tracks statistics
func (m *Manager) analyzeWithProvider(provider Provider, req AnalysisRequest) (*AnalysisResult, error) {
	start := time.Now()

	// Update stats - increment total requests
	m.updateStats(provider.Name(), func(stats *providerStats) {
		atomic.AddInt64(&stats.totalRequests, 1)
		stats.mu.Lock()
		stats.lastUsed = time.Now()
		stats.mu.Unlock()
	})

	// Perform analysis
	result, err := provider.Analyze(req)
	duration := time.Since(start)

	// Update stats based on result
	if err != nil {
		m.updateStats(provider.Name(), func(stats *providerStats) {
			atomic.AddInt64(&stats.failureCount, 1)
		})
		return nil, err
	}

	m.updateStats(provider.Name(), func(stats *providerStats) {
		atomic.AddInt64(&stats.successCount, 1)
		atomic.AddInt64(&stats.totalLatency, int64(duration))
	})

	return result, nil
}

// updateStats safely updates provider statistics
func (m *Manager) updateStats(providerName string, updateFunc func(*providerStats)) {
	m.mu.RLock()
	stats, exists := m.stats[providerName]
	m.mu.RUnlock()

	if exists && stats != nil {
		updateFunc(stats)
	}
}

// GetAvailableProviders returns all currently available providers
func (m *Manager) GetAvailableProviders() []Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := make([]Provider, 0)
	for _, p := range m.providers {
		if p.IsAvailable() {
			available = append(available, p)
		}
	}
	return available
}

// SetPrimaryProvider sets the primary provider by name
func (m *Manager) SetPrimaryProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.providers {
		if p.Name() == name {
			if !p.IsAvailable() {
				return fmt.Errorf("provider not available: %s", name)
			}
			m.primary = p
			return nil
		}
	}
	return fmt.Errorf("provider not found: %s", name)
}

// GetPrimaryProvider returns the current primary provider
func (m *Manager) GetPrimaryProvider() Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primary
}

// EnableFallback enables or disables fallback behavior
func (m *Manager) EnableFallback(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallbackEnabled = enabled
}

// IsFallbackEnabled returns whether fallback is enabled
func (m *Manager) IsFallbackEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fallbackEnabled
}

// HealthCheck performs health checks on all providers
func (m *Manager) HealthCheck() map[string]bool {
	m.mu.RLock()
	providers := m.providers
	m.mu.RUnlock()

	status := make(map[string]bool)
	for _, p := range providers {
		available := p.IsAvailable()
		status[p.Name()] = available

		// Update health cache
		m.mu.Lock()
		m.healthCache[p.Name()] = healthStatus{
			available: available,
			lastCheck: time.Now(),
		}
		m.mu.Unlock()
	}
	return status
}

// GetHealthStatus returns the health status for a specific provider
func (m *Manager) GetHealthStatus(providerName string) (bool, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health, exists := m.healthCache[providerName]
	if !exists {
		return false, time.Time{}, fmt.Errorf("provider not found: %s", providerName)
	}

	return health.available, health.lastCheck, nil
}

// StartPeriodicHealthCheck runs health checks in background
func (m *Manager) StartPeriodicHealthCheck(stopCh <-chan struct{}) {
	interval := m.config.HealthCheckInterval
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.HealthCheck()
		case <-stopCh:
			return
		}
	}
}

// ProviderStats contains statistics for a provider
type ProviderStats struct {
	Name           string
	Available      bool
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	AverageLatency time.Duration
	LastUsed       time.Time
}

// GetStats returns statistics for all providers
func (m *Manager) GetStats() []ProviderStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]ProviderStats, 0, len(m.providers))
	for _, p := range m.providers {
		providerStats, exists := m.stats[p.Name()]
		if !exists {
			continue
		}

		providerStats.mu.RLock()
		totalRequests := atomic.LoadInt64(&providerStats.totalRequests)
		successCount := atomic.LoadInt64(&providerStats.successCount)
		failureCount := atomic.LoadInt64(&providerStats.failureCount)
		totalLatency := atomic.LoadInt64(&providerStats.totalLatency)
		lastUsed := providerStats.lastUsed
		providerStats.mu.RUnlock()

		var avgLatency time.Duration
		if successCount > 0 {
			avgLatency = time.Duration(totalLatency / successCount)
		}

		stats = append(stats, ProviderStats{
			Name:           p.Name(),
			Available:      p.IsAvailable(),
			TotalRequests:  totalRequests,
			SuccessCount:   successCount,
			FailureCount:   failureCount,
			AverageLatency: avgLatency,
			LastUsed:       lastUsed,
		})
	}
	return stats
}

// GetProviderStats returns statistics for a specific provider
func (m *Manager) GetProviderStats(providerName string) (*ProviderStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providerStats, exists := m.stats[providerName]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	var provider Provider
	for _, p := range m.providers {
		if p.Name() == providerName {
			provider = p
			break
		}
	}

	if provider == nil {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	providerStats.mu.RLock()
	totalRequests := atomic.LoadInt64(&providerStats.totalRequests)
	successCount := atomic.LoadInt64(&providerStats.successCount)
	failureCount := atomic.LoadInt64(&providerStats.failureCount)
	totalLatency := atomic.LoadInt64(&providerStats.totalLatency)
	lastUsed := providerStats.lastUsed
	providerStats.mu.RUnlock()

	var avgLatency time.Duration
	if successCount > 0 {
		avgLatency = time.Duration(totalLatency / successCount)
	}

	return &ProviderStats{
		Name:           providerName,
		Available:      provider.IsAvailable(),
		TotalRequests:  totalRequests,
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		AverageLatency: avgLatency,
		LastUsed:       lastUsed,
	}, nil
}

// LoadConfig applies a new configuration to the manager
func (m *Manager) LoadConfig(config *ManagerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Set default provider
	if config.DefaultProvider != "" {
		if err := m.SetPrimaryProvider(config.DefaultProvider); err != nil {
			return fmt.Errorf("failed to set default provider: %w", err)
		}
	}

	// Set fallback
	m.EnableFallback(config.FallbackEnabled)

	// Apply priority order
	if len(config.Priority) > 0 {
		m.applyPriorityOrder(config.Priority)
	}

	// Update config
	m.mu.Lock()
	m.config = config
	m.mu.Unlock()

	return nil
}

// applyPriorityOrder reorders providers based on priority list
func (m *Manager) applyPriorityOrder(priority []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a map for quick lookup
	priorityMap := make(map[string]int)
	for i, name := range priority {
		priorityMap[name] = i
	}

	// Sort providers based on priority
	ordered := make([]Provider, 0, len(m.providers))

	// First, add providers in priority order
	for _, name := range priority {
		for _, p := range m.providers {
			if p.Name() == name {
				ordered = append(ordered, p)
				break
			}
		}
	}

	// Then add remaining providers not in priority list
	for _, p := range m.providers {
		found := false
		for _, op := range ordered {
			if op.Name() == p.Name() {
				found = true
				break
			}
		}
		if !found {
			ordered = append(ordered, p)
		}
	}

	m.providers = ordered
}

// GetProviders returns all registered providers
func (m *Manager) GetProviders() []Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	providers := make([]Provider, len(m.providers))
	copy(providers, m.providers)
	return providers
}

// ResetStats resets statistics for all providers
func (m *Manager) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, stats := range m.stats {
		atomic.StoreInt64(&stats.totalRequests, 0)
		atomic.StoreInt64(&stats.successCount, 0)
		atomic.StoreInt64(&stats.failureCount, 0)
		atomic.StoreInt64(&stats.totalLatency, 0)
		stats.mu.Lock()
		stats.lastUsed = time.Time{}
		stats.mu.Unlock()
	}
}

// CreateManagerWithTelos creates a manager configured with a specific telos
// This is a helper function for backward compatibility
func CreateManagerWithTelos(telos *models.Telos) *Manager {
	config := DefaultManagerConfig()
	manager := NewManager(config)
	return manager
}

// GetPrimaryProviderName returns the name of the current primary provider
// This is a helper method for CLI integration
func (m *Manager) GetPrimaryProviderName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.primary == nil {
		return ""
	}
	return m.primary.Name()
}

// GetAllProviders returns a map of all registered providers (available or not)
// This is a helper method for CLI integration
func (m *Manager) GetAllProviders() map[string]Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Convert slice to map for CLI compatibility
	result := make(map[string]Provider)
	for _, provider := range m.providers {
		result[provider.Name()] = provider
	}
	return result
}

// AnalyzeWithTelos is a helper that performs analysis with idea content and telos
// This is a convenience method for CLI integration
func (m *Manager) AnalyzeWithTelos(ideaContent string, telos *models.Telos) (*AnalysisResult, error) {
	req := AnalysisRequest{
		IdeaContent: ideaContent,
		Telos:       telos,
	}
	return m.Analyze(req)
}
