/*
 * Tencent is pleased to support the open source community by making BK-CI 蓝鲸持续集成平台 available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company.  All rights reserved.
 *
 * BK-CI 蓝鲸持续集成平台 is licensed under the MIT license.
 *
 * A copy of the MIT License is included in this file.
 *
 *
 * Terms of the MIT License:
 * ---------------------------------------------------
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
 * documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of
 * the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN
 * NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package api

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Tencent/bk-ci/src/agent/src/pkg/config"
	"github.com/Tencent/bk-ci/src/agent/src/pkg/util/httputil"
	"github.com/Tencent/bk-ci/src/agent/src/pkg/util/systemutil"
)

func buildUrl(url string) string {
	if strings.HasPrefix(config.GAgentConfig.Gateway, "http") {
		return config.GAgentConfig.Gateway + url
	} else {
		return "http://" + config.GAgentConfig.Gateway + url
	}
}

func Heartbeat(buildInfos []ThirdPartyBuildInfo, jdkVersion []string) (*httputil.DevopsResult, error) {
	url := buildUrl("/ms/environment/api/buildAgent/agent/thirdPartyAgent/agents/newHeartbeat")

	agentHeartbeatInfo := &AgentHeartbeatInfo{
		MasterVersion:     config.AgentVersion,
		SlaveVersion:      config.GAgentEnv.SlaveVersion,
		HostName:          config.GAgentEnv.HostName,
		AgentIp:           config.GAgentEnv.AgentIp,
		ParallelTaskCount: config.GAgentConfig.ParallelTaskCount,
		AgentInstallPath:  systemutil.GetExecutableDir(),
		StartedUser:       systemutil.GetCurrentUser().Username,
		TaskList:          buildInfos,
		Props: AgentPropsInfo{
			Arch:       runtime.GOARCH,
			JdkVersion: jdkVersion,
		},
	}

	return httputil.NewHttpClient().Post(url).Body(agentHeartbeatInfo).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoDevopsResult()
}

func CheckUpgrade(jdkVersion []string) (*httputil.AgentResult, error) {
	url := buildUrl("/ms/dispatch/api/buildAgent/agent/thirdPartyAgent/upgradeNew")

	info := &UpgradeInfo{
		WorkerVersion:  config.GAgentEnv.SlaveVersion,
		GoAgentVersion: config.AgentVersion,
		JdkVersion:     jdkVersion,
	}

	return httputil.NewHttpClient().Post(url).Body(info).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoAgentResult()
}

func FinishUpgrade(success bool) (*httputil.AgentResult, error) {
	url := buildUrl("/ms/dispatch/api/buildAgent/agent/thirdPartyAgent/upgrade?success=" + strconv.FormatBool(success))
	return httputil.NewHttpClient().Delete(url).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoAgentResult()
}

func DownloadUpgradeFile(serverFile string, saveFile string) (fileMd5 string, err error) {
	url := buildUrl("/ms/environment/api/buildAgent/agent/thirdPartyAgent/upgrade/files/download?file=" + serverFile)
	return httputil.DownloadUpgradeFile(url, config.GAgentConfig.GetAuthHeaderMap(), saveFile)
}

func DownloadAgentInstallScript(saveFile string) error {
	url := buildUrl(fmt.Sprintf("/external/agents/%s/install", config.GAgentConfig.AgentId))
	return httputil.DownloadAgentInstallScript(url, config.GAgentConfig.GetAuthHeaderMap(), saveFile)
}

func AgentStartup() (*httputil.DevopsResult, error) {
	url := buildUrl("/ms/environment/api/buildAgent/agent/thirdPartyAgent/startup")

	startInfo := &ThirdPartyAgentStartInfo{
		HostName:      config.GAgentEnv.HostName,
		HostIp:        config.GAgentEnv.AgentIp,
		DetectOs:      config.GAgentEnv.OsName,
		MasterVersion: config.AgentVersion,
		SlaveVersion:  config.GAgentEnv.SlaveVersion,
	}

	return httputil.NewHttpClient().Post(url).Body(startInfo).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoDevopsResult()
}

func GetAgentStatus() (*httputil.DevopsResult, error) {
	url := buildUrl("/ms/environment/api/buildAgent/agent/thirdPartyAgent/status")
	return httputil.NewHttpClient().Get(url).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoDevopsResult()
}

func GetBuild() (*httputil.AgentResult, error) {
	url := buildUrl("/ms/dispatch/api/buildAgent/agent/thirdPartyAgent/startup")
	return httputil.NewHttpClient().Get(url).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoAgentResult()
}

func WorkerBuildFinish(buildInfo *ThirdPartyBuildWithStatus) (*httputil.DevopsResult, error) {
	url := buildUrl("/ms/dispatch/api/buildAgent/agent/thirdPartyAgent/workerBuildFinish")
	return httputil.NewHttpClient().Post(url).Body(buildInfo).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoDevopsResult()
}

func GetAgentPipeline() (*httputil.DevopsResult, error) {
	url := buildUrl("/ms/environment/api/buildAgent/agent/thirdPartyAgent/agents/pipelines")
	return httputil.NewHttpClient().Get(url).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoDevopsResult()
}

func UpdatePipelineStatus(response *PipelineResponse) (*httputil.DevopsResult, error) {
	url := buildUrl("/ms/environment/api/buildAgent/agent/thirdPartyAgent/agents/pipelines")
	return httputil.NewHttpClient().Put(url).Body(response).SetHeaders(config.GAgentConfig.GetAuthHeaderMap()).Execute().IntoDevopsResult()
}

func DownloadAgentInstallBatchZip(saveFile string) error {
	url := buildUrl(fmt.Sprintf("/ms/environment/api/external/thirdPartyAgent/%s/batch_zip",
		config.GAgentConfig.BatchInstallKey))
	return httputil.DownloadAgentInstallScript(url, config.GAgentConfig.GetAuthHeaderMap(), saveFile)
}
