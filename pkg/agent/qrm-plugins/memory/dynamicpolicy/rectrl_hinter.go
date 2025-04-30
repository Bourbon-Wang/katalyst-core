/*
Copyright 2022 The Katalyst Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dynamicpolicy

import (
	"encoding/json"
	"fmt"
	"strconv"

	apiconsts "github.com/kubewharf/katalyst-api/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/resourceplugin/v1alpha1"

	"github.com/kubewharf/katalyst-core/pkg/agent/qrm-plugins/commonstate"
	"github.com/kubewharf/katalyst-core/pkg/agent/qrm-plugins/util"
	"github.com/kubewharf/katalyst-core/pkg/config/agent/qrm"
	"github.com/kubewharf/katalyst-core/pkg/util/general"
)

const (
	templateSharedSubgroup = "shared-%02d"
	sharedGroup            = "shared"
)

type ResctrlHinter interface {
	HintResourceAllocation(podMeta commonstate.AllocationMeta, resourceAllocation *pluginapi.ResourceAllocation)
}

type resctrlHinter struct {
	option               *qrm.ResctrlOptions
	closidEnablingGroups sets.String
}

func identifyCPUSetPool(annoInReq map[string]string) string {
	if pool, ok := annoInReq[apiconsts.PodAnnotationCPUEnhancementCPUSet]; ok {
		return pool
	}

	// fall back to original composite (not flattened) form
	enhancementValue, ok := annoInReq[apiconsts.PodAnnotationCPUEnhancementKey]
	if !ok {
		return ""
	}

	flattenedEnhancements := map[string]string{}
	err := json.Unmarshal([]byte(enhancementValue), &flattenedEnhancements)
	if err != nil {
		return ""
	}
	return identifyCPUSetPool(flattenedEnhancements)
}

func getSharedSubgroup(val int) string {
	// typical mon group is like "shared-xx", except for
	// negative value indicates using "shared" mon group
	if val < 0 {
		return sharedGroup
	}
	return fmt.Sprintf(templateSharedSubgroup, val)
}

func (r *resctrlHinter) getSharedSubgroupByPool(pool string) string {
	if v, ok := r.option.CPUSetPoolToSharedSubgroup[pool]; ok {
		return getSharedSubgroup(v)
	}
	return getSharedSubgroup(r.option.DefaultSharedSubgroup)
}

func ensureToGetMemAllocInfo(resourceAllocation *pluginapi.ResourceAllocation) *pluginapi.ResourceAllocationInfo {
	if _, ok := resourceAllocation.ResourceAllocation[string(v1.ResourceMemory)]; !ok {
		resourceAllocation.ResourceAllocation[string(v1.ResourceMemory)] = &pluginapi.ResourceAllocationInfo{}
	}

	allocInfo := resourceAllocation.ResourceAllocation[string(v1.ResourceMemory)]
	if allocInfo.Annotations == nil {
		allocInfo.Annotations = make(map[string]string)
	}

	return allocInfo
}

func injectRespAnnotationSharedGroup(resourceAllocation *pluginapi.ResourceAllocation, group string) {
	allocInfo := ensureToGetMemAllocInfo(resourceAllocation)
	allocInfo.Annotations[util.AnnotationRdtClosID] = group
}

func injectRespAnnotationPodMonGroup(podMeta commonstate.AllocationMeta, resourceAllocation *pluginapi.ResourceAllocation,
	enablingGroups sets.String, group string,
) {
	// check

	if len(enablingGroups) == 0 || enablingGroups.Has(group) {
		return
	}

	allocInfo := ensureToGetMemAllocInfo(resourceAllocation)
	general.InfofV(6, "mbm: pod %s/%s qos %s not need pod mon_groups",
		podMeta.PodNamespace, podMeta.PodName, group)
	allocInfo.Annotations[util.AnnotationRdtNeedPodMonGroups] = strconv.FormatBool(false)
}

func (r *resctrlHinter) HintResourceAllocation(podMeta commonstate.AllocationMeta, resourceAllocation *pluginapi.ResourceAllocation) {
	if r.option == nil || !r.option.EnableResctrlHint {
		return
	}

	podShortQoS, ok := annoQoSLevelToShortQoSLevel[podMeta.QoSLevel]
	if !ok {
		general.Errorf("pod admit: fail to identify short qos level for %s; skip resctl hint", podMeta.QoSLevel)
		return
	}

	// inject shared subgroup if applicable
	if podMeta.QoSLevel == apiconsts.PodAnnotationQoSLevelSharedCores {
		cpusetPool := identifyCPUSetPool(podMeta.Annotations)
		podShortQoS = r.getSharedSubgroupByPool(cpusetPool)
		injectRespAnnotationSharedGroup(resourceAllocation, podShortQoS)
	}

	// inject pod mon group (false only) if applicable
	injectRespAnnotationPodMonGroup(podMeta, resourceAllocation, r.closidEnablingGroups, podShortQoS)

	return
}

func newResctrlHinter(option *qrm.ResctrlOptions) ResctrlHinter {
	closidEnablingGroups := make(sets.String)
	if option != nil && option.MonGroupsPolicy != nil {
		closidEnablingGroups = sets.NewString(option.MonGroupsPolicy.EnabledClosIDs...)
	}

	return &resctrlHinter{
		option:               option,
		closidEnablingGroups: closidEnablingGroups,
	}
}
