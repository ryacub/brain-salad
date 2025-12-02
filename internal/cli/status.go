package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/config"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/profile"
	"github.com/spf13/cobra"
)

// Status constants
const (
	statusHealthy = "healthy"
	statusWarning = "warning"
	statusError   = "error"
)

func newStatusCommand() *cobra.Command {
	var jsonOutput bool
	var watch bool
	var interval int

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system status",
		Long: `Show system health, configuration, and recent activity.

Examples:
  tm status              # Full status check
  tm status --json       # JSON output
  tm status --watch      # Continuous monitoring`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if watch {
				return runStatusWatch(interval, jsonOutput)
			}
			return runStatus(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&watch, "watch", false, "Continuous monitoring")
	cmd.Flags().IntVar(&interval, "interval", 30, "Watch interval in seconds")

	return cmd
}

type systemStatus struct {
	Status    string        `json:"status"`
	Timestamp string        `json:"timestamp"`
	Mode      string        `json:"mode"`
	Sections  []statusGroup `json:"sections"`
}

type statusGroup struct {
	Name    string            `json:"name"`
	Status  string            `json:"status"`
	Details map[string]string `json:"details"`
	Issues  []string          `json:"issues,omitempty"`
}

func runStatus(jsonOutput bool) error {
	status := gatherStatus()

	if jsonOutput {
		return outputStatusJSON(status)
	}
	return outputStatusText(status)
}

func runStatusWatch(interval int, jsonOutput bool) error {
	for {
		// Clear screen
		fmt.Print("\033[H\033[2J")

		if err := runStatus(jsonOutput); err != nil {
			log.Warn().Err(err).Msg("status check failed")
		}

		fmt.Printf("\nRefreshing every %ds... (Ctrl+C to stop)\n", interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func gatherStatus() systemStatus {
	status := systemStatus{
		Status:    statusHealthy,
		Timestamp: time.Now().Format(time.RFC3339),
		Mode:      "unknown",
	}

	// Determine mode
	profilePath, _ := profile.DefaultPath()
	if profile.Exists(profilePath) {
		status.Mode = "universal"
	} else {
		status.Mode = "legacy"
	}

	// Gather sections
	status.Sections = []statusGroup{
		gatherConfigStatus(),
		gatherDatabaseStatus(),
		gatherLLMStatus(),
		gatherActivityStatus(),
		gatherSystemStatus(),
	}

	// Determine overall status
	for _, section := range status.Sections {
		switch section.Status {
		case statusError:
			status.Status = statusError
		case statusWarning:
			if status.Status != statusError {
				status.Status = statusWarning
			}
		}
	}

	return status
}

func gatherConfigStatus() statusGroup {
	group := statusGroup{
		Name:    "Configuration",
		Status:  statusHealthy,
		Details: make(map[string]string),
	}

	profilePath, _ := profile.DefaultPath()
	if profile.Exists(profilePath) {
		group.Details["mode"] = "Universal (profile-based)"
		group.Details["profile"] = profilePath

		p, err := profile.Load(profilePath)
		if err != nil {
			group.Status = statusError
			group.Issues = append(group.Issues, fmt.Sprintf("Failed to load profile: %v", err))
		} else {
			group.Details["goals"] = fmt.Sprintf("%d defined", len(p.Goals))
			group.Details["avoids"] = fmt.Sprintf("%d defined", len(p.Avoid))
		}
	} else {
		cfg, err := config.Load()
		if err != nil {
			group.Status = statusError
			group.Issues = append(group.Issues, "No configuration found")
			group.Issues = append(group.Issues, "Run 'tm init' to set up")
			return group
		}

		group.Details["mode"] = "Legacy (telos.md)"
		group.Details["telos"] = cfg.Telos.FilePath

		if _, err := os.Stat(cfg.Telos.FilePath); err != nil {
			group.Status = statusError
			group.Issues = append(group.Issues, "Telos file not found")
		}
	}

	return group
}

func gatherDatabaseStatus() statusGroup {
	group := statusGroup{
		Name:    "Database",
		Status:  statusHealthy,
		Details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		group.Status = statusError
		group.Issues = append(group.Issues, "Config not loaded")
		return group
	}

	dbPath := cfg.Database.Path
	group.Details["path"] = dbPath

	info, err := os.Stat(dbPath)
	if err != nil {
		group.Status = statusError
		group.Issues = append(group.Issues, "Database not found")
		return group
	}

	sizeMB := float64(info.Size()) / 1024 / 1024
	group.Details["size"] = fmt.Sprintf("%.1f MB", sizeMB)

	// Check connectivity
	repo, err := database.NewRepository(dbPath)
	if err != nil {
		group.Status = statusError
		group.Issues = append(group.Issues, "Connection failed")
		return group
	}
	defer func() { _ = repo.Close() }()

	if err := repo.Ping(); err != nil {
		group.Status = statusError
		group.Issues = append(group.Issues, "Ping failed")
		return group
	}

	// Count ideas
	ideas, err := repo.List(database.ListOptions{})
	if err == nil {
		active := 0
		for _, idea := range ideas {
			if idea.Status == "active" {
				active++
			}
		}
		group.Details["ideas"] = fmt.Sprintf("%d total (%d active)", len(ideas), active)
	}

	return group
}

func gatherLLMStatus() statusGroup {
	group := statusGroup{
		Name:    "AI Providers",
		Status:  statusHealthy,
		Details: make(map[string]string),
	}

	available := 0

	// Check Ollama
	ollamaURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(ollamaURL)
	if err == nil {
		_ = resp.Body.Close()
		group.Details["ollama"] = "available"
		available++
	} else {
		group.Details["ollama"] = "unavailable"
	}

	// Check OpenAI
	if os.Getenv("OPENAI_API_KEY") != "" {
		group.Details["openai"] = "configured"
		available++
	} else {
		group.Details["openai"] = "not configured"
	}

	// Check Claude
	if os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("CLAUDE_API_KEY") != "" {
		group.Details["claude"] = "configured"
		available++
	} else {
		group.Details["claude"] = "not configured"
	}

	// Rule-based always available
	group.Details["rule-based"] = "available"
	available++

	if available == 1 {
		group.Status = statusWarning
		group.Issues = append(group.Issues, "Only rule-based analysis available")
	}

	return group
}

func gatherActivityStatus() statusGroup {
	group := statusGroup{
		Name:    "Activity",
		Status:  statusHealthy,
		Details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		group.Status = statusWarning
		return group
	}

	repo, err := database.NewRepository(cfg.Database.Path)
	if err != nil {
		group.Status = statusWarning
		return group
	}
	defer func() { _ = repo.Close() }()

	ideas, err := repo.List(database.ListOptions{})
	if err != nil {
		group.Status = statusWarning
		return group
	}

	now := time.Now()
	today := 0
	week := 0
	var scores []float64

	for _, idea := range ideas {
		days := int(now.Sub(idea.CreatedAt).Hours() / 24)
		if days == 0 {
			today++
		}
		if days < 7 {
			week++
		}
		if days < 30 {
			scores = append(scores, idea.FinalScore)
		}
	}

	group.Details["today"] = fmt.Sprintf("%d ideas", today)
	group.Details["this week"] = fmt.Sprintf("%d ideas", week)

	if len(scores) > 0 {
		sum := 0.0
		for _, s := range scores {
			sum += s
		}
		avg := sum / float64(len(scores))
		group.Details["avg score (30d)"] = fmt.Sprintf("%.1f/10", avg)
	}

	return group
}

func gatherSystemStatus() statusGroup {
	group := statusGroup{
		Name:    "System",
		Status:  statusHealthy,
		Details: make(map[string]string),
	}

	group.Details["go"] = runtime.Version()
	group.Details["platform"] = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	group.Details["memory"] = fmt.Sprintf("%.1f MB", float64(m.Alloc)/1024/1024)

	// Simple health checks
	if ctx != nil && ctx.Repository != nil {
		// Check database
		if err := ctx.Repository.Ping(); err != nil {
			group.Status = statusError
			group.Issues = append(group.Issues, "Database connection failed")
		}

		// Check memory (warn if > 500MB)
		if float64(m.Alloc)/1024/1024 > 500.0 {
			group.Status = statusWarning
			group.Issues = append(group.Issues, "Memory usage high")
		}
	}

	return group
}

func outputStatusJSON(status systemStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputStatusText(status systemStatus) error {
	fmt.Println()
	fmt.Println("Brain Salad Status")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	// Overall status
	var statusIcon string
	var statusColor *color.Color
	switch status.Status {
	case statusError:
		statusIcon = "✗"
		statusColor = color.New(color.FgRed)
	case statusWarning:
		statusIcon = "⚠"
		statusColor = color.New(color.FgYellow)
	default:
		statusIcon = "✓"
		statusColor = color.New(color.FgGreen)
	}

	_, _ = statusColor.Printf("%s %s\n", statusIcon, strings.ToUpper(status.Status))
	fmt.Printf("Mode: %s\n", status.Mode)
	fmt.Println()

	// Sections
	for _, section := range status.Sections {
		printStatusSection(section)
	}

	fmt.Println(strings.Repeat("━", 50))
	return nil
}

func printStatusSection(section statusGroup) {
	// Section icon
	var icon string
	var sectionColor *color.Color
	switch section.Status {
	case statusError:
		icon = "✗"
		sectionColor = color.New(color.FgRed)
	case statusWarning:
		icon = "⚠"
		sectionColor = color.New(color.FgYellow)
	default:
		icon = "✓"
		sectionColor = color.New(color.FgGreen)
	}

	_, _ = sectionColor.Printf("%s %s\n", icon, section.Name)

	// Details
	for key, value := range section.Details {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Issues
	for _, issue := range section.Issues {
		_, _ = color.New(color.Faint).Printf("  → %s\n", issue)
	}

	fmt.Println()
}
