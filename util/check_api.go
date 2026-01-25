// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
)

const (
	// MaxCheckFailureTextLength is the maximum length of failure text to include in comments
	MaxCheckFailureTextLength = 500
	// CopilotUsername is the GitHub username of the copilot bot
	CopilotUsername = "copilot"
	// MaxFixAttempts is the maximum number of fix attempts for a failing check
	MaxFixAttempts = 3
)

// GetPRCheckRuns retrieves all check runs for a specific commit SHA
func GetPRCheckRuns(owner string, repo string, ref string) ([]*github.CheckRun, error) {
	client := GetClient()
	opts := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	result, _, err := client.Checks.ListCheckRunsForRef(context.Background(), owner, repo, ref, opts)
	if err != nil {
		return nil, err
	}

	return result.CheckRuns, nil
}

// GetCheckRunDetails retrieves details for a specific check run
func GetCheckRunDetails(owner string, repo string, checkRunID int64) (*github.CheckRun, error) {
	client := GetClient()
	checkRun, _, err := client.Checks.GetCheckRun(context.Background(), owner, repo, checkRunID)
	if err != nil {
		return nil, err
	}
	return checkRun, nil
}

// IsLinterCheck determines if a check is a linter check based on its name
func IsLinterCheck(checkName string) bool {
	linterKeywords := []string{
		"lint", "linter", "eslint", "golangci", "golint",
		"prettier", "style", "format", "rubocop", "pylint",
		"flake8", "clippy", "checkstyle", "pmd", "spotbugs",
	}

	checkNameLower := strings.ToLower(checkName)
	for _, keyword := range linterKeywords {
		if strings.Contains(checkNameLower, keyword) {
			return true
		}
	}
	return false
}

// GetCheckFailureDetails extracts failure details from a check run
func GetCheckFailureDetails(checkRun *github.CheckRun) string {
	if checkRun == nil {
		return "No check run details available"
	}

	var details strings.Builder
	details.WriteString(fmt.Sprintf("Check: %s\n", checkRun.GetName()))
	details.WriteString(fmt.Sprintf("Status: %s\n", checkRun.GetStatus()))
	details.WriteString(fmt.Sprintf("Conclusion: %s\n", checkRun.GetConclusion()))

	if checkRun.Output != nil {
		if checkRun.Output.Title != nil {
			details.WriteString(fmt.Sprintf("Title: %s\n", checkRun.Output.GetTitle()))
		}
		if checkRun.Output.Summary != nil {
			details.WriteString(fmt.Sprintf("Summary: %s\n", checkRun.Output.GetSummary()))
		}
		if checkRun.Output.Text != nil {
			text := checkRun.Output.GetText()
			// Limit the text to avoid too long messages
			if len(text) > MaxCheckFailureTextLength {
				text = text[:MaxCheckFailureTextLength] + "...(truncated)"
			}
			details.WriteString(fmt.Sprintf("Details: %s\n", text))
		}
	}

	return details.String()
}

// CommentOnPRWithCopilotTag comments on a PR and tags the copilot for fixing
func CommentOnPRWithCopilotTag(owner string, repo string, prNumber int, failureDetails string, attemptNumber int) error {
	commentBody := fmt.Sprintf(`@%s The CI check has failed. Please help fix the following issue:

**Attempt**: %d/%d

**Failure Details**:
%s

Please investigate and fix this issue.`, CopilotUsername, attemptNumber, MaxFixAttempts, failureDetails)

	success := Comment(commentBody, owner, repo, prNumber)
	if !success {
		return fmt.Errorf("failed to post comment on PR #%d", prNumber)
	}
	return nil
}

// RequestCopilotReview requests a review from copilot on a PR
func RequestCopilotReview(owner string, repo string, prNumber int) error {
	// First, comment to notify about the review request
	commentBody := fmt.Sprintf(`@%s Please review this PR.`, CopilotUsername)
	success := Comment(commentBody, owner, repo, prNumber)
	if !success {
		return fmt.Errorf("failed to post review request comment on PR #%d", prNumber)
	}

	// Try to request reviewer (may fail if copilot is not a collaborator)
	// We ignore errors here as the comment is the primary notification
	_ = RequestReviewers(owner, repo, prNumber, []string{CopilotUsername})

	return nil
}

// GetPRCommits retrieves commits for a PR
func GetPRCommits(owner string, repo string, prNumber int) ([]*github.RepositoryCommit, error) {
	client := GetClient()
	opts := &github.ListOptions{PerPage: 100}

	commits, _, err := client.PullRequests.ListCommits(context.Background(), owner, repo, prNumber, opts)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

// GetPRDetails retrieves details for a specific PR
func GetPRDetails(owner string, repo string, prNumber int) (*github.PullRequest, error) {
	client := GetClient()
	pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, prNumber)
	if err != nil {
		return nil, err
	}
	return pr, nil
}
