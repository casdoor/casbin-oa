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
	"io"

	"github.com/google/go-github/v38/github"
)

func GetOrgWebhooks(org string) []*github.Hook {
	client := GetClient()
	organizations := client.Organizations
	hooks, response, err := organizations.ListHooks(context.Background(), org, nil)

	if err != nil {
		panic(err)
	}

	if response.StatusCode == 404 {
		return nil
	}

	return hooks
}

func GetWebhookDelivers(org string, webhook int64) []*github.HookDelivery {
	client := GetClient()
	organizations := client.Organizations

	option := github.ListCursorOptions{PerPage: 100}

	var deliveries []*github.HookDelivery
	var response *github.Response
	var err error
	times := 0
	for {
		deliveries, response, err = organizations.ListHookDeliveries(context.Background(), org, webhook, &option)
		if err != nil {
			times += 1
			fmt.Printf("GetWebhookDelivers() error: %s, org = %s, times = %d\n", err.Error(), org, times)
			if times >= 5 {
				panic(err)
			}
		} else {
			break
		}
	}

	if response.StatusCode == 404 {
		return nil
	}

	return deliveries
}

func RedeliverWebhook(org string, webhook int64, delivery int64) {
	url := fmt.Sprintf("https://api.github.com/orgs/%v/hooks/%v/deliveries/%v/attempts", org, webhook, delivery)
	httpClient := GetHttpClient()
	var buf io.ReadWriter
	httpClient.Post(url, "application/json", buf)
}
