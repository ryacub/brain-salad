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
	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/config"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/telos"
	"github.com/spf13/cobra"
)

// Health status constants
const (
	statusHealthy = "healthy"
	statusWarning = "warning"
	statusError   = "error"
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
	status   string // statusHealthy, statusWarning, or statusError
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
	switch dbStatus.status {
	case statusWarning:
		warningCount++
	case statusError:
		errorCount++
	}

	// Check telos configuration
	telosStatus := checkTelosConfig()
	statuses = append(statuses, telosStatus)
	switch telosStatus.status {
	case statusWarning:
		warningCount++
	case statusError:
		errorCount++
	}

	// Check LLM providers
	llmStatus := checkLLMProviders()
	statuses = append(statuses, llmStatus)
	switch llmStatus.status {
	case statusWarning:
		warningCount++
	case statusError:
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
		status:  statusHealthy,
		details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		status.status = statusError
		status.messages = append(status.messages, "Failed to load configuration")
		return status
	}

	dbPath := cfg.Database.Path
	status.details["Location"] = dbPath

	// Check if database exists
	info, err := os.Stat(dbPath)
	if err != nil {
		status.status = statusError
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
		status.status = statusError
		status.messages = append(status.messages, fmt.Sprintf("Failed to connect: %v", err))
		return status
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			// Log but don't fail
			log.Warn().Err(closeErr).Msg("failed to close database in doctor check")
		}
	}()

	// Check connection
	if err := repo.Ping(); err != nil {
		status.status = statusError
		status.messages = append(status.messages, "Database connection failed")
		return status
	}

	// Get idea counts
	allIdeas, err := repo.List(database.ListOptions{})
	if err != nil {
		status.status = statusWarning
		status.messages = append(status.messages, "Could not query ideas")
	} else {
		activeCount := 0
		archivedCount := 0
		for _, idea := range allIdeas {
			switch idea.Status {
			case "active":
				activeCount++
			case "archived":
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
			status.status = statusWarning
			status.messages = append(status.messages, fmt.Sprintf("Integrity check: %s", integrityCheck))
		}
	}

	return status
}

func checkTelosConfig() healthStatus {
	status := healthStatus{
		section: "Telos Configuration",
		status:  statusHealthy,
		details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		status.status = statusError
		status.messages = append(status.messages, "Failed to load configuration")
		return status
	}

	telosPath := cfg.Telos.FilePath
	status.details["Location"] = telosPath

	// Check if file exists
	info, err := os.Stat(telosPath)
	if err != nil {
		status.status = statusError
		status.messages = append(status.messages, "Telos file not found")
		status.messages = append(status.messages, fmt.Sprintf("Create %s with your goals and strategies", telosPath))
		return status
	}

	// Parse telos file
	parser := telos.NewParser()
	telosData, err := parser.ParseFile(telosPath)
	if err != nil {
		status.status = statusError
		status.messages = append(status.messages, fmt.Sprintf("Failed to parse telos file: %v", err))
		return status
	}

	// Count sections
	status.details["Goals"] = fmt.Sprintf("%d defined", len(telosData.Goals))
	status.details["Strategies"] = fmt.Sprintf("%d defined", len(telosData.Strategies))
	status.details["Failure Patterns"] = fmt.Sprintf("%d defined", len(telosData.Challenges))

	// Check if empty
	if len(telosData.Goals) == 0 && len(telosData.Strategies) == 0 {
		status.status = statusWarning
		status.messages = append(status.messages, "Telos file is empty or invalid")
		status.messages = append(status.messages, "Add your goals and strategies for better idea scoring")
	}

	// Last modified
	modTime := info.ModTime()
	daysAgo := int(time.Since(modTime).Hours() / 24)
	switch daysAgo {
	case 0:
		status.details["Last modified"] = "today"
	case 1:
		status.details["Last modified"] = "yesterday"
	default:
		status.details["Last modified"] = fmt.Sprintf("%d days ago", daysAgo)
	}

	return status
}

func checkLLMProviders() healthStatus {
	status := healthStatus{
		section:  "LLM Providers",
		status:   statusWarning,
		messages: []string{},
		details:  make(map[string]string),
	}

	// Check OpenAI
	if os.Getenv("OPENAI_API_KEY") != "" {
		status.details["openai"] = "✓ Configured"
		status.status = statusHealthy
	} else {
		status.details["openai"] = "✗ OPENAI_API_KEY not set"
	}

	// Check Claude
	if os.Getenv("CLAUDE_API_KEY") != "" {
		status.details["claude"] = "✓ Configured"
		status.status = statusHealthy
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
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close response body")
		}
		status.details["ollama"] = fmt.Sprintf("✓ Available (%s)", ollamaURL)
		status.status = statusHealthy
	} else {
		status.details["ollama"] = fmt.Sprintf("✗ Connection failed (%s)", ollamaURL)
	}

	// Rule-based always available
	status.details["rule-based"] = "✓ Available (no API key needed)"
	// Rule-based is always available, so we always have at least one provider
	if status.status != statusHealthy {
		status.status = statusHealthy
	}

	return status
}

func checkSystem() healthStatus {
	status := healthStatus{
		section: "System",
		status:  statusHealthy,
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
			if closeErr := repo.Close(); closeErr != nil {
				log.Warn().Err(closeErr).Msg("failed to close repository in system check")
			}
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
		status:  statusHealthy,
		details: make(map[string]string),
	}

	cfg, err := config.Load()
	if err != nil {
		status.status = statusWarning
		status.messages = append(status.messages, "Could not load configuration")
		return status
	}

	repo, err := database.NewRepository(cfg.Database.Path)
	if err != nil {
		status.status = statusWarning
		status.messages = append(status.messages, "Could not connect to database")
		return status
	}
	defer func() {
		if closeErr := repo.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close repository in activity check")
		}
	}()

	allIdeas, err := repo.List(database.ListOptions{})
	if err != nil {
		status.status = statusWarning
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
	case statusHealthy:
		icon = "✓"
		statusColor = color.New(color.FgGreen)
	case statusWarning:
		icon = "⚠"
		statusColor = color.New(color.FgYellow)
	case statusError:
		icon = "✗"
		statusColor = color.New(color.FgRed)
	default:
		icon = "•"
		statusColor = color.New(color.FgWhite)
	}

	// Print section header
	if _, err := statusColor.Printf("%s %s\n", icon, status.section); err != nil {
		log.Warn().Err(err).Msg("failed to print status header")
	}

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
		if _, err := color.New(color.Faint).Printf("  %s\n", msg); err != nil {
			log.Warn().Err(err).Msg("failed to print status message")
		}
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
