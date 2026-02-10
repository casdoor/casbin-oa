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
	"strconv"

	"github.com/casbin/casbin-oa/object"
)

// GetPrChecks retrieves all PR check records for a specific PR.
func (c *ApiController) GetPrChecks() {
	org := c.Input().Get("org")
	repo := c.Input().Get("repo")
	prNumberStr := c.Input().Get("prNumber")

	if org == "" || repo == "" || prNumberStr == "" {
		c.Data["json"] = map[string]interface{}{
			"status": "error",
			"msg":    "Missing required parameters",
		}
		c.ServeJSON()
		return
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"status": "error",
			"msg":    "Invalid PR number",
		}
		c.ServeJSON()
		return
	}

	prChecks := object.GetPrChecksByPR(org, repo, prNumber)
	c.Data["json"] = prChecks
	c.ServeJSON()
}

// GetPrCheck retrieves a specific PR check record.
func (c *ApiController) GetPrCheck() {
	org := c.Input().Get("org")
	repo := c.Input().Get("repo")
	prNumberStr := c.Input().Get("prNumber")
	checkName := c.Input().Get("checkName")

	if org == "" || repo == "" || prNumberStr == "" || checkName == "" {
		c.Data["json"] = map[string]interface{}{
			"status": "error",
			"msg":    "Missing required parameters",
		}
		c.ServeJSON()
		return
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"status": "error",
			"msg":    "Invalid PR number",
		}
		c.ServeJSON()
		return
	}

	prCheck := object.GetPrCheck(org, repo, prNumber, checkName)
	c.Data["json"] = prCheck
	c.ServeJSON()
}
