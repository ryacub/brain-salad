package analytics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRenderSparkline tests sparkline generation
func TestRenderSparkline(t *testing.T) {
	t.Run("basic sparkline", func(_ *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		result := RenderSparkline(values)

		assert.NotEmpty(t, result)
		assert.Equal(t, 5, len([]rune(result)), "should have 5 characters")
	})

	t.Run("empty values", func(_ *testing.T) {
		values := []float64{}
		result := RenderSparkline(values)

		assert.Empty(t, result)
	})

	t.Run("flat values", func(_ *testing.T) {
		values := []float64{5.0, 5.0, 5.0, 5.0}
		result := RenderSparkline(values)

		assert.NotEmpty(t, result)
		// All characters should be the same
		runes := []rune(result)
		for i := 1; i < len(runes); i++ {
			assert.Equal(t, runes[0], runes[i])
		}
	})

	t.Run("increasing trend", func(_ *testing.T) {
		values := []float64{1.0, 3.0, 5.0, 7.0, 9.0}
		result := RenderSparkline(values)

		assert.NotEmpty(t, result)
		assert.Equal(t, 5, len([]rune(result)))
	})
}

// TestRenderBarChart tests bar chart generation
func TestRenderBarChart(t *testing.T) {
	t.Run("basic bar chart", func(_ *testing.T) {
		labels := []string{"A", "B", "C"}
		values := []float64{5.0, 10.0, 7.5}
		result := RenderBarChart(labels, values, 20)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "A")
		assert.Contains(t, result, "B")
		assert.Contains(t, result, "C")
		assert.Contains(t, result, "█")
	})

	t.Run("empty inputs", func(_ *testing.T) {
		labels := []string{}
		values := []float64{}
		result := RenderBarChart(labels, values, 20)

		assert.Empty(t, result)
	})

	t.Run("mismatched lengths", func(_ *testing.T) {
		labels := []string{"A", "B"}
		values := []float64{5.0}
		result := RenderBarChart(labels, values, 20)

		assert.Empty(t, result)
	})

	t.Run("zero values", func(_ *testing.T) {
		labels := []string{"A", "B"}
		values := []float64{0.0, 0.0}
		result := RenderBarChart(labels, values, 20)

		assert.Empty(t, result)
	})
}

// TestRenderTrendChart tests trend chart generation
func TestRenderTrendChart(t *testing.T) {
	t.Run("basic trend chart", func(_ *testing.T) {
		trends := []TrendData{
			{Period: "2024-01", AvgScore: 6.0, IdeaCount: 5},
			{Period: "2024-02", AvgScore: 7.0, IdeaCount: 8},
			{Period: "2024-03", AvgScore: 8.0, IdeaCount: 10},
		}
		result := RenderTrendChart(trends, 5)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "│")
		assert.Contains(t, result, "●")
	})

	t.Run("empty trends", func(_ *testing.T) {
		trends := []TrendData{}
		result := RenderTrendChart(trends, 5)

		assert.Empty(t, result)
	})

	t.Run("flat trends", func(_ *testing.T) {
		trends := []TrendData{
			{Period: "2024-01", AvgScore: 7.0, IdeaCount: 5},
			{Period: "2024-02", AvgScore: 7.0, IdeaCount: 5},
			{Period: "2024-03", AvgScore: 7.0, IdeaCount: 5},
		}
		result := RenderTrendChart(trends, 5)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Flat trend at 7.0")
	})
}

// TestRenderDistribution tests distribution histogram
func TestRenderDistribution(t *testing.T) {
	t.Run("equal distribution", func(_ *testing.T) {
		result := RenderDistribution(10, 10, 10, 30)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
		assert.Contains(t, result, "█") // High
		assert.Contains(t, result, "▓") // Medium
		assert.Contains(t, result, "░") // Low
	})

	t.Run("all high", func(_ *testing.T) {
		result := RenderDistribution(30, 0, 0, 30)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "█")
		assert.Equal(t, 30, strings.Count(result, "█"))
	})

	t.Run("zero total", func(_ *testing.T) {
		result := RenderDistribution(0, 0, 0, 30)

		assert.Empty(t, result)
	})

	t.Run("mixed distribution", func(_ *testing.T) {
		result := RenderDistribution(5, 3, 2, 10)

		assert.NotEmpty(t, result)
		// Should have all three types
		assert.Contains(t, result, "█")
		assert.Contains(t, result, "▓")
		assert.Contains(t, result, "░")
	})
}

// TestRenderProgressBar tests progress bar generation
func TestRenderProgressBar(t *testing.T) {
	t.Run("50% progress", func(_ *testing.T) {
		result := RenderProgressBar(50, 100, 20)

		assert.NotEmpty(t, result)
		assert.Contains(t, result, "50%")
		assert.Contains(t, result, "█")
		assert.Contains(t, result, "░")
	})

	t.Run("100% progress", func(_ *testing.T) {
		result := RenderProgressBar(100, 100, 20)

		assert.Contains(t, result, "100%")
		assert.Contains(t, result, "█")
		assert.Equal(t, 20, strings.Count(result, "█"))
	})

	t.Run("0% progress", func(_ *testing.T) {
		result := RenderProgressBar(0, 100, 20)

		assert.Contains(t, result, "0%")
		assert.Equal(t, 20, strings.Count(result, "░"))
	})

	t.Run("zero total", func(_ *testing.T) {
		result := RenderProgressBar(10, 0, 20)

		assert.Contains(t, result, "0%")
	})

	t.Run("current exceeds total", func(_ *testing.T) {
		result := RenderProgressBar(150, 100, 20)

		assert.Contains(t, result, "150%")
		assert.Equal(t, 20, strings.Count(result, "█"))
	})
}
