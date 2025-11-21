package analytics

import (
	"github.com/fatih/color"
)

// getScoreColor returns a color based on the score value
func getScoreColor(score float64) *color.Color {
	switch {
	case score >= 8.5:
		return color.New(color.FgGreen, color.Bold)
	case score >= 7.0:
		return color.New(color.FgGreen)
	case score >= 5.0:
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgRed)
	}
}
