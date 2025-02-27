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

package com.tencent.devops.process.engine.utils

import com.tencent.devops.common.api.exception.ErrorCodeException
import com.tencent.devops.common.api.exception.OperationException
import com.tencent.devops.common.pipeline.Model
import com.tencent.devops.common.pipeline.container.Stage
import com.tencent.devops.common.pipeline.container.TriggerContainer
import com.tencent.devops.common.pipeline.pojo.BuildFormProperty
import com.tencent.devops.common.pipeline.pojo.BuildNo
import com.tencent.devops.common.service.utils.MessageCodeUtil
import com.tencent.devops.process.constant.ProcessMessageCode
import com.tencent.devops.process.engine.common.VMUtils
import com.tencent.devops.process.engine.compatibility.BuildPropertyCompatibilityTools
import org.slf4j.LoggerFactory
import java.util.regex.Pattern

object PipelineUtils {

    private val logger = LoggerFactory.getLogger(PipelineUtils::class.java)

    private const val ENGLISH_NAME_PATTERN = "[A-Za-z_][A-Za-z_0-9\\.]*"

    fun checkPipelineName(name: String, maxPipelineNameSize: Int) {
        if (name.toCharArray().size > maxPipelineNameSize) {
            throw ErrorCodeException(
                errorCode = ProcessMessageCode.ERROR_PIPELINE_NAME_TOO_LONG,
                defaultMessage = "Pipeline's name is too long"
            )
        }
    }

    fun checkPipelineParams(params: List<BuildFormProperty>) {
        params.forEach {
            if (!Pattern.matches(ENGLISH_NAME_PATTERN, it.id)) {
                logger.warn("Pipeline's start params Name is iregular")
                throw OperationException(
                    message = MessageCodeUtil.getCodeLanMessage(ProcessMessageCode.ERROR_PIPELINE_PARAMS_NAME_ERROR)
                )
            }
        }
    }

    fun checkPipelineDescLength(desc: String?, maxPipelineNameSize: Int) {
        if (desc != null && desc.toCharArray().size > maxPipelineNameSize) {
            throw ErrorCodeException(
                errorCode = ProcessMessageCode.ERROR_PIPELINE_DESC_TOO_LONG,
                defaultMessage = "Pipeline's desc is too long"
            )
        }
    }

    fun getFixedStages(
        model: Model,
        fixedTriggerContainer: TriggerContainer,
        defaultStageTagId: String?
    ): List<Stage> {
        val stages = ArrayList<Stage>()
        val defaultTagIds = if (defaultStageTagId.isNullOrBlank()) emptyList() else listOf(defaultStageTagId)
        model.stages.forEachIndexed { index, stage ->
            stage.id = stage.id ?: VMUtils.genStageId(index + 1)
            if (index == 0) {
                stages.add(stage.copy(containers = listOf(fixedTriggerContainer)))
            } else {
                model.stages.forEach {
                    if (it.name.isNullOrBlank()) it.name = it.id
                    if (it.tag == null) it.tag = defaultTagIds
                }
                stages.add(stage)
            }
        }
        return stages
    }

    /**
     * 通过流水线参数和模板编排生成新Model
     */
    @Suppress("ALL")
    fun instanceModel(
        templateModel: Model,
        pipelineName: String,
        buildNo: BuildNo?,
        param: List<BuildFormProperty>?,
        instanceFromTemplate: Boolean,
        labels: List<String>? = null,
        defaultStageTagId: String?
    ): Model {
        val templateTrigger = templateModel.stages[0].containers[0] as TriggerContainer
        val instanceParam = if (templateTrigger.templateParams == null) {
            BuildPropertyCompatibilityTools.mergeProperties(templateTrigger.params, param ?: emptyList())
        } else {
            BuildPropertyCompatibilityTools.mergeProperties(
                from = templateTrigger.params,
                to = BuildPropertyCompatibilityTools.mergeProperties(
                    from = templateTrigger.templateParams!!, to = param ?: emptyList()
                )
            )
        }

        val triggerContainer = TriggerContainer(
            name = templateTrigger.name,
            elements = templateTrigger.elements,
            params = instanceParam,
            buildNo = buildNo,
            canRetry = templateTrigger.canRetry,
            containerId = templateTrigger.containerId,
            containerHashId = templateTrigger.containerHashId
        )

        return Model(
            name = pipelineName,
            desc = "",
            stages = getFixedStages(templateModel, triggerContainer, defaultStageTagId),
            labels = labels ?: templateModel.labels,
            instanceFromTemplate = instanceFromTemplate
        )
    }
}
