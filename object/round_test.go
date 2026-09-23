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
	"testing"
	"time"
)

func getDateFromString(date string) time.Time {
	dateString := date + "T00:00:00+08:00"
	res, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		panic(err)
	}

	return res
}

func getAddedDate(t time.Time, i int) string {
	return t.AddDate(0, 0, i).Format("2006-01-02")
}

func TestAddRounds(t *testing.T) {
	InitConfig()

	startDate := getDateFromString("2026-08-03")

	now := time.Now()
	date := now.Format("2006-01-02")
	date = "2026-08-03"

	j := 200
	for i := 0; i < 100; i++ {
		round := &Round{
			Owner:       "admin_next",
			Name:        fmt.Sprintf("talent2023-week-%02d", j),
			CreatedTime: fmt.Sprintf("%sT00:00:%02d+08:00", date, i+1),
			Title:       fmt.Sprintf("第%d周", j),
			Program:     "talent2023",
			StartDate:   getAddedDate(startDate, 7*i),
			EndDate:     getAddedDate(startDate, 7*(i+1)),
		}

		AddRound(round)
		j += 1
		fmt.Printf("%v\n", round)
	}
}

func TestAddRounds2023(t *testing.T) {
	InitConfig()

	startDate := getDateFromString("2023-01-16")

	now := time.Now()
	date := now.Format("2006-01-02")
	//date = "2022-12-12"

	for i := 80; i < 200; i++ {
		round := &Round{
			Owner:       "admin",
			Name:        fmt.Sprintf("talent2023-week-%02d", i),
			CreatedTime: fmt.Sprintf("%sT00:00:%02d+08:00", date, i),
			Title:       fmt.Sprintf("第%d周", i),
			Program:     "talent2023",
			StartDate:   getAddedDate(startDate, 7*(i-15)),
			EndDate:     getAddedDate(startDate, 7*(i-14)),
		}

		AddRound(round)
		fmt.Printf("%v\n", round)
	}
}

func TestAddRounds2(t *testing.T) {
	InitConfig()

	startDate := getDateFromString("2022-01-13")

	for i := 0; i < 20; i++ {
		newDate := getAddedDate(startDate, i)
		round := &Round{
			Owner:       "admin",
			Name:        fmt.Sprintf("%s-%s", ProgramName, newDate),
			CreatedTime: fmt.Sprintf("%sT00:00:00+08:00", newDate),
			Title:       newDate,
			Program:     ProgramName,
			StartDate:   getAddedDate(startDate, i),
			EndDate:     getAddedDate(startDate, i+1),
		}

		AddRound(round)
		fmt.Printf("%v\n", round)
	}
}
