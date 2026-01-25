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
	"fmt"
	"time"
)

type PrCheck struct {
	Id            int       `xorm:"int notnull pk autoincr" json:"id"`
	Org           string    `xorm:"varchar(100)" json:"org"`
	Repo          string    `xorm:"varchar(100)" json:"repo"`
	PrNumber      int       `xorm:"int" json:"prNumber"`
	CheckRunId    int64     `xorm:"bigint" json:"checkRunId"`
	CheckName     string    `xorm:"varchar(200)" json:"checkName"`
	Status        string    `xorm:"varchar(50)" json:"status"`
	Conclusion    string    `xorm:"varchar(50)" json:"conclusion"`
	FailureReason string    `xorm:"text" json:"failureReason"`
	FixAttempts   int       `xorm:"int" json:"fixAttempts"`
	LastAttemptAt time.Time `xorm:"datetime" json:"lastAttemptAt"`
	IsFixed       bool      `xorm:"bool" json:"isFixed"`
	CreatedAt     time.Time `xorm:"datetime" json:"createdAt"`
}

// GetPrCheck retrieves a PR check by org, repo, PR number and check name
func GetPrCheck(org string, repo string, prNumber int, checkName string) *PrCheck {
	prCheck := PrCheck{}
	existed, err := adapter.Engine.Where("org = ? and repo = ? and pr_number = ? and check_name = ?",
		org, repo, prNumber, checkName).Desc("id").Get(&prCheck)
	if err != nil {
		panic(err)
	}
	if existed {
		return &prCheck
	}
	return nil
}

// GetPrChecksByPR retrieves all checks for a specific PR
func GetPrChecksByPR(org string, repo string, prNumber int) []*PrCheck {
	prChecks := []*PrCheck{}
	err := adapter.Engine.Where("org = ? and repo = ? and pr_number = ?",
		org, repo, prNumber).Find(&prChecks)
	if err != nil {
		panic(err)
	}
	return prChecks
}

// AddPrCheck adds a new PR check record
func AddPrCheck(prCheck *PrCheck) bool {
	affected, err := adapter.Engine.Insert(prCheck)
	if err != nil {
		panic(err)
	}
	return affected != 0
}

// UpdatePrCheck updates an existing PR check record
func UpdatePrCheck(id int, prCheck *PrCheck) bool {
	_, err := adapter.Engine.ID(id).AllCols().Update(prCheck)
	if err != nil {
		panic(err)
	}
	return true
}

// IncrementFixAttempts increments the fix attempts counter
func IncrementFixAttempts(org string, repo string, prNumber int, checkName string) bool {
	prCheck := GetPrCheck(org, repo, prNumber, checkName)
	if prCheck == nil {
		return false
	}
	prCheck.FixAttempts++
	prCheck.LastAttemptAt = time.Now()
	return UpdatePrCheck(prCheck.Id, prCheck)
}

// MarkAsFixed marks a check as fixed
func MarkAsFixed(org string, repo string, prNumber int, checkName string) bool {
	prCheck := GetPrCheck(org, repo, prNumber, checkName)
	if prCheck == nil {
		return false
	}
	prCheck.IsFixed = true
	prCheck.Conclusion = "success"
	return UpdatePrCheck(prCheck.Id, prCheck)
}

// ShouldAttemptFix checks if we should attempt to fix this check (max 3 attempts)
func ShouldAttemptFix(org string, repo string, prNumber int, checkName string) bool {
	prCheck := GetPrCheck(org, repo, prNumber, checkName)
	if prCheck == nil {
		return true // First time, should attempt
	}
	return prCheck.FixAttempts < 3 && !prCheck.IsFixed
}

// GetFailureReason returns a formatted failure reason for display
func GetFailureReason(prCheck *PrCheck) string {
	if prCheck == nil || prCheck.FailureReason == "" {
		return "No failure details available"
	}
	return fmt.Sprintf("Check '%s' failed with status: %s\nDetails:\n%s",
		prCheck.CheckName, prCheck.Conclusion, prCheck.FailureReason)
}
