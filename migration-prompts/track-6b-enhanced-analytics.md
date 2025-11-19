# Track 6B: Enhanced Analytics

**Phase**: 6 - Advanced CLI Features
**Estimated Time**: 6-8 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 5D, 6A)

---

## Mission

You are implementing enhanced analytics and trend analysis for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has analytics in `src/commands/analytics.rs`
- Need trend analysis over time (score trends, pattern frequency)
- Pattern frequency analysis
- Comprehensive reporting with visualizations

## Reference Implementation

Review `/home/user/brain-salad/src/commands/analytics.rs` for:
- Trend calculations
- Pattern frequency
- Report generation
- Time-based grouping

## Your Task

Implement enhanced analytics using strict TDD methodology.

## Directory Structure

Enhance existing `go/internal/cli/analytics.go` and create:
- `internal/analytics/trends.go` - Trend analysis
- `internal/analytics/reports.go` - Report generation
- `internal/analytics/trends_test.go` - Trend tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/analytics/trends_test.go`:
- `TestTrends_ScoreByWeek()`
- `TestTrends_ScoreByMonth()`
- `TestTrends_PatternFrequency()`
- `TestTrends_IdeaCreationRate()`

Run: `go test ./internal/analytics -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/analytics/trends.go`:

```go
package analytics

import (
    "time"
    
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

type TrendData struct {
    Period      string
    AvgScore    float64
    IdeaCount   int
    TopPatterns []string
}

// CalculateScoreTrends groups ideas by time period and calculates average scores
func CalculateScoreTrends(ideas []*models.Idea, groupBy string) []TrendData {
    groups := make(map[string][]*models.Idea)
    
    for _, idea := range ideas {
        var key string
        switch groupBy {
        case "week":
            year, week := idea.CreatedAt.ISOWeek()
            key = fmt.Sprintf("%d-W%02d", year, week)
        case "month":
            key = idea.CreatedAt.Format("2006-01")
        case "day":
            key = idea.CreatedAt.Format("2006-01-02")
        }
        groups[key] = append(groups[key], idea)
    }
    
    trends := make([]TrendData, 0, len(groups))
    for period, periodIdeas := range groups {
        totalScore := 0.0
        for _, idea := range periodIdeas {
            totalScore += idea.Score
        }
        
        avgScore := totalScore / float64(len(periodIdeas))
        
        trends = append(trends, TrendData{
            Period:    period,
            AvgScore:  avgScore,
            IdeaCount: len(periodIdeas),
        })
    }
    
    // Sort by period
    sort.Slice(trends, func(i, j int) bool {
        return trends[i].Period < trends[j].Period
    })
    
    return trends
}

// CalculatePatternFrequency counts how often each pattern appears
func CalculatePatternFrequency(ideas []*models.Idea) map[string]int {
    freq := make(map[string]int)
    
    for _, idea := range ideas {
        if idea.Analysis == nil {
            continue
        }
        
        for _, pattern := range idea.Analysis.DetectedPatterns {
            freq[pattern.Name]++
        }
    }
    
    return freq
}

// CalculateCreationRate returns ideas created per day
func CalculateCreationRate(ideas []*models.Idea, days int) float64 {
    if len(ideas) == 0 {
        return 0.0
    }
    
    cutoff := time.Now().AddDate(0, 0, -days)
    count := 0
    
    for _, idea := range ideas {
        if idea.CreatedAt.After(cutoff) {
            count++
        }
    }
    
    return float64(count) / float64(days)
}
```

#### B. Implement `go/internal/analytics/reports.go`:

```go
package analytics

import (
    "fmt"
    "strings"
    
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

type Report struct {
    Title       string
    Summary     string
    Sections    []ReportSection
    GeneratedAt time.Time
}

type ReportSection struct {
    Title   string
    Content string
}

// GenerateReport creates a comprehensive analytics report
func GenerateReport(ideas []*models.Idea) Report {
    report := Report{
        Title:       "Telos Idea Matrix Analytics Report",
        GeneratedAt: time.Now(),
        Sections:    make([]ReportSection, 0),
    }
    
    // Summary section
    report.Summary = fmt.Sprintf("Total Ideas: %d\nReport Generated: %s",
        len(ideas), time.Now().Format("2006-01-02 15:04"))
    
    // Score distribution
    high, medium, low := 0, 0, 0
    for _, idea := range ideas {
        if idea.Score >= 8.0 {
            high++
        } else if idea.Score >= 5.0 {
            medium++
        } else {
            low++
        }
    }
    
    report.Sections = append(report.Sections, ReportSection{
        Title: "Score Distribution",
        Content: fmt.Sprintf(`
High Score (8.0+):   %d ideas (%d%%)
Medium Score (5-8):  %d ideas (%d%%)
Low Score (<5):      %d ideas (%d%%)
`,
            high, (high*100)/len(ideas),
            medium, (medium*100)/len(ideas),
            low, (low*100)/len(ideas)),
    })
    
    // Trends section
    trends := CalculateScoreTrends(ideas, "month")
    trendsContent := "Score Trends by Month:\n"
    for _, trend := range trends {
        trendsContent += fmt.Sprintf("  %s: %.1f avg (% d ideas)\n",
            trend.Period, trend.AvgScore, trend.IdeaCount)
    }
    
    report.Sections = append(report.Sections, ReportSection{
        Title:   "Trends",
        Content: trendsContent,
    })
    
    // Pattern frequency
    patterns := CalculatePatternFrequency(ideas)
    patternsContent := "Most Common Patterns:\n"
    for pattern, count := range patterns {
        patternsContent += fmt.Sprintf("  %s: %d occurrences\n", pattern, count)
    }
    
    report.Sections = append(report.Sections, ReportSection{
        Title:   "Patterns",
        Content: patternsContent,
    })
    
    return report
}

// RenderReport converts report to markdown
func RenderReport(report Report) string {
    var sb strings.Builder
    
    sb.WriteString(fmt.Sprintf("# %s\n\n", report.Title))
    sb.WriteString(fmt.Sprintf("%s\n\n", report.Summary))
    
    for _, section := range report.Sections {
        sb.WriteString(fmt.Sprintf("## %s\n\n", section.Title))
        sb.WriteString(section.Content)
        sb.WriteString("\n\n")
    }
    
    return sb.String()
}
```

#### C. Enhance `go/internal/cli/analytics.go`:

```go
// Add new subcommands
func newAnalyticsTrendsCommand(ctx *CLIContext) *cobra.Command {
    var days int
    var groupBy string
    
    cmd := &cobra.Command{
        Use:   "trends",
        Short: "Show score trends over time",
        RunE: func(cmd *cobra.Command, args []string) error {
            ideas, err := ctx.Repo.List(0, "", 10000)
            if err != nil {
                return err
            }
            
            trends := analytics.CalculateScoreTrends(ideas, groupBy)
            
            fmt.Println("Score Trends:")
            for _, trend := range trends {
                fmt.Printf("  %s: %.1f avg (%d ideas)\n",
                    trend.Period, trend.AvgScore, trend.IdeaCount)
            }
            
            return nil
        },
    }
    
    cmd.Flags().IntVar(&days, "days", 30, "Number of days to analyze")
    cmd.Flags().StringVar(&groupBy, "group-by", "week", "Group by: day, week, month")
    
    return cmd
}

func newAnalyticsReportCommand(ctx *CLIContext) *cobra.Command {
    var outputFile string
    
    cmd := &cobra.Command{
        Use:   "report",
        Short: "Generate comprehensive analytics report",
        RunE: func(cmd *cobra.Command, args []string) error {
            ideas, err := ctx.Repo.List(0, "", 10000)
            if err != nil {
                return err
            }
            
            report := analytics.GenerateReport(ideas)
            markdown := analytics.RenderReport(report)
            
            if outputFile != "" {
                return os.WriteFile(outputFile, []byte(markdown), 0644)
            }
            
            fmt.Println(markdown)
            return nil
        },
    }
    
    cmd.Flags().StringVar(&outputFile, "output", "", "Output file (default: stdout)")
    
    return cmd
}
```

Run: `go test ./internal/analytics -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add ASCII charts for trends
- Optimize queries for large datasets
- Add caching for expensive analytics
- Extract visualization helpers

## Success Criteria

- ✅ All tests pass with >80% coverage
- ✅ Trends calculated correctly
- ✅ Reports readable and actionable
- ✅ Performance <1s for 10,000 ideas

## Validation

```bash
# View trends
tm analytics trends --days 30 --group-by week
tm analytics trends --group-by month

# Generate report
tm analytics report
tm analytics report --output report.md

# Pattern frequency
tm analytics patterns
```

## Deliverables

- `go/internal/analytics/trends.go`
- `go/internal/analytics/reports.go`
- `go/internal/analytics/trends_test.go`
- Enhanced `go/internal/cli/analytics.go`
