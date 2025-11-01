package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gh-standup/internal/github"
	"github.com/gh-standup/internal/llm"
	"github.com/gh-standup/internal/types"
	"github.com/spf13/cobra"
)

const extensionName = "standup"

var rootCmd = &cobra.Command{
	Use:   extensionName,
	Short: "Generate AI-powered standup reports",
	Long:  "A GitHub CLI extension that generates standup reports using GitHub Models and GitHub API data",
	RunE:  runStandup,
}

var (
	flagDays    int
	flagModel   string
	flagPrompts []string
	flagRepo    string
	flagUser    string
)

func init() {
	rootCmd.Flags().IntVarP(&flagDays, "days", "d", 1, "Number of days to look back for activity")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", "openai/gpt-4o", "GitHub Models model to use")
	rootCmd.Flags().StringArrayVarP(&flagPrompts, "prompts", "p", nil, "Override default prompt messages (can be specified multiple times) in format role:message")
	rootCmd.Flags().StringVarP(&flagRepo, "repo", "r", "", "Repository to generate standup for (owner/repo)")
	rootCmd.Flags().StringVarP(&flagUser, "user", "u", "", "User to generate standup for (defaults to authenticated user)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStandup(cmd *cobra.Command, args []string) error {
	githubClient, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	if flagUser == "" {
		user, err := githubClient.GetCurrentUser()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}
		flagUser = user
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -flagDays)

	activities, err := githubClient.CollectActivity(flagUser, flagRepo, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to collect GitHub activity: %w", err)
	}

	if len(activities) == 0 {
		log.Println("No GitHub activity found for the specified period.")
		return nil
	}

	log.Printf("Found %d activities\n", len(activities))

	commits, prs, issues, reviews := countActivities(activities)
	log.Printf("   %d commits, %d pull requests, %d issues, %d reviews\n", commits, prs, issues, reviews)

	llmClient, err := llm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	var promptMessages []llm.PromptMessage
	if len(flagPrompts) > 0 {
		for _, promptStr := range flagPrompts {
			parts := strings.SplitN(promptStr, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid prompt format (expected 'rule:message'): %s", promptStr)
			}
			promptMessages = append(promptMessages, llm.PromptMessage{
				Role:    parts[0],
				Content: parts[1],
			})
		}
	}

	// Generate standup report using GitHub Models
	report, err := llmClient.GenerateStandupReport(activities, flagModel, promptMessages)
	if err != nil {
		return fmt.Errorf("failed to generate standup report: %w", err)
	}

	fmt.Println(report)

	return nil
}

func countActivities(activities []types.GitHubActivity) (commits, prs, issues, reviews int) {
	for _, activity := range activities {
		switch activity.Type {
		case "commit":
			commits++
		case "pull_request":
			prs++
		case "issue":
			issues++
		case "review":
			reviews++
		}
	}
	return
}
