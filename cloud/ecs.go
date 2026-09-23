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
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func GetInstances() ([]ecs.Instance, error) {
	r := ecs.CreateDescribeInstancesRequest()
	r.InstanceChargeType = "PostPaid"
	r.PageSize = requests.NewInteger(100)
	r.PageNumber = requests.NewInteger(1)

	var resp *ecs.DescribeInstancesResponse
	var err error
	for i := 0; i < 100; i++ {
		resp, err = ecsClient.DescribeInstances(r)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("GetInstances() error: %s", err.Error())
	}

	instances := resp.Instances.Instance
	res := []ecs.Instance{}
	for _, instance := range instances {
		if strings.HasPrefix(instance.InstanceName, "auto") {
			res = append(res, instance)
		}
	}

	return res, nil
}

func AddInstance(instanceName string) error {
	r := ecs.CreateRunInstancesRequest()
	r.LaunchTemplateName = "auto"

	resp, err := ecsClient.RunInstances(r)
	if err != nil {
		return fmt.Errorf("AddInstance() error: %s", err.Error())
	}

	if len(resp.InstanceIdSets.InstanceIdSet) == 0 {
		return fmt.Errorf("AddInstance() error: no instance created, name = %s", instanceName)
	}

	instanceId := resp.InstanceIdSets.InstanceIdSet[0]

	err = renameInstance(instanceId, instanceName)
	if err != nil {
		return err
	}

	err = AddServerToSlb(instanceId, 9095)
	if err != nil {
		return err
	}

	fmt.Printf("1 instance added, name = %s\n", instanceName)
	return nil
}

func renameInstanceOnce(r *ecs.ModifyInstanceAttributeRequest) error {
	time.Sleep(3000 * time.Millisecond)

	var err error
	for i := 0; i < 100; i++ {
		_, err = ecsClient.ModifyInstanceAttribute(r)
		if err == nil {
			return nil
		}

		fmt.Printf("renameInstance() error: %s\n", err.Error())
		time.Sleep(2000 * time.Millisecond)
	}

	return fmt.Errorf("renameInstance() error: instance: %s, %s", r.InstanceId, err.Error())
}

func renameInstance(instanceId string, instanceName string) error {
	r := ecs.CreateModifyInstanceAttributeRequest()
	r.InstanceId = instanceId
	r.InstanceName = instanceName

	for i := 0; i < 3; i++ {
		err := renameInstanceOnce(r)
		if err != nil {
			return err
		}
	}

	fmt.Printf("instance: %s renamed to: %s\n", instanceId, instanceName)
	return nil
}

func DeleteInstance(instanceId string, instanceName string) error {
	r := ecs.CreateDeleteInstancesRequest()
	r.InstanceId = &[]string{instanceId}
	r.Force = requests.NewBoolean(true)

	_, err := ecsClient.DeleteInstances(r)
	if err != nil {
		return fmt.Errorf("DeleteInstance() error: name = %s, id = %s, %s", instanceName, instanceId, err.Error())
	}

	fmt.Printf("1 instance deleted, name = %s\n", instanceName)
	return nil
}
