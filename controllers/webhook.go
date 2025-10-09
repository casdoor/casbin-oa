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

	"github.com/casbin/casbin-oa/object"
	"github.com/casbin/casbin-oa/util"
	"github.com/google/go-github/v74/github"
)

func (c *ApiController) WebhookOpen() {
	var issueEvent github.IssuesEvent
	var pullRequestEvent github.PullRequestEvent
	json.Unmarshal(c.Ctx.Input.RequestBody, &pullRequestEvent)

	var result bool
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
	return true
}
