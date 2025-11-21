// Package health provides health check and diagnostic commands.
package health

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/config"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/spf13/cobra"
)

// NewDoctorCommand creates the doctor command
func NewDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and configuration",
		Long:  "Runs diagnostics on database, telos configuration, LLM providers, and system status",
		RunE:  runDoctor,
	}
}

type healthStatus struct {
	section  string
	status   string // "healthy", "warning", "error"
	messages []string
	details  map[string]string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Print header
	printHeader("Telos Idea Matrix Health Check")

	var statuses []healthStatus
	warningCount := 0
	errorCount := 0

	// Check database
	dbStatus := checkDatabase()
	statuses = append(statuses, dbStatus)
	if dbStatus.status == "warning" {
		warningCount++
	} else if dbStatus.status == "error" {
		errorCount++
	}

	// Check telos configuration
	telosStatus := checkTelosConfig()
	statuses = append(statuses, telosStatus)
	if telosStatus.status == "warning" {
		warningCount++
	} else if telosStatus.status == "error" {
		errorCount++
	}

	// Check LLM providers
	llmStatus := checkLLMProviders()
	statuses = append(statuses, llmStatus)
	if llmStatus.status == "warning" {
		warningCount++
	} else if llmStatus.status == "error" {
		errorCount++
	}

	// Check system
	sysStatus := checkSystem()
	statuses = append(statuses, sysStatus)

	// Check recent activity
	activityStatus := checkRecentActivity()
	statuses = append(statuses, activityStatus)

	// Print all statuses
	for _, status := range statuses {
		printStatus(status)
	}

	// Print summary
	printSummary(errorCount, warningCount)

	return nil
}

func checkDatabase() healthStatus {
	status := healthStatus{
		section: "Database",
		status:  "healthy",
		details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		status.status = "error"
		status.messages = append(status.messages, "Failed to load configuration")
		return status
	}

	dbPath := cfg.Database.Path
	status.details["Location"] = dbPath

	// Check if database exists
	info, err := os.Stat(dbPath)
	if err != nil {
		status.status = "error"
		status.messages = append(status.messages, "Database file not found")
		status.messages = append(status.messages, "Run 'tm init' to initialize")
		return status
	}

	// Get database size
	sizeMB := float64(info.Size()) / 1024 / 1024
	status.details["Size"] = fmt.Sprintf("%.1f MB", sizeMB)

	// Connect to database
	repo, err := database.NewRepository(dbPath)
	if err != nil {
		status.status = "error"
		status.messages = append(status.messages, fmt.Sprintf("Failed to connect: %v", err))
		return status
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			// Log but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to close database: %v\n", closeErr)
		}
	}()

	// Check connection
	if err := repo.Ping(); err != nil {
		status.status = "error"
		status.messages = append(status.messages, "Database connection failed")
		return status
	}

	// Get idea counts
	allIdeas, err := repo.List(database.ListOptions{})
	if err != nil {
		status.status = "warning"
		status.messages = append(status.messages, "Could not query ideas")
	} else {
		activeCount := 0
		archivedCount := 0
		for _, idea := range allIdeas {
			if idea.Status == "active" {
				activeCount++
			} else if idea.Status == "archived" {
				archivedCount++
			}
		}
		status.details["Ideas"] = fmt.Sprintf("%d total (%d active, %d archived)",
			len(allIdeas), activeCount, archivedCount)
	}

	// Check database integrity
	db := repo.DB()
	var integrityCheck string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&integrityCheck); err == nil {
		if integrityCheck != "ok" {
			status.status = "warning"
			status.messages = append(status.messages, fmt.Sprintf("Integrity check: %s", integrityCheck))
		}
	}

	return status
}

func checkTelosConfig() healthStatus {
	status := healthStatus{
		section: "Telos Configuration",
		status:  "healthy",
		details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		status.status = "error"
		status.messages = append(status.messages, "Failed to load configuration")
		return status
	}

	telosPath := cfg.Telos.FilePath
	status.details["Location"] = telosPath

	// Check if file exists
	info, err := os.Stat(telosPath)
	if err != nil {
		status.status = "error"
		status.messages = append(status.messages, "Telos file not found")
		status.messages = append(status.messages, fmt.Sprintf("Create %s with your goals and strategies", telosPath))
		return status
	}

	// Parse telos file
	parser := telos.NewParser()
	telosData, err := parser.ParseFile(telosPath)
	if err != nil {
		status.status = "error"
		status.messages = append(status.messages, fmt.Sprintf("Failed to parse telos file: %v", err))
		return status
	}

	// Count sections
	status.details["Goals"] = fmt.Sprintf("%d defined", len(telosData.Goals))
	status.details["Strategies"] = fmt.Sprintf("%d defined", len(telosData.Strategies))
	status.details["Failure Patterns"] = fmt.Sprintf("%d defined", len(telosData.Challenges))

	// Check if empty
	if len(telosData.Goals) == 0 && len(telosData.Strategies) == 0 {
		status.status = "warning"
		status.messages = append(status.messages, "Telos file is empty or invalid")
		status.messages = append(status.messages, "Add your goals and strategies for better idea scoring")
	}

	// Last modified
	modTime := info.ModTime()
	daysAgo := int(time.Since(modTime).Hours() / 24)
	if daysAgo == 0 {
		status.details["Last modified"] = "today"
	} else if daysAgo == 1 {
		status.details["Last modified"] = "yesterday"
	} else {
		status.details["Last modified"] = fmt.Sprintf("%d days ago", daysAgo)
	}

	return status
}

func checkLLMProviders() healthStatus {
	status := healthStatus{
		section:  "LLM Providers",
		status:   "warning",
		messages: []string{},
		details:  make(map[string]string),
	}

	hasWorkingProvider := false

	// Check OpenAI
	if os.Getenv("OPENAI_API_KEY") != "" {
		status.details["openai"] = "✓ Configured"
		hasWorkingProvider = true
	} else {
		status.details["openai"] = "✗ OPENAI_API_KEY not set"
	}

	// Check Claude
	if os.Getenv("CLAUDE_API_KEY") != "" {
		status.details["claude"] = "✓ Configured"
		hasWorkingProvider = true
	} else {
		status.details["claude"] = "✗ CLAUDE_API_KEY not set"
	}

	// Check Ollama (try to connect with timeout)
	ollamaURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	// Quick connectivity check with 2 second timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(ollamaURL)
	if err == nil {
		resp.Body.Close()
		status.details["ollama"] = fmt.Sprintf("✓ Available (%s)", ollamaURL)
		hasWorkingProvider = true
	} else {
		status.details["ollama"] = fmt.Sprintf("✗ Connection failed (%s)", ollamaURL)
	}

	// Rule-based always available
	status.details["rule-based"] = "✓ Available (no API key needed)"
	hasWorkingProvider = true

	// Update status based on findings
	if hasWorkingProvider {
		status.status = "healthy"
	} else {
		status.messages = append(status.messages, "No LLM providers configured")
		status.messages = append(status.messages, "Set OPENAI_API_KEY for AI-powered analysis:")
		status.messages = append(status.messages, "  export OPENAI_API_KEY=sk-...")
	}

	return status
}

func checkSystem() healthStatus {
	status := healthStatus{
		section: "System",
		status:  "healthy",
		details: make(map[string]string),
	}

	// Go version
	status.details["Go version"] = runtime.Version()

	// Platform
	status.details["Platform"] = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// SQLite version
	cfg, err := config.Load()
	if err == nil {
		// Try to get SQLite version
		if repo, err := database.NewRepository(cfg.Database.Path); err == nil {
			var sqliteVersion string
			if err := repo.DB().QueryRow("SELECT sqlite_version()").Scan(&sqliteVersion); err == nil {
				status.details["SQLite version"] = sqliteVersion
			}
			repo.Close()
		}
	}

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := float64(m.Alloc) / 1024 / 1024
	status.details["Memory usage"] = fmt.Sprintf("%.1f MB", allocMB)

	return status
}

func checkRecentActivity() healthStatus {
	status := healthStatus{
		section: "Recent Activity",
		status:  "healthy",
		details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		status.status = "warning"
		status.messages = append(status.messages, "Could not load configuration")
		return status
	}

	repo, err := database.NewRepository(cfg.Database.Path)
	if err != nil {
		status.status = "warning"
		status.messages = append(status.messages, "Could not connect to database")
		return status
	}
	defer repo.Close()

	allIdeas, err := repo.List(database.ListOptions{})
	if err != nil {
		status.status = "warning"
		status.messages = append(status.messages, "Could not query ideas")
		return status
	}

	// Count ideas by recency
	now := time.Now()
	todayCount := 0
	weekCount := 0
	monthCount := 0
	var monthScores []float64

	for _, idea := range allIdeas {
		daysSince := int(now.Sub(idea.CreatedAt).Hours() / 24)

		if daysSince == 0 {
			todayCount++
		}
		if daysSince < 7 {
			weekCount++
		}
		if daysSince < 30 {
			monthCount++
			monthScores = append(monthScores, idea.FinalScore)
		}
	}

	status.details["Ideas created today"] = fmt.Sprintf("%d", todayCount)
	status.details["Ideas created this week"] = fmt.Sprintf("%d", weekCount)

	// Average score
	if len(monthScores) > 0 {
		sum := 0.0
		for _, score := range monthScores {
			sum += score
		}
		avg := sum / float64(len(monthScores))
		status.details["Average score (last 30 days)"] = fmt.Sprintf("%.1f/10", avg)
	} else {
		status.details["Average score (last 30 days)"] = "N/A (no ideas)"
	}

	return status
}

func printHeader(title string) {
	fmt.Println()
	fmt.Println(title)
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()
}

func printStatus(status healthStatus) {
	// Status icon and color
	var icon string
	var statusColor *color.Color

	switch status.status {
	case "healthy":
		icon = "✓"
		statusColor = color.New(color.FgGreen)
	case "warning":
		icon = "⚠"
		statusColor = color.New(color.FgYellow)
	case "error":
		icon = "✗"
		statusColor = color.New(color.FgRed)
	default:
		icon = "•"
		statusColor = color.New(color.FgWhite)
	}

	// Print section header
	statusColor.Printf("%s %s\n", icon, status.section)

	// Print details (sorted for consistency)
	detailKeys := []string{
		"Location", "Status", "Size", "Ideas", "Goals", "Strategies",
		"Failure Patterns", "Last modified", "openai", "claude", "ollama",
		"rule-based", "Go version", "SQLite version", "Platform",
		"Memory usage", "Ideas created today", "Ideas created this week",
		"Average score (last 30 days)",
	}

	for _, key := range detailKeys {
		if value, ok := status.details[key]; ok {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Print messages
	for _, msg := range status.messages {
		color.New(color.Faint).Printf("  %s\n", msg)
	}

	fmt.Println()
}

func printSummary(errorCount, warningCount int) {
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	if errorCount > 0 {
		color.Red("Overall Status: UNHEALTHY (%d errors, %d warnings)", errorCount, warningCount)
	} else if warningCount > 0 {
		color.Yellow("Overall Status: HEALTHY (%d warnings)", warningCount)
	} else {
		color.Green("Overall Status: HEALTHY")
	}
	fmt.Println()
}
