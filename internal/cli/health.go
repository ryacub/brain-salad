package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/health"
	"github.com/spf13/cobra"
)

func newHealthCommand() *cobra.Command {
	var jsonFormat bool

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check system health status",
		Long: `Display health status of the Telos Matrix system.

This command checks:
- Database connectivity
- Memory usage
- Disk space availability
- System uptime

Examples:
  tm health          # Display health status with colored output
  tm health --json   # Display health status in JSON format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHealth(jsonFormat)
		},
	}

	cmd.Flags().BoolVar(&jsonFormat, "json", false, "Output health status in JSON format")

	return cmd
}

func runHealth(jsonFormat bool) error {
	// Create health monitor
	monitor := health.NewHealthMonitor()
	monitor.SetVersion("1.0.0")

	// Add database health checker
	monitor.AddCheck(health.NewDatabaseHealthChecker(ctx.Repository.DB()))

	// Add memory health checker (warn if using > 500MB)
	monitor.AddCheck(health.NewMemoryHealthChecker(500.0))

	// Add disk space health checker (warn if < 1GB free)
	monitor.AddCheck(health.NewDiskSpaceHealthChecker("/tmp", 1024))

	// Run all health checks with timeout
	checkCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status := monitor.RunAllChecks(checkCtx)

	if jsonFormat {
		return displayHealthJSON(status)
	}

	return displayHealthText(status)
}

func displayHealthJSON(status health.HealthStatus) error {
	// Use standard JSON encoding
	fmt.Printf(`{
  "status": "%s",
  "timestamp": "%s",
  "uptime": "%s",
  "version": "%s",
  "checks": [
`, status.Status, status.Timestamp.Format(time.RFC3339), status.Uptime, status.Version)

	for i, check := range status.Checks {
		comma := ","
		if i == len(status.Checks)-1 {
			comma = ""
		}
		fmt.Printf(`    {
      "name": "%s",
      "status": "%s",
      "message": "%s",
      "duration": "%s"
    }%s
`, check.Name, check.Status, check.Message, check.Duration, comma)
	}

	fmt.Println("  ]")
	fmt.Println("}")

	return nil
}

func displayHealthText(status health.HealthStatus) error {
	fmt.Println()
	fmt.Println("🏥 System Health Check")
	fmt.Println("═════════════════════════════════════════════")
	fmt.Println()

	// Display overall status
	statusIcon := ""
	var statusText string
	switch status.Status {
	case health.Healthy:
		statusIcon = "✅"
		statusText = successColor.Sprint("HEALTHY")
	case health.Degraded:
		statusIcon = "⚠️"
		statusText = warningColor.Sprint("DEGRADED")
	case health.Unhealthy:
		statusIcon = "❌"
		statusText = errorColor.Sprint("UNHEALTHY")
	default:
		statusIcon = "❓"
		statusText = infoColor.Sprint("UNKNOWN")
	}

	fmt.Printf("Overall Status: %s %s\n", statusIcon, statusText)
	fmt.Printf("Timestamp:      %s\n", status.Timestamp.Format(time.RFC3339))
	fmt.Printf("Uptime:         %s\n", formatDuration(status.Uptime))
	fmt.Printf("Version:        %s\n", status.Version)
	fmt.Println()

	// Display individual checks
	fmt.Println("Health Checks:")
	fmt.Println("─────────────────────────────────────────────")

	for _, check := range status.Checks {
		checkIcon := ""
		var checkStatus string

		switch check.Status {
		case health.Ok:
			checkIcon = "✅"
			checkStatus = successColor.Sprint("OK")
		case health.Warning:
			checkIcon = "⚠️"
			checkStatus = warningColor.Sprint("WARNING")
		case health.Error:
			checkIcon = "❌"
			checkStatus = errorColor.Sprint("ERROR")
		default:
			checkIcon = "❓"
			checkStatus = infoColor.Sprint("UNKNOWN")
		}

		fmt.Printf("\n%s %s: %s\n", checkIcon, check.Name, checkStatus)
		if check.Message != "" {
			if _, err := infoColor.Printf("  └─ %s\n", check.Message); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
		fmt.Printf("  └─ Duration: %s\n", check.Duration)
	}

	fmt.Println()
	fmt.Println("═════════════════════════════════════════════")

	// Return error if unhealthy (for CI/CD integration)
	if status.Status == health.Unhealthy {
		fmt.Println()
		if _, err := errorColor.Println("⚠️  System is unhealthy. Please address the issues above."); err != nil {
			log.Warn().Err(err).Msg("failed to print error message")
		}
		return fmt.Errorf("system health check failed")
	}

	if status.Status == health.Degraded {
		fmt.Println()
		if _, err := warningColor.Println("⚠️  System is degraded. Some components may not be operating optimally."); err != nil {
			log.Warn().Err(err).Msg("failed to print warning")
		}
	}

	return nil
}

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-d.Minutes()*60)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) - hours*60
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
