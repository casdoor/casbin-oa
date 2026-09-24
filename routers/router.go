// Copyright 2020 The casbin Authors. All Rights Reserved.
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

package routers

import (
	"github.com/astaxie/beego"

	"github.com/casbin/casbin-oa/controllers"
)

func init() {
	initAPI()
}

func initAPI() {
	ns :=
		beego.NewNamespace("/api",
			beego.NSInclude(
				&controllers.ApiController{},
			),
		)
	beego.AddNamespace(ns)

	beego.Router("/api/signin", &controllers.ApiController{}, "POST:Signin")
	beego.Router("/api/signout", &controllers.ApiController{}, "POST:Signout")
	beego.Router("/api/get-account", &controllers.ApiController{}, "GET:GetAccount")
	beego.Router("/api/get-users", &controllers.ApiController{}, "GET:GetUsers")

	beego.Router("/api/get-programs", &controllers.ApiController{}, "GET:GetPrograms")
	beego.Router("/api/get-program", &controllers.ApiController{}, "GET:GetProgram")
	beego.Router("/api/update-program", &controllers.ApiController{}, "POST:UpdateProgram")
	beego.Router("/api/add-program", &controllers.ApiController{}, "POST:AddProgram")
	beego.Router("/api/delete-program", &controllers.ApiController{}, "POST:DeleteProgram")

	beego.Router("/api/get-students", &controllers.ApiController{}, "GET:GetStudents")
	beego.Router("/api/get-filtered-students", &controllers.ApiController{}, "GET:GetFilteredStudents")
	beego.Router("/api/get-student", &controllers.ApiController{}, "GET:GetStudent")
	beego.Router("/api/update-student", &controllers.ApiController{}, "POST:UpdateStudent")
	beego.Router("/api/add-student", &controllers.ApiController{}, "POST:AddStudent")
	beego.Router("/api/delete-student", &controllers.ApiController{}, "POST:DeleteStudent")

	beego.Router("/api/get-rounds", &controllers.ApiController{}, "GET:GetRounds")
	beego.Router("/api/get-filtered-rounds", &controllers.ApiController{}, "GET:GetFilteredRounds")
	beego.Router("/api/get-round", &controllers.ApiController{}, "GET:GetRound")
	beego.Router("/api/update-round", &controllers.ApiController{}, "POST:UpdateRound")
	beego.Router("/api/add-round", &controllers.ApiController{}, "POST:AddRound")
	beego.Router("/api/delete-round", &controllers.ApiController{}, "POST:DeleteRound")

	beego.Router("/api/get-reports", &controllers.ApiController{}, "GET:GetReports")
	beego.Router("/api/get-filtered-reports", &controllers.ApiController{}, "GET:GetFilteredReports")
	beego.Router("/api/get-report", &controllers.ApiController{}, "GET:GetReport")
	beego.Router("/api/update-report", &controllers.ApiController{}, "POST:UpdateReport")
	beego.Router("/api/add-report", &controllers.ApiController{}, "POST:AddReport")
	beego.Router("/api/delete-report", &controllers.ApiController{}, "POST:DeleteReport")
	beego.Router("/api/auto-update-report", &controllers.ApiController{}, "POST:AutoUpdateReport")

	beego.Router("/api/get-repositories", &controllers.ApiController{}, "Get:GetRepositoryByOrg")
	beego.Router("/api/get-project-columns", &controllers.ApiController{}, "Get:GetProjectColumns")
	beego.Router("/api/get-github-user", &controllers.ApiController{}, "Get:GetGithubUserByUsername")

	beego.Router("/api/get-issue", &controllers.ApiController{}, "GET:GetIssue")
	beego.Router("/api/get-filtered-issue", &controllers.ApiController{}, "Get:GetIssueByName")
	beego.Router("/api/update-issue", &controllers.ApiController{}, "POST:UpdateIssue")
	beego.Router("/api/add-issue", &controllers.ApiController{}, "POST:AddIssue")
	beego.Router("/api/delete-issue", &controllers.ApiController{}, "POST:DeleteIssue")

	beego.Router("/api/webhook", &controllers.ApiController{}, "Post:WebhookOpen")
	beego.Router("/api/is-mainland-ip", &controllers.ApiController{}, "GET:IsMainlandIp")

	beego.Router("/api/get-pr-checks", &controllers.ApiController{}, "GET:GetPrChecks")
	beego.Router("/api/get-pr-check", &controllers.ApiController{}, "GET:GetPrCheck")

	beego.Router("/api/get-machines", &controllers.ApiController{}, "GET:GetMachines")
	beego.Router("/api/get-machine", &controllers.ApiController{}, "GET:GetMachine")
	beego.Router("/api/update-machine", &controllers.ApiController{}, "POST:UpdateMachine")
	beego.Router("/api/add-machine", &controllers.ApiController{}, "POST:AddMachine")
	beego.Router("/api/delete-machine", &controllers.ApiController{}, "POST:DeleteMachine")
}
