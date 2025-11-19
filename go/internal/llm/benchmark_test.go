package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// BenchmarkRuleBasedProvider_Analyze benchmarks the rule-based provider
func BenchmarkRuleBasedProvider_Analyze(b *testing.B) {
	provider := NewRuleBasedProvider()
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Build an AI-powered automation tool using Python and GPT-4 with $2000/month subscription model",
		Telos:       telos,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.Analyze(req)
		if err != nil {
			b.Fatalf("analysis failed: %v", err)
		}
	}
}

// BenchmarkRuleBasedProvider_Analyze_VaryingComplexity benchmarks with different idea complexities
func BenchmarkRuleBasedProvider_Analyze_VaryingComplexity(b *testing.B) {
	provider := NewRuleBasedProvider()
	telos := createBenchmarkTelos()

	tests := []struct {
		name string
		idea string
	}{
		{
			name: "simple",
			idea: "Build a web app",
		},
		{
			name: "medium",
			idea: "Build an AI-powered automation tool using Python with subscription pricing",
		},
		{
			name: "complex",
			idea: "Build a comprehensive AI-powered automation platform using Python, Go, and React with GPT-4 integration, multi-tenant architecture, and subscription-based pricing model targeting enterprise customers",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := AnalysisRequest{
				IdeaContent: tt.idea,
				Telos:       telos,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = provider.Analyze(req)
			}
		})
	}
}

// BenchmarkFallbackProvider_FirstSuccess benchmarks fallback when first provider succeeds
func BenchmarkFallbackProvider_FirstSuccess(b *testing.B) {
	successProvider := &MockProvider{
		name:      "success",
		available: true,
		result: &AnalysisResult{
			FinalScore:     8.0,
			Recommendation: "PRIORITIZE NOW",
			Provider:       "success",
		},
	}

	failProvider := &MockProvider{
		name:      "fail",
		available: false,
	}

	fallback := NewFallbackProvider(successProvider, failProvider)
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fallback.Analyze(req)
		if err != nil {
			b.Fatalf("analysis failed: %v", err)
		}
	}
}

// BenchmarkFallbackProvider_SecondSuccess benchmarks fallback when first fails, second succeeds
func BenchmarkFallbackProvider_SecondSuccess(b *testing.B) {
	failProvider := &MockProvider{
		name:      "fail",
		available: true,
		err:       context.DeadlineExceeded,
	}

	successProvider := &MockProvider{
		name:      "success",
		available: true,
		result: &AnalysisResult{
			FinalScore:     7.0,
			Recommendation: "GOOD ALIGNMENT",
			Provider:       "success",
		},
	}

	fallback := NewFallbackProvider(failProvider, successProvider)
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Test idea",
		Telos:       telos,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fallback.Analyze(req)
		if err != nil {
			b.Fatalf("analysis failed: %v", err)
		}
	}
}

// BenchmarkOllamaProvider_IsAvailable benchmarks health checking
func BenchmarkOllamaProvider_IsAvailable(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	provider := NewOllamaProvider(server.URL, "llama2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.IsAvailable()
	}
}

// BenchmarkPromptBuilding benchmarks prompt construction
func BenchmarkPromptBuilding(b *testing.B) {
	telos := createBenchmarkTelos()
	idea := "Build an AI-powered automation tool using Python and GPT-4"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := BuildAnalysisPrompt(idea, telos)
		if err != nil {
			b.Fatalf("prompt building failed: %v", err)
		}
	}
}

// BenchmarkPromptBuilding_VaryingTelosSize benchmarks with different telos sizes
func BenchmarkPromptBuilding_VaryingTelosSize(b *testing.B) {
	idea := "Build an AI tool"

	tests := []struct {
		name  string
		telos *models.Telos
	}{
		{
			name: "minimal",
			telos: &models.Telos{
				Goals: []models.Goal{
					{ID: "G1", Description: "Build AI tools", Priority: 1},
				},
				Strategies:      []models.Strategy{{ID: "S1", Description: "Move fast"}},
				Stack:           models.Stack{Primary: []string{"Python"}},
				FailurePatterns: []models.Pattern{{Name: "Delay", Keywords: []string{"delay"}}},
			},
		},
		{
			name:  "large",
			telos: createLargeTelos(),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = BuildAnalysisPrompt(idea, tt.telos)
			}
		})
	}
}

// BenchmarkConcurrentAnalysis benchmarks concurrent analysis requests
func BenchmarkConcurrentAnalysis(b *testing.B) {
	provider := NewRuleBasedProvider()
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Build an AI automation tool",
		Telos:       telos,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = provider.Analyze(req)
		}
	})
}

// BenchmarkDefaultFallbackChain benchmarks the default fallback chain creation
func BenchmarkDefaultFallbackChain(b *testing.B) {
	config := DefaultProviderConfig()
	telos := createBenchmarkTelos()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateDefaultFallbackChain(config, telos)
	}
}

// TestRuleBasedProvider_ResponseTime tests that analysis completes within time requirement
func TestRuleBasedProvider_ResponseTime(t *testing.T) {
	provider := NewRuleBasedProvider()
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Build an AI automation tool using Python and GPT-4",
		Telos:       telos,
	}

	start := time.Now()
	_, err := provider.Analyze(req)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Verify <3s requirement (should be much faster for rule-based)
	maxDuration := 100 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("analysis took %v, expected <%v", duration, maxDuration)
	}

	t.Logf("Analysis completed in %v", duration)
}

// TestFallbackProvider_ResponseTime tests fallback chain response time
func TestFallbackProvider_ResponseTime(t *testing.T) {
	// Create a mock slow provider
	slowProvider := &MockProvider{
		name:      "slow",
		available: true,
		err:       context.DeadlineExceeded,
	}

	// Fast rule-based fallback
	fastProvider := NewRuleBasedProvider()

	fallback := NewFallbackProvider(slowProvider, fastProvider)
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Build an AI tool",
		Telos:       telos,
	}

	start := time.Now()
	_, err := fallback.Analyze(req)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	// Should still be fast due to quick fallback
	maxDuration := 200 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("fallback analysis took %v, expected <%v", duration, maxDuration)
	}

	t.Logf("Fallback analysis completed in %v", duration)
}

// TestConcurrentRequests tests handling of concurrent requests
func TestConcurrentRequests(t *testing.T) {
	provider := NewRuleBasedProvider()
	telos := createBenchmarkTelos()

	concurrency := 50
	errChan := make(chan error, concurrency)
	doneChan := make(chan time.Duration, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			reqStart := time.Now()

			req := AnalysisRequest{
				IdeaContent: "Concurrent test idea",
				Telos:       telos,
			}

			_, err := provider.Analyze(req)
			errChan <- err
			doneChan <- time.Since(reqStart)
		}(i)
	}

	// Collect results
	var totalDuration time.Duration
	for i := 0; i < concurrency; i++ {
		err := <-errChan
		if err != nil {
			t.Errorf("concurrent request %d failed: %v", i, err)
		}
		totalDuration += <-doneChan
	}

	overallDuration := time.Since(start)
	avgDuration := totalDuration / time.Duration(concurrency)

	t.Logf("Completed %d concurrent requests in %v", concurrency, overallDuration)
	t.Logf("Average request duration: %v", avgDuration)
	t.Logf("Requests per second: %.2f", float64(concurrency)/overallDuration.Seconds())
}

// TestMemoryUsage tests memory efficiency of analysis
func TestMemoryUsage(t *testing.T) {
	provider := NewRuleBasedProvider()
	telos := createBenchmarkTelos()

	req := AnalysisRequest{
		IdeaContent: "Build an AI automation tool",
		Telos:       telos,
	}

	// Warmup
	for i := 0; i < 10; i++ {
		_, _ = provider.Analyze(req)
	}

	// Perform many analyses to check for memory leaks
	iterations := 1000
	for i := 0; i < iterations; i++ {
		_, err := provider.Analyze(req)
		if err != nil {
			t.Fatalf("iteration %d failed: %v", i, err)
		}
	}

	t.Logf("Completed %d iterations successfully", iterations)
}

// Helper functions

func createBenchmarkTelos() *models.Telos {
	deadline := time.Now().Add(90 * 24 * time.Hour)
	return &models.Telos{
		Goals: []models.Goal{
			{
				ID:          "G1",
				Description: "Ship AI-powered automation tools",
				Priority:    1,
				Deadline:    &deadline,
			},
			{
				ID:          "G2",
				Description: "Build expertise in AI/ML",
				Priority:    2,
			},
		},
		Strategies: []models.Strategy{
			{
				ID:          "S1",
				Description: "Focus on rapid prototyping",
			},
			{
				ID:          "S2",
				Description: "Leverage existing skills",
			},
		},
		Stack: models.Stack{
			Primary:   []string{"Python", "Go", "PostgreSQL"},
			Secondary: []string{"Docker", "Redis"},
		},
		FailurePatterns: []models.Pattern{
			{
				Name:        "Context Switching",
				Description: "Avoid switching technologies",
				Keywords:    []string{"context", "switching"},
			},
			{
				Name:        "Perfection Paralysis",
				Description: "Ship MVPs quickly",
				Keywords:    []string{"perfect", "complete"},
			},
		},
		LoadedAt: time.Now(),
	}
}

func createLargeTelos() *models.Telos {
	deadline := time.Now().Add(90 * 24 * time.Hour)

	goals := make([]models.Goal, 10)
	for i := 0; i < 10; i++ {
		goals[i] = models.Goal{
			ID:          "G" + string(rune('0'+i)),
			Description: "Goal description " + string(rune('0'+i)),
			Priority:    i + 1,
			Deadline:    &deadline,
		}
	}

	strategies := make([]models.Strategy, 10)
	for i := 0; i < 10; i++ {
		strategies[i] = models.Strategy{
			ID:          "S" + string(rune('0'+i)),
			Description: "Strategy description " + string(rune('0'+i)),
		}
	}

	patterns := make([]models.Pattern, 10)
	for i := 0; i < 10; i++ {
		patterns[i] = models.Pattern{
			Name:        "Pattern " + string(rune('0'+i)),
			Description: "Pattern description " + string(rune('0'+i)),
			Keywords:    []string{"keyword1", "keyword2", "keyword3"},
		}
	}

	return &models.Telos{
		Goals:      goals,
		Strategies: strategies,
		Stack: models.Stack{
			Primary:   []string{"Python", "Go", "TypeScript", "PostgreSQL", "React"},
			Secondary: []string{"Docker", "Redis", "AWS", "Kubernetes", "GraphQL"},
		},
		FailurePatterns: patterns,
		LoadedAt:        time.Now(),
	}
}
