package dump

import (
	"fmt"
	"strings"

	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/llm"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/ryacub/telos-idea-matrix/internal/patterns"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
	"github.com/ryacub/telos-idea-matrix/internal/utils"
	"github.com/spf13/cobra"
)

// CLIContext represents the shared CLI dependencies
type CLIContext struct {
	Repository interface{ Create(*models.Idea) error }
	Engine     *scoring.Engine
	Detector   *patterns.Detector
	Telos      *models.Telos
	LLMManager *llm.Manager
}

// NewDumpCommand creates the dump command
func NewDumpCommand(getContext func() *CLIContext) *cobra.Command {
	var fromClipboard bool
	var toClipboard bool
	var interactive bool
	var quick bool
	var useAI bool
	var provider string
	var model string

	cmd := &cobra.Command{
		Use:   "dump <idea text>",
		Short: "Capture and analyze an idea immediately",
		Long: `Capture a new idea, analyze it against your telos, and save it to the database.
The idea will be scored and analyzed for patterns immediately.

Modes:
  Normal      - Standard analysis with scoring engine (default)
  Interactive - Step-by-step analysis with LLM and user confirmations
  Quick       - Fast capture without detailed analysis

Examples:
  tm dump "Build a SaaS product for developers"
  tm dump "Start a podcast" --use-ai
  tm dump "Learn Rust" --use-ai --provider ollama
  tm dump --interactive "Start a podcast"
  tm dump --quick "Write a blog post"
  tm dump --from-clipboard
  tm dump "Quick idea" --to-clipboard
  tm dump --quick "Fast idea capture"`,
		Args: func(cmd *cobra.Command, args []string) error {
			fromClipboard, _ := cmd.Flags().GetBool("from-clipboard")
			if !fromClipboard && len(args) < 1 {
				return fmt.Errorf("provide idea text or use --from-clipboard")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext()
			if ctx == nil {
				return fmt.Errorf("CLI context not initialized")
			}

			// Get idea text from clipboard or arguments
			var ideaText string
			if fromClipboard {
				text, err := utils.PasteFromClipboard()
				if err != nil {
					return fmt.Errorf("read clipboard: %w", err)
				}
				ideaText = strings.TrimSpace(text)
				if ideaText == "" {
					return fmt.Errorf("clipboard is empty")
				}
			} else {
				ideaText = strings.Join(args, " ")
			}

			// Route to appropriate mode
			if interactive {
				return runInteractiveDump(ideaText, provider, ctx.Telos, ctx.Repository)
			}

			if quick {
				return runQuickDump(ideaText, toClipboard, ctx.Repository)
			}

			// Normal dump
			dumpCtx := DumpContext{
				Repository: ctx.Repository,
				Engine:     ctx.Engine,
				Detector:   ctx.Detector,
			}

			llmAnalyzer := func(idea, prov, mdl string) (*models.Analysis, error) {
				return runLLMAnalysisWithProvider(idea, prov, mdl, ctx.LLMManager, ctx.Telos)
			}

			return runNormalDump(ideaText, fromClipboard, toClipboard, useAI, provider, model, dumpCtx, llmAnalyzer)
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode with step-through analysis")
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, "Quick mode with rule-based analysis")
	cmd.Flags().BoolVar(&useAI, "use-ai", false, "Use LLM analysis (requires Ollama or API keys)")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "LLM provider to use (ollama|openai|claude|rule_based)")
	cmd.Flags().StringVar(&model, "model", "", "LLM model to use")
	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "Read idea from clipboard")
	cmd.Flags().BoolVar(&toClipboard, "to-clipboard", false, "Copy result to clipboard")

	return cmd
}

// RunQuickDump is exported for use by other commands like batch dump
func RunQuickDump(content string, toClipboard bool, repo interface{ Create(*models.Idea) error }) error {
	return runQuickDump(content, toClipboard, repo)
}

// TruncateText is exported for use by other commands
func TruncateText(text string, maxLen int) string {
	return cliutil.TruncateText(text, maxLen)
}
