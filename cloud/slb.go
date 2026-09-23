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

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/casbin/casbin-oa/util"
)

type Server struct {
	ServerId string `json:"ServerId"`
	Port     int    `json:"Port"`
	Weight   int    `json:"Weight"`
	Type     string `json:"Type"`
}

func GetVsgServerIdMap() (map[string]int, error) {
	r := slb.CreateDescribeVServerGroupAttributeRequest()
	r.VServerGroupId = vsgId

	resp, err := slbClient.DescribeVServerGroupAttribute(r)
	if err != nil {
		return nil, fmt.Errorf("GetVsgServerIdMap() error: %s", err.Error())
	}

	servers := resp.BackendServers.BackendServer
	res := map[string]int{}
	for _, server := range servers {
		id := server.ServerId
		res[id] = 1
	}
	return res, nil
}

func AddServerToSlb(serverId string, port int) error {
	r := slb.CreateAddVServerGroupBackendServersRequest()
	r.VServerGroupId = vsgId
	r.BackendServers = util.StructToJson([]Server{{
		ServerId: serverId,
		Port:     port,
		Weight:   100,
		Type:     "ecs",
	}})

	var err error
	for i := 0; i < 100; i++ {
		_, err = slbClient.AddVServerGroupBackendServers(r)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("AddServerToSlb() error: server: %s, %s", serverId, err.Error())
}
