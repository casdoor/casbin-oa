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
	"github.com/casbin/casbin-oa/util"
	"github.com/google/go-github/v74/github"
	"github.com/mileusna/crontab"
)

type OrgWebhook struct {
	Org      string         `json:"org"`
	Webhooks []*github.Hook `json:"webhooks"`
}

func GetAllOrgWebhooks() []*OrgWebhook {
	orgWebhooks := []*OrgWebhook{}
	orgs := GetWebhookOrgs()

	for i := 0; i < len(orgs); i++ {
		webhooks := util.GetOrgWebhooks(orgs[i])
		newOrgWebhook := OrgWebhook{Org: orgs[i], Webhooks: webhooks}
		orgWebhooks = append(orgWebhooks, &newOrgWebhook)
	}

	return orgWebhooks
}

func ReDeliverOrgWebhook(orgWebhook OrgWebhook) {
	orgName := orgWebhook.Org
	for _, webhook := range orgWebhook.Webhooks {
		webhookMap := make(map[string]bool)
		delivers := util.GetWebhookDelivers(orgName, webhook.GetID())

		for _, deliver := range delivers {
			if deliver.GetAction() == "opened" {
				if deliver.GetStatus() == "OK" {
					webhookMap[deliver.GetGUID()] = true
				} else if !deliver.GetRedelivery() {
					_, ok := webhookMap[deliver.GetGUID()]
					if !ok {
						util.RedeliverWebhook(orgName, webhook.GetID(), deliver.GetID())
					}
				}
			}
		}

	}
}

func RedeliverAllOrgWebhook() {
	orgWebhooks := GetAllOrgWebhooks()

	for _, orgWebhook := range orgWebhooks {
		go ReDeliverOrgWebhook(*orgWebhook)
	}
}

func RegularRedeliver() {
	ctab := crontab.New()
	ctab.MustAddJob("*/5 * * * *", RedeliverAllOrgWebhook)
}
