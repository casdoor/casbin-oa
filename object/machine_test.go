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
	"strings"
	"testing"
)

func TestSyncMachine(t *testing.T) {
	InitConfig()

	machine := getMachine("admin", "casbintest")
	machine.syncProcessIds()
	machine.DoActions()
	updateMachine(machine.Owner, machine.Name, machine)
}

func TestDeployMachineService(t *testing.T) {
	InitConfig()

	machine := getMachine("admin", "casbintest")
	machine.syncProcessIds()
	for _, service := range machine.Services {
		if service.Name != "casbin-oa" {
			continue
		}

		var err error

		err = doStop(machine, service)
		if err != nil {
			panic(err)
		}

		err = doPull(machine, service)
		if err != nil {
			if !strings.Contains(err.Error(), "wincredman") && !strings.Contains(err.Error(), "bash: /dev/tty") {
				panic(err)
			}
		}

		err = doBuild(machine, service)
		if err != nil {
			panic(err)
		}

		err = doDeploy(machine, service)
		if err != nil {
			panic(err)
		}

		err = doStart(machine, service)
		if err != nil {
			panic(err)
		}
	}
}

func TestSyncImpermanentMachines(t *testing.T) {
	InitConfig()

	err := syncImpermanentMachines()
	if err != nil {
		t.Fatal(err)
	}
}
