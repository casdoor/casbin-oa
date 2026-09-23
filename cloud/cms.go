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

package cloud

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/casbin/casbin-oa/util"
)

type Dimension struct {
	InstanceId string `json:"instanceId"`
}

type Datapoint struct {
	Timestamp  int     `json:"timestamp"`
	UserId     string  `json:"userId"`
	InstanceId string  `json:"instanceId"`
	Vip        string  `json:"vip"`
	Maximum    float64 `json:"Maximum"`
	Minimum    float64 `json:"Minimum"`
	Average    float64 `json:"Average"`
	Sum        float64 `json:"Sum"`
}

func GetSlbPacketRate() int {
	r := cms.CreateDescribeMetricLastRequest()
	r.MetricName = "InstancePacketRX"
	r.Period = "60"
	r.Dimensions = util.StructToJson([]Dimension{{InstanceId: slbId}})
	r.Namespace = "acs_slb_dashboard"

	resp, err := cmsClient.DescribeMetricLast(r)
	if err != nil {
		logError("GetSlbPacketRate() error: %s", err.Error())
		return -1
	}

	var datapoints []Datapoint
	err = util.JsonToStruct(resp.Datapoints, &datapoints)
	if err != nil {
		logError("GetSlbPacketRate() error: %s", err.Error())
		return -1
	}

	if len(datapoints) <= 0 {
		return -1
	}

	res := int(datapoints[0].Average)
	fmt.Printf("[%s] SLB packet rate = %d\n", util.GetCurrentTimeFormatted(), res)

	return res
}
