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
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

var ecsClient *ecs.Client
var slbClient *slb.Client
var cmsClient *cms.Client

func init() {
	ecsClient = initEcsClient()
	slbClient = initSlbClient()
	cmsClient = initCmsClient()
}

func initEcsClient() *ecs.Client {
	client, err := ecs.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		logError("initEcsClient() error: %s", err.Error())
		return nil
	}

	return client
}

func initSlbClient() *slb.Client {
	client, err := slb.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		logError("initSlbClient() error: %s", err.Error())
		return nil
	}

	return client
}

func initCmsClient() *cms.Client {
	client, err := cms.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		logError("initCmsClient() error: %s", err.Error())
		return nil
	}

	return client
}
