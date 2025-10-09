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
	"strings"

	"github.com/astaxie/beego"
	"github.com/casbin/casbin-oa/util"
	"github.com/google/go-github/v74/github"
)

type OrgRepositories struct {
	Organization string   `json:"organization"`
	Repositories []string `json:"repositories"`
}

func GetRepositoryByOrganization(org string) OrgRepositories {
	client := util.GetClient()
	curPage := 1
	var repositories []string
	options := github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100, Page: curPage}}

	lists, _, err := client.Repositories.ListByOrg(context.Background(), org, &options)
	for {
		if err != nil {
			panic(err)
		}
		if len(lists) != 0 {
			for i := range lists {
				repositories = append(repositories, *lists[i].Name)
			}
			curPage++
			options = github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100, Page: curPage}}
			lists, _, err = client.Repositories.ListByOrg(context.Background(), org, &options)
		} else {
			break
		}
	}

	return OrgRepositories{Organization: org, Repositories: repositories}
}

func getDefaultOrg() map[string]bool {

	defaultOrgMap := make(map[string]bool)
	orgStr := beego.AppConfig.String("defaultOrg")

	orgStr = strings.Replace(orgStr, " ", "", -1)
	orgStr = strings.Replace(orgStr, "\n", "", -1)
	organizations := strings.Split(orgStr, ",")
	for i := range organizations {
		defaultOrgMap[organizations[i]] = true
	}
	return defaultOrgMap
}
