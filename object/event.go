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

package object

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/casbin/casbin-oa/util"
	"github.com/google/go-github/v38/github"
)

type Event struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	HtmlURL  string `json:"html_url"`
	CreateAt string `json:"create_at"`
	State    string `json:"state"`
	OrgName  string `json:"org_name"`
	RepoName string `json:"repo_name"`
	Number   int    `json:"number"`
}

func GetEvents(author string, orgMap map[string]bool, startDate time.Time, endDate time.Time) []*Event {
	client := util.GetClient()
	activity := client.Activity

	curPage := 1
	options := github.ListOptions{Page: curPage, PerPage: 100}

	Events, _, err := activity.ListEventsPerformedByUser(context.Background(), author, true, &options)
	if err != nil {
		panic(err)
	}

	i := 0

	var authorEvents []*Event
	prMap := make(map[string]bool)
	issueMap := make(map[string]bool)
	reviewMap := make(map[string]bool)
	for {
		i++
		if len(Events) == 0 {
			return authorEvents
		}

		if i > len(Events) {
			i = 0
			curPage++
			options = github.ListOptions{Page: curPage}
			Events, _, err = activity.ListEventsPerformedByUser(context.Background(), author, true, &options)
			if err != nil {
				if strings.Contains(err.Error(), "pagination is limited for this resource") {
					fmt.Printf("GetEvents() error: %s\n", err.Error())
					break
				}
				panic(err)
			}
			continue
		}
		curEvent := Events[i-1]
		creatTime := *curEvent.CreatedAt
		if creatTime.After(endDate) {
			continue
		}
		if creatTime.Before(startDate) {
			return authorEvents
		}

		fullName := *curEvent.Repo.Name
		orgName, repoName := util.GetOwnerAndNameFromId(fullName)

		_, okOrg := orgMap[orgName]
		_, okFullName := orgMap[fullName]
		if !okOrg && !okFullName {
			continue
		}

		Date := fmt.Sprintf("%d", creatTime.YearDay()-startDate.YearDay())

		if *curEvent.Type == "PullRequestEvent" {
			var payLoad PayLoad
			err = json.Unmarshal(*curEvent.RawPayload, &payLoad)
			if err != nil {
				panic(err)
			}

			request := payLoad.PullRequest
			state := GetPRState(orgName, repoName, request.GetNumber())
			_, ok := prMap[*request.HTMLURL]
			if ok {
				continue
			} else {
				prMap[*request.HTMLURL] = true
			}
			event := Event{Type: "PR", Title: *request.Title, HtmlURL: *request.HTMLURL, CreateAt: Date, State: state, OrgName: orgName, RepoName: repoName, Number: *request.Number}
			authorEvents = append(authorEvents, &event)

		} else if *curEvent.Type == "IssueCommentEvent" {
			var payLoad PayLoad
			err = json.Unmarshal(*curEvent.RawPayload, &payLoad)
			if err != nil {
				panic(err)
			}

			comment := payLoad.Comment
			issue := payLoad.Issue
			body := *comment.Body
			issueId := fmt.Sprintf("%s%s%v", orgName, repoName, *issue.Number)
			_, ok := issueMap[issueId]
			if ok {
				continue
			} else {
				issueMap[issueId] = true
			}
			if len(body) > 100 {
				body = fmt.Sprintf("%s%s", body[0:99], ". . .")
			}
			body = strings.ReplaceAll(body, "\r\n", " ")
			body = strings.ReplaceAll(body, "\n", "")
			event := Event{Type: "IssueComment", Title: body, HtmlURL: *comment.HTMLURL, CreateAt: Date, OrgName: orgName, RepoName: repoName, Number: *issue.Number}
			authorEvents = append(authorEvents, &event)
		} else if *curEvent.Type == "PullRequestReviewEvent" {
			var payLoad PayLoad
			err = json.Unmarshal(*curEvent.RawPayload, &payLoad)
			if err != nil {
				panic(err)
			}

			pr := payLoad.PullRequest
			if *pr.User.Login == author {
				continue
			}
			review := payLoad.Review

			reviewURL := *review.HTMLURL
			_, ok := reviewMap[reviewURL]
			if ok {
				continue
			} else {
				reviewMap[reviewURL] = true
			}

			event := Event{Type: "CodeReview", HtmlURL: reviewURL, CreateAt: Date, OrgName: orgName, RepoName: repoName, Number: *pr.Number}
			authorEvents = append(authorEvents, &event)
		}

	}
	return authorEvents
}

func GetPRState(org string, repo string, pullNumber int) string {
	client := util.GetClient()
	pr, _, err := client.PullRequests.Get(context.Background(), org, repo, pullNumber)
	if err != nil {
		panic(err)
	}
	if pr.MergedAt != nil {
		return "Merged"
	}
	if *pr.Draft && *pr.State == "open" {
		return "Draft"
	}

	return *pr.State
}
