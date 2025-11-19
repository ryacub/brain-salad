package llm

import (
	"errors"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// mockProviderForManager is a mock provider for manager testing
// (separate from provider_test.go's MockProvider to avoid conflicts)
type mockProviderForManager struct {
	name      string
	available bool
	err       error
	result    *AnalysisResult
	callCount int
}

func (m *mockProviderForManager) Name() string {
	return m.name
}

func (m *mockProviderForManager) IsAvailable() bool {
	return m.available
}

func (m *mockProviderForManager) Analyze(req AnalysisRequest) (*AnalysisResult, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	if m.result == nil {
		return &AnalysisResult{
			Scores: ScoreBreakdown{
				MissionAlignment: 3.0,
				AntiChallenge:    2.5,
				StrategicFit:     2.0,
			},
			FinalScore:     7.5,
			Recommendation: "Test recommendation",
			Provider:       m.name,
		}, nil
	}
	return m.result, nil
}

// createTestTelos creates a minimal valid telos for testing
func createTestTelos() *models.Telos {
	return &models.Telos{
		Goals: []models.Goal{
			{
				ID:          "test-goal",
				Description: "Test goal",
				Priority:    1,
			},
		},
		Missions: []models.Mission{
			{
				ID:          "test-mission",
				Description: "Test mission",
			},
		},
	}
}

func TestNewManager(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	if manager == nil {
		t.Fatal("Expected manager, got nil")
	}

	if !manager.IsFallbackEnabled() {
		t.Error("Expected fallback to be enabled by default")
	}

	providers := manager.GetProviders()
	if len(providers) == 0 {
		t.Error("Expected at least one provider (rule-based)")
	}
}

func TestManager_RegisterProvider(t *testing.T) {
	config := &ManagerConfig{
		FallbackEnabled: true,
		Priority:        []string{},
		ProviderConfig:  DefaultProviderConfig(),
	}
	manager := NewManager(config)

	// Register a mock provider
	mockProvider := &mockProviderForManager{
		name:      "mock",
		available: true,
	}
	manager.RegisterProvider(mockProvider)

	// Check if provider was registered
	providers := manager.GetProviders()
	found := false
	for _, p := range providers {
		if p.Name() == "mock" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Mock provider was not registered")
	}
}

func TestManager_FallbackChain(t *testing.T) {
	// Create a manager without default providers for precise control
	config := &ManagerConfig{
		FallbackEnabled: true,
		Priority:        []string{"primary", "fallback"},
		ProviderConfig:  ProviderConfig{}, // Empty config to avoid auto-registration
	}
	manager := &Manager{
		providers:       make([]Provider, 0),
		fallbackEnabled: config.FallbackEnabled,
		healthCache:     make(map[string]healthStatus),
		stats:           make(map[string]*providerStats),
		config:          config,
	}

	// Create mock providers
	primaryProvider := &mockProviderForManager{
		name:      "primary",
		available: true,
		err:       errors.New("primary failed"),
	}

	fallbackProvider := &mockProviderForManager{
		name:      "fallback",
		available: true,
	}

	manager.RegisterProvider(primaryProvider)
	manager.RegisterProvider(fallbackProvider)

	// Set primary provider
	err := manager.SetPrimaryProvider("primary")
	if err != nil {
		t.Fatalf("Failed to set primary provider: %v", err)
	}

	// Create a test telos
	telos := createTestTelos()

	// Perform analysis - should fall back to second provider
	result, err := manager.Analyze(AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	})

	if err != nil {
		t.Fatalf("Expected successful analysis with fallback, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify primary was called
	if primaryProvider.callCount != 1 {
		t.Errorf("Expected primary provider to be called once, got %d", primaryProvider.callCount)
	}

	// Verify fallback was called
	if fallbackProvider.callCount != 1 {
		t.Errorf("Expected fallback provider to be called once, got %d", fallbackProvider.callCount)
	}
}

func TestManager_FallbackDisabled(t *testing.T) {
	config := &ManagerConfig{
		FallbackEnabled: false,
		Priority:        []string{"primary"},
		ProviderConfig:  DefaultProviderConfig(),
	}
	manager := NewManager(config)

	// Create mock provider that fails
	primaryProvider := &mockProviderForManager{
		name:      "primary",
		available: true,
		err:       errors.New("primary failed"),
	}

	manager.RegisterProvider(primaryProvider)
	err := manager.SetPrimaryProvider("primary")
	if err != nil {
		t.Fatalf("Failed to set primary provider: %v", err)
	}

	telos := createTestTelos()

	// Perform analysis - should fail without fallback
	_, err = manager.Analyze(AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	})

	if err == nil {
		t.Error("Expected error when fallback is disabled and primary fails")
	}
}

func TestManager_SetPrimaryProvider(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Register mock providers
	provider1 := &mockProviderForManager{
		name:      "provider1",
		available: true,
	}
	provider2 := &mockProviderForManager{
		name:      "provider2",
		available: true,
	}

	manager.RegisterProvider(provider1)
	manager.RegisterProvider(provider2)

	// Set primary to provider1
	err := manager.SetPrimaryProvider("provider1")
	if err != nil {
		t.Fatalf("Failed to set primary provider: %v", err)
	}

	primary := manager.GetPrimaryProvider()
	if primary.Name() != "provider1" {
		t.Errorf("Expected 'provider1', got '%s'", primary.Name())
	}

	// Set primary to provider2
	err = manager.SetPrimaryProvider("provider2")
	if err != nil {
		t.Fatalf("Failed to set primary provider: %v", err)
	}

	primary = manager.GetPrimaryProvider()
	if primary.Name() != "provider2" {
		t.Errorf("Expected 'provider2', got '%s'", primary.Name())
	}

	// Try to set non-existent provider
	err = manager.SetPrimaryProvider("nonexistent")
	if err == nil {
		t.Error("Expected error when setting non-existent provider")
	}
}

func TestManager_SetPrimaryProviderUnavailable(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Register unavailable provider
	provider := &mockProviderForManager{
		name:      "unavailable",
		available: false,
	}

	manager.RegisterProvider(provider)

	// Try to set unavailable provider as primary
	err := manager.SetPrimaryProvider("unavailable")
	if err == nil {
		t.Error("Expected error when setting unavailable provider as primary")
	}
}

func TestManager_HealthCheck(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Register mock providers
	availableProvider := &mockProviderForManager{
		name:      "available",
		available: true,
	}
	unavailableProvider := &mockProviderForManager{
		name:      "unavailable",
		available: false,
	}

	manager.RegisterProvider(availableProvider)
	manager.RegisterProvider(unavailableProvider)

	// Perform health check
	status := manager.HealthCheck()

	if !status["available"] {
		t.Error("Expected 'available' provider to be available")
	}

	if status["unavailable"] {
		t.Error("Expected 'unavailable' provider to be unavailable")
	}
}

func TestManager_GetHealthStatus(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
	}
	manager.RegisterProvider(provider)

	// Perform health check to populate cache
	manager.HealthCheck()

	// Get health status
	available, lastCheck, err := manager.GetHealthStatus("test")
	if err != nil {
		t.Fatalf("Failed to get health status: %v", err)
	}

	if !available {
		t.Error("Expected provider to be available")
	}

	if lastCheck.IsZero() {
		t.Error("Expected lastCheck to be set")
	}

	// Try to get status for non-existent provider
	_, _, err = manager.GetHealthStatus("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent provider")
	}
}

func TestManager_GetAvailableProviders(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Register mock providers
	available1 := &mockProviderForManager{
		name:      "available1",
		available: true,
	}
	available2 := &mockProviderForManager{
		name:      "available2",
		available: true,
	}
	unavailable := &mockProviderForManager{
		name:      "unavailable",
		available: false,
	}

	manager.RegisterProvider(available1)
	manager.RegisterProvider(available2)
	manager.RegisterProvider(unavailable)

	// Get available providers
	providers := manager.GetAvailableProviders()

	// Should have at least 2 available providers (not counting any default ones)
	foundAvailable1 := false
	foundAvailable2 := false
	foundUnavailable := false

	for _, p := range providers {
		switch p.Name() {
		case "available1":
			foundAvailable1 = true
		case "available2":
			foundAvailable2 = true
		case "unavailable":
			foundUnavailable = true
		}
	}

	if !foundAvailable1 || !foundAvailable2 {
		t.Error("Expected both available providers to be returned")
	}

	if foundUnavailable {
		t.Error("Did not expect unavailable provider to be returned")
	}
}

func TestManager_Stats(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
	}
	manager.RegisterProvider(provider)
	manager.SetPrimaryProvider("test")

	telos := createTestTelos()

	// Perform multiple analyses
	for i := 0; i < 3; i++ {
		_, err := manager.Analyze(AnalysisRequest{
			IdeaContent: "Test idea",
			Telos:       telos,
		})
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}
	}

	// Get stats
	stats, err := manager.GetProviderStats("test")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", stats.TotalRequests)
	}

	if stats.SuccessCount != 3 {
		t.Errorf("Expected 3 successful requests, got %d", stats.SuccessCount)
	}

	if stats.FailureCount != 0 {
		t.Errorf("Expected 0 failed requests, got %d", stats.FailureCount)
	}
}

func TestManager_StatsWithFailures(t *testing.T) {
	config := &ManagerConfig{
		FallbackEnabled: false, // Disable fallback to capture failures
		Priority:        []string{"test"},
		ProviderConfig:  DefaultProviderConfig(),
	}
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
		err:       errors.New("test error"),
	}
	manager.RegisterProvider(provider)
	manager.SetPrimaryProvider("test")

	telos := createTestTelos()

	// Perform analysis that will fail
	_, err := manager.Analyze(AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	})

	if err == nil {
		t.Fatal("Expected analysis to fail")
	}

	// Get stats
	stats, err := manager.GetProviderStats("test")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", stats.TotalRequests)
	}

	if stats.FailureCount != 1 {
		t.Errorf("Expected 1 failed request, got %d", stats.FailureCount)
	}
}

func TestManager_GetStats(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	provider1 := &mockProviderForManager{
		name:      "provider1",
		available: true,
	}
	provider2 := &mockProviderForManager{
		name:      "provider2",
		available: true,
	}

	manager.RegisterProvider(provider1)
	manager.RegisterProvider(provider2)

	// Get all stats
	allStats := manager.GetStats()

	// Should have at least our 2 providers
	if len(allStats) < 2 {
		t.Errorf("Expected at least 2 providers in stats, got %d", len(allStats))
	}
}

func TestManager_ResetStats(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
	}
	manager.RegisterProvider(provider)
	manager.SetPrimaryProvider("test")

	telos := createTestTelos()

	// Perform an analysis
	_, err := manager.Analyze(AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	})
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	// Verify stats are non-zero
	stats, _ := manager.GetProviderStats("test")
	if stats.TotalRequests == 0 {
		t.Error("Expected non-zero stats before reset")
	}

	// Reset stats
	manager.ResetStats()

	// Verify stats are zero
	stats, _ = manager.GetProviderStats("test")
	if stats.TotalRequests != 0 {
		t.Errorf("Expected zero requests after reset, got %d", stats.TotalRequests)
	}
}

func TestManager_ApplyPriorityOrder(t *testing.T) {
	config := &ManagerConfig{
		FallbackEnabled: true,
		Priority:        []string{"provider2", "provider1", "provider3"},
		ProviderConfig:  DefaultProviderConfig(),
	}
	manager := NewManager(config)

	provider1 := &mockProviderForManager{name: "provider1", available: true}
	provider2 := &mockProviderForManager{name: "provider2", available: true}
	provider3 := &mockProviderForManager{name: "provider3", available: true}

	manager.RegisterProvider(provider1)
	manager.RegisterProvider(provider2)
	manager.RegisterProvider(provider3)

	// Apply priority order
	manager.applyPriorityOrder(config.Priority)

	providers := manager.GetProviders()

	// Check order (accounting for potentially existing providers)
	var foundOrder []string
	for _, p := range providers {
		if p.Name() == "provider1" || p.Name() == "provider2" || p.Name() == "provider3" {
			foundOrder = append(foundOrder, p.Name())
		}
	}

	expectedOrder := []string{"provider2", "provider1", "provider3"}
	for i, name := range expectedOrder {
		if i >= len(foundOrder) || foundOrder[i] != name {
			t.Errorf("Expected provider %s at position %d, got %v", name, i, foundOrder)
		}
	}
}

func TestManager_LoadConfig(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
	}
	manager.RegisterProvider(provider)

	// Load new config
	newConfig := &ManagerConfig{
		DefaultProvider: "test",
		FallbackEnabled: false,
		Priority:        []string{"test"},
		ProviderConfig:  DefaultProviderConfig(),
	}

	err := manager.LoadConfig(newConfig)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config was applied
	primary := manager.GetPrimaryProvider()
	if primary.Name() != "test" {
		t.Errorf("Expected primary provider 'test', got '%s'", primary.Name())
	}

	if manager.IsFallbackEnabled() {
		t.Error("Expected fallback to be disabled")
	}
}

func TestManager_LoadConfigInvalidProvider(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Try to load config with non-existent provider
	newConfig := &ManagerConfig{
		DefaultProvider: "nonexistent",
		FallbackEnabled: true,
		ProviderConfig:  DefaultProviderConfig(),
	}

	err := manager.LoadConfig(newConfig)
	if err == nil {
		t.Error("Expected error when loading config with invalid default provider")
	}
}

func TestManager_EnableFallback(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Disable fallback
	manager.EnableFallback(false)
	if manager.IsFallbackEnabled() {
		t.Error("Expected fallback to be disabled")
	}

	// Enable fallback
	manager.EnableFallback(true)
	if !manager.IsFallbackEnabled() {
		t.Error("Expected fallback to be enabled")
	}
}

func TestManager_PeriodicHealthCheck(t *testing.T) {
	config := &ManagerConfig{
		FallbackEnabled:     true,
		HealthCheckInterval: 100 * time.Millisecond,
		ProviderConfig:      DefaultProviderConfig(),
	}
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
	}
	manager.RegisterProvider(provider)

	// Start periodic health check
	stopCh := make(chan struct{})
	go manager.StartPeriodicHealthCheck(stopCh)

	// Wait for a couple of health checks
	time.Sleep(250 * time.Millisecond)

	// Stop health check
	close(stopCh)

	// Verify health check was performed
	available, lastCheck, err := manager.GetHealthStatus("test")
	if err != nil {
		t.Fatalf("Failed to get health status: %v", err)
	}

	if !available {
		t.Error("Expected provider to be available")
	}

	if lastCheck.IsZero() {
		t.Error("Expected lastCheck to be set")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	config := DefaultManagerConfig()
	manager := NewManager(config)

	provider := &mockProviderForManager{
		name:      "test",
		available: true,
	}
	manager.RegisterProvider(provider)
	manager.SetPrimaryProvider("test")

	telos := createTestTelos()

	// Run concurrent analyses
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := manager.Analyze(AnalysisRequest{
				IdeaContent: "Test idea",
				Telos:       telos,
			})
			if err != nil {
				t.Errorf("Analysis failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify stats
	stats, err := manager.GetProviderStats("test")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalRequests != 10 {
		t.Errorf("Expected 10 requests, got %d", stats.TotalRequests)
	}
}

func TestCreateManagerWithTelos(t *testing.T) {
	telos := createTestTelos()

	manager := CreateManagerWithTelos(telos)

	if manager == nil {
		t.Fatal("Expected manager, got nil")
	}

	if !manager.IsFallbackEnabled() {
		t.Error("Expected fallback to be enabled")
	}
}
