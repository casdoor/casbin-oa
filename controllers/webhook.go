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

package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/casbin/casbin-oa/object"
	"github.com/casbin/casbin-oa/util"
	"github.com/google/go-github/v74/github"
)

func (c *ApiController) WebhookOpen() {
	var issueEvent github.IssuesEvent
	var pullRequestEvent github.PullRequestEvent
	var checkRunEvent github.CheckRunEvent
	var checkSuiteEvent github.CheckSuiteEvent

	// Try to parse as different event types
	eventType := c.Ctx.Request.Header.Get("X-GitHub-Event")

	result := false
	switch eventType {
	case "check_run":
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &checkRunEvent)
		if err == nil && checkRunEvent.CheckRun != nil {
			result = HandleCheckRunEvent(checkRunEvent)
			c.Data["json"] = result
			c.ServeJSON()
			return
		}
	case "check_suite":
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &checkSuiteEvent)
		if err == nil && checkSuiteEvent.CheckSuite != nil {
			result = HandleCheckSuiteEvent(checkSuiteEvent)
			c.Data["json"] = result
			c.ServeJSON()
			return
		}
	}

	// Legacy handling for issues and pull requests
	json.Unmarshal(c.Ctx.Input.RequestBody, &pullRequestEvent)

	if pullRequestEvent.PullRequest != nil {
		result = PullRequestOpen(pullRequestEvent)
	} else {
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &issueEvent)
		if err != nil {
			panic(err)
		}
		result = IssueOpen(issueEvent)
	}

	c.Data["json"] = result
	c.ServeJSON()
}

func IssueOpen(issueEvent github.IssuesEvent) bool {
	if issueEvent.GetAction() != "opened" {
		return false
	}
	owner, repo := util.GetOwnerAndNameFromId(issueEvent.Repo.GetFullName())
	issueWebhook := object.GetIssueIfExist(owner, repo)
	if issueWebhook != nil {
		issueNumber := issueEvent.Issue.GetNumber()

		label := util.GetIssueLabel(issueEvent.Issue.GetTitle(), issueEvent.Issue.GetBody())
		if label != "" {
			go util.SetIssueLabel(owner, repo, issueNumber, label)
		}

		if issueWebhook.ProjectId != -1 {
			go util.AddIssueToProjectCard(issueWebhook.ProjectId, issueEvent.GetIssue().GetID())
		}

		if issueWebhook.Assignee != "" {
			go util.SetIssueAssignee(owner, repo, issueNumber, issueWebhook.Assignee)

		}

		if len(issueWebhook.AtPeople) != 0 {
			go util.AtPeople(issueWebhook.AtPeople, owner, repo, issueNumber)
		}

	}
	return true
}

func PullRequestOpen(pullRequestEvent github.PullRequestEvent) bool {
	if pullRequestEvent.GetAction() != "opened" {
		return false
	}
	owner, repo := util.GetOwnerAndNameFromId(pullRequestEvent.Repo.GetFullName())
	issueWebhook := object.GetIssueIfExist(owner, repo)

	if issueWebhook != nil {
		atPeople := issueWebhook.AtPeople
		sender := pullRequestEvent.Sender.GetLogin()

		if len(atPeople) != 0 {
			members := util.GetOrgMembers(owner)
			for i := 0; i < len(atPeople); i++ {
				_, existed := members[atPeople[i]]
				if !existed || atPeople[i] == sender {
					atPeople = append(atPeople[:i], atPeople[i+1:]...)
					i = i - 1
				}
			}
			if len(atPeople) != 0 {
				go util.RequestReviewers(owner, repo, pullRequestEvent.GetNumber(), atPeople)

				var commentStr string
				for i := range atPeople {
					commentStr = fmt.Sprintf("%s @%s", commentStr, atPeople[i])
				}
				commentStr = fmt.Sprintf("%s %s", commentStr, "please review")

				go util.Comment(commentStr, owner, repo, pullRequestEvent.GetNumber())
			}
		}
	}

	// Request automatic code review from copilot for human-created PRs
	// Check if webhook is configured for this repo
	if issueWebhook != nil {
		go util.RequestCopilotReview(owner, repo, pullRequestEvent.GetNumber())
	}

	return true
}

// HandleCheckRunEvent handles check_run webhook events
func HandleCheckRunEvent(event github.CheckRunEvent) bool {
	if event.GetAction() != "completed" {
		return false
	}

	checkRun := event.GetCheckRun()
	if checkRun == nil {
		return false
	}

	// Only process failed checks
	if checkRun.GetConclusion() != "failure" && checkRun.GetConclusion() != "cancelled" {
		return false
	}

	// Get PR information
	prs := checkRun.PullRequests
	if len(prs) == 0 {
		return false
	}

	owner, repo := util.GetOwnerAndNameFromId(event.Repo.GetFullName())
	issueWebhook := object.GetIssueIfExist(owner, repo)
	if issueWebhook == nil {
		return false
	}

	for _, pr := range prs {
		prNumber := pr.GetNumber()
		checkName := checkRun.GetName()

		// Check if this is a linter check
		if !util.IsLinterCheck(checkName) {
			continue
		}

		// Check if we should attempt to fix (max 3 attempts)
		if !object.ShouldAttemptFix(owner, repo, prNumber, checkName) {
			continue
		}

		// Get or create PR check record
		prCheck := object.GetPrCheck(owner, repo, prNumber, checkName)
		if prCheck == nil {
			// Create new record
			prCheck = &object.PrCheck{
				Org:           owner,
				Repo:          repo,
				PrNumber:      prNumber,
				CheckRunId:    checkRun.GetID(),
				CheckName:     checkName,
				Status:        checkRun.GetStatus(),
				Conclusion:    checkRun.GetConclusion(),
				FailureReason: util.GetCheckFailureDetails(checkRun),
				FixAttempts:   0,
				LastAttemptAt: time.Now(),
				IsFixed:       false,
				CreatedAt:     time.Now(),
			}
			object.AddPrCheck(prCheck)
		} else {
			// Update existing record
			prCheck.CheckRunId = checkRun.GetID()
			prCheck.Status = checkRun.GetStatus()
			prCheck.Conclusion = checkRun.GetConclusion()
			prCheck.FailureReason = util.GetCheckFailureDetails(checkRun)
			object.UpdatePrCheck(prCheck.Id, prCheck)
		}

		// Increment fix attempts and get updated record
		prCheck = object.IncrementFixAttempts(owner, repo, prNumber, checkName)
		if prCheck != nil {
			// Comment on PR with failure details and tag copilot
			go util.CommentOnPRWithCopilotTag(owner, repo, prNumber, prCheck.FailureReason, prCheck.FixAttempts)
		}
	}

	return true
}

// HandleCheckSuiteEvent handles check_suite webhook events
func HandleCheckSuiteEvent(event github.CheckSuiteEvent) bool {
	if event.GetAction() != "completed" {
		return false
	}

	checkSuite := event.GetCheckSuite()
	if checkSuite == nil {
		return false
	}

	// Only process failed check suites
	if checkSuite.GetConclusion() != "failure" && checkSuite.GetConclusion() != "cancelled" {
		return false
	}

	// Get PR information
	prs := checkSuite.PullRequests
	if len(prs) == 0 {
		return false
	}

	owner, repo := util.GetOwnerAndNameFromId(event.Repo.GetFullName())
	issueWebhook := object.GetIssueIfExist(owner, repo)
	if issueWebhook == nil {
		return false
	}

	// For check suites, we need to get individual check runs
	for _, pr := range prs {
		prNumber := pr.GetNumber()
		headSHA := checkSuite.GetHeadSHA()

		// Get all check runs for this commit
		checkRuns, err := util.GetPRCheckRuns(owner, repo, headSHA)
		if err != nil {
			continue
		}

		for _, checkRun := range checkRuns {
			if checkRun.GetConclusion() != "failure" && checkRun.GetConclusion() != "cancelled" {
				continue
			}

			checkName := checkRun.GetName()

			// Only process linter checks
			if !util.IsLinterCheck(checkName) {
				continue
			}

			// Check if we should attempt to fix
			if !object.ShouldAttemptFix(owner, repo, prNumber, checkName) {
				continue
			}

			// Get or create PR check record
			prCheck := object.GetPrCheck(owner, repo, prNumber, checkName)
			if prCheck == nil {
				prCheck = &object.PrCheck{
					Org:           owner,
					Repo:          repo,
					PrNumber:      prNumber,
					CheckRunId:    checkRun.GetID(),
					CheckName:     checkName,
					Status:        checkRun.GetStatus(),
					Conclusion:    checkRun.GetConclusion(),
					FailureReason: util.GetCheckFailureDetails(checkRun),
					FixAttempts:   0,
					LastAttemptAt: time.Now(),
					IsFixed:       false,
					CreatedAt:     time.Now(),
				}
				object.AddPrCheck(prCheck)
			}

			// Increment fix attempts and get updated record
			prCheck = object.IncrementFixAttempts(owner, repo, prNumber, checkName)
			if prCheck != nil {
				go util.CommentOnPRWithCopilotTag(owner, repo, prNumber, prCheck.FailureReason, prCheck.FixAttempts)
			}
		}
	}

	return true
}
