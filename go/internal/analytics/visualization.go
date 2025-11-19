package analytics

import (
	"fmt"
	"math"
	"strings"
)

// RenderSparkline generates a simple ASCII sparkline chart
func RenderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	// Sparkline characters from lowest to highest
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Find min and max
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Handle case where all values are the same
	if min == max {
		return strings.Repeat(string(chars[len(chars)/2]), len(values))
	}

	// Map values to characters
	var result strings.Builder
	for _, v := range values {
		normalized := (v - min) / (max - min)
		idx := int(normalized * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		result.WriteRune(chars[idx])
	}

	return result.String()
}

// RenderBarChart generates a simple ASCII bar chart
func RenderBarChart(labels []string, values []float64, maxWidth int) string {
	if len(labels) == 0 || len(values) == 0 || len(labels) != len(values) {
		return ""
	}

	// Find max value
	maxValue := values[0]
	for _, v := range values {
		if v > maxValue {
			maxValue = v
		}
	}

	if maxValue == 0 {
		return ""
	}

	// Find max label width for alignment
	maxLabelWidth := 0
	for _, label := range labels {
		if len(label) > maxLabelWidth {
			maxLabelWidth = len(label)
		}
	}

	var result strings.Builder

	for i, label := range labels {
		value := values[i]

		// Calculate bar width
		barWidth := int((value / maxValue) * float64(maxWidth))
		if barWidth < 1 && value > 0 {
			barWidth = 1
		}

		// Create bar
		bar := strings.Repeat("█", barWidth)

		// Format line
		result.WriteString(fmt.Sprintf("%-*s │ %-*s %.1f\n",
			maxLabelWidth, label,
			maxWidth, bar,
			value))
	}

	return result.String()
}

// RenderTrendChart generates an ASCII line chart for trend data
func RenderTrendChart(trends []TrendData, height int) string {
	if len(trends) == 0 {
		return ""
	}

	// Extract values
	values := make([]float64, len(trends))
	for i, trend := range trends {
		values[i] = trend.AvgScore
	}

	// Find min and max
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Handle case where all values are the same
	if min == max {
		return fmt.Sprintf("Flat trend at %.1f\n", min)
	}

	// Create chart grid
	var chart strings.Builder

	// Y-axis labels and grid
	for row := height - 1; row >= 0; row-- {
		// Calculate value for this row
		yValue := min + (float64(row)/float64(height-1))*(max-min)

		// Y-axis label
		chart.WriteString(fmt.Sprintf("%4.1f │", yValue))

		// Plot points
		for col := 0; col < len(values); col++ {
			normalized := (values[col] - min) / (max - min)
			pointRow := int(normalized * float64(height-1))

			if pointRow == row {
				chart.WriteString("●")
			} else {
				chart.WriteString(" ")
			}
		}
		chart.WriteString("\n")
	}

	// X-axis
	chart.WriteString("     └")
	chart.WriteString(strings.Repeat("─", len(values)))
	chart.WriteString("\n      ")

	// X-axis labels (show first, middle, last)
	for i := range trends {
		if i == 0 || i == len(trends)/2 || i == len(trends)-1 {
			chart.WriteString("│")
		} else {
			chart.WriteString(" ")
		}
	}
	chart.WriteString("\n")

	return chart.String()
}

// RenderDistribution creates a simple distribution histogram
func RenderDistribution(high, medium, low int, width int) string {
	total := high + medium + low
	if total == 0 {
		return ""
	}

	// Calculate widths
	highWidth := int(math.Round(float64(high) / float64(total) * float64(width)))
	mediumWidth := int(math.Round(float64(medium) / float64(total) * float64(width)))
	lowWidth := width - highWidth - mediumWidth // Ensure total equals width

	// Ensure at least 1 char if count > 0
	if high > 0 && highWidth == 0 {
		highWidth = 1
	}
	if medium > 0 && mediumWidth == 0 {
		mediumWidth = 1
	}
	if low > 0 && lowWidth == 0 {
		lowWidth = 1
	}

	// Build distribution bar
	var bar strings.Builder
	bar.WriteString("[")
	bar.WriteString(strings.Repeat("█", highWidth))    // High
	bar.WriteString(strings.Repeat("▓", mediumWidth))  // Medium
	bar.WriteString(strings.Repeat("░", lowWidth))     // Low
	bar.WriteString("]")

	return bar.String()
}

// RenderProgressBar creates a simple progress/percentage bar
func RenderProgressBar(current, total int, width int) string {
	if total == 0 {
		return "[" + strings.Repeat(" ", width) + "] 0%"
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))

	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %d%%", bar, int(percentage*100))
}
